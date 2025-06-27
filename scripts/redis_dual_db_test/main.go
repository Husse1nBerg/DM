package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/dockworks/dm-web-backend/internal/config"
	"github.com/dockworks/dm-web-backend/pkg/logger"
	"github.com/dockworks/dm-web-backend/pkg/redis"
)

func main() {
	// Load configuration
	cfg := config.New()

	// Initialize logger properly
	zapLogger := logger.NewLogger(cfg.Logger)

	// Create Redis client with dual databases
	redisClient := redis.NewClient(cfg.Redis, zapLogger)

	ctx := context.Background()

	fmt.Println("🚀 Testing Redis Dual Database Setup")
	fmt.Println("=====================================")

	// Test 1: Ping both databases
	fmt.Println("\n🔹 Testing connection to both databases...")
	err := redisClient.Ping(ctx)
	if err != nil {
		log.Fatalf("❌ Redis ping failed: %v", err)
	}
	fmt.Println("✅ Both cache (DB 0) and task (DB 1) databases are connected!")

	// Test 2: Cache operations (DB 0)
	fmt.Println("\n🔹 Testing CACHE database operations (DB 0)...")
	testCacheOperations(ctx, redisClient)

	// Test 3: Task operations (DB 1)
	fmt.Println("\n🔹 Testing TASK database operations (DB 1)...")
	testTaskOperations(ctx, redisClient)

	// Test 4: Queue operations (DB 1)
	fmt.Println("\n🔹 Testing QUEUE operations (DB 1)...")
	testQueueOperations(ctx, redisClient)

	// Test 5: Pub/Sub operations (DB 0)
	fmt.Println("\n🔹 Testing PUB/SUB operations (DB 0)...")
	testPubSubOperations(ctx, redisClient)

	// Test 6: Real-world notification scenario
	fmt.Println("\n🔹 Testing NOTIFICATION system scenario...")
	testNotificationScenario(ctx, redisClient)

	fmt.Println("\n🎉 All dual database tests passed!")
	fmt.Printf("📊 Redis Config: Host=%s, KeyPrefix=%s\n",
		cfg.Redis.Host, cfg.Redis.KeyPrefix)
}

func testCacheOperations(ctx context.Context, client *redis.Client) {
	// Test caching user session
	userSession := map[string]interface{}{
		"user_id":    123,
		"marina_id":  456,
		"login_time": time.Now().Unix(),
		"role":       "marina_admin",
	}

	err := client.SetCache(ctx, "session:user:123", userSession, 1*time.Hour)
	if err != nil {
		log.Fatalf("❌ Cache SET failed: %v", err)
	}
	fmt.Println("✅ Cached user session in DB 0")

	// Retrieve from cache
	var retrievedSession map[string]interface{}
	err = client.GetCacheJSON(ctx, "session:user:123", &retrievedSession)
	if err != nil {
		log.Fatalf("❌ Cache GET failed: %v", err)
	}
	fmt.Printf("✅ Retrieved session from cache: user_id=%v, role=%v\n",
		retrievedSession["user_id"], retrievedSession["role"])

	// Test rate limiting cache
	err = client.SetCache(ctx, "rate_limit:user:123", 1, 60*time.Second)
	if err != nil {
		log.Fatalf("❌ Rate limit cache failed: %v", err)
	}
	fmt.Println("✅ Set rate limiting data in cache")

	// Check TTL
	ttl, err := client.TTLCache(ctx, "rate_limit:user:123")
	if err != nil {
		log.Fatalf("❌ TTL check failed: %v", err)
	}
	fmt.Printf("✅ Rate limit TTL: %s\n", ttl)
}

func testTaskOperations(ctx context.Context, client *redis.Client) {
	// Test storing notification task
	notificationTask := map[string]interface{}{
		"notification_id": "notif_789",
		"user_id":         123,
		"type":            "message",
		"title":           "New Message",
		"content":         "You have a new message from the marina",
		"created_at":      time.Now().Unix(),
		"retry_count":     0,
		"max_retries":     3,
	}

	err := client.SetTask(ctx, "notification:pending:notif_789", notificationTask, 24*time.Hour)
	if err != nil {
		log.Fatalf("❌ Task SET failed: %v", err)
	}
	fmt.Println("✅ Stored notification task in DB 1")

	// Retrieve task
	var retrievedTask map[string]interface{}
	err = client.GetTaskJSON(ctx, "notification:pending:notif_789", &retrievedTask)
	if err != nil {
		log.Fatalf("❌ Task GET failed: %v", err)
	}
	fmt.Printf("✅ Retrieved task from DB 1: type=%v, title=%v\n",
		retrievedTask["type"], retrievedTask["title"])

	// Test delivery tracking
	deliveryStatus := map[string]interface{}{
		"notification_id": "notif_789",
		"status":          "delivered",
		"delivered_at":    time.Now().Unix(),
		"attempts":        1,
	}

	err = client.SetTask(ctx, "notification:status:notif_789", deliveryStatus, 7*24*time.Hour)
	if err != nil {
		log.Fatalf("❌ Delivery status failed: %v", err)
	}
	fmt.Println("✅ Stored delivery status in DB 1")
}

func testQueueOperations(ctx context.Context, client *redis.Client) {
	// Test notification delivery queue
	notifications := []map[string]interface{}{
		{
			"id":       "notif_001",
			"user_id":  123,
			"type":     "message",
			"title":    "Welcome Message",
			"priority": "normal",
		},
		{
			"id":       "notif_002",
			"user_id":  124,
			"type":     "alert",
			"title":    "System Maintenance",
			"priority": "high",
		},
		{
			"id":       "notif_003",
			"user_id":  123,
			"type":     "invite",
			"title":    "Marina Invitation",
			"priority": "normal",
		},
	}

	// Enqueue notifications
	for _, notif := range notifications {
		err := client.EnqueueTask(ctx, "notifications_pending", notif)
		if err != nil {
			log.Fatalf("❌ Enqueue failed: %v", err)
		}
		fmt.Printf("✅ Enqueued notification: %s (priority: %s)\n",
			notif["title"], notif["priority"])
	}

	// Check queue length
	length, err := client.QueueLength(ctx, "notifications_pending")
	if err != nil {
		log.Fatalf("❌ Queue length check failed: %v", err)
	}
	fmt.Printf("✅ Queue length: %d notifications pending\n", length)

	// Process queue (dequeue notifications)
	fmt.Println("\n📤 Processing notification queue...")
	for i := 0; i < int(length); i++ {
		task, err := client.DequeueTask(ctx, "notifications_pending")
		if err != nil {
			if err.Error() == fmt.Sprintf("queue %dqueue:notifications_pending is empty", client.Cache().Options().DB) {
				break
			}
			log.Fatalf("❌ Dequeue failed: %v", err)
		}
		fmt.Printf("✅ Processed notification: %s\n", task[:50]+"...")
	}

	// Test priority queue for high-priority notifications
	highPriorityNotif := map[string]interface{}{
		"id":       "notif_urgent",
		"user_id":  125,
		"type":     "emergency",
		"title":    "Emergency Alert",
		"priority": "urgent",
	}

	err = client.EnqueueTask(ctx, "notifications_urgent", highPriorityNotif)
	if err != nil {
		log.Fatalf("❌ Urgent queue failed: %v", err)
	}
	fmt.Println("✅ Added to urgent notification queue")
}

func testPubSubOperations(ctx context.Context, client *redis.Client) {
	// Test real-time notification delivery via pub/sub
	var wg sync.WaitGroup
	var mu sync.Mutex
	receivedNotifications := make([]string, 0)

	// Start subscriber for user notifications
	wg.Add(1)
	go func() {
		defer wg.Done()

		pubsub := client.Subscribe(ctx, "notifications:user:123")
		defer pubsub.Close()

		fmt.Println("📡 Subscribed to real-time notifications for user 123")

		subCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()

		for {
			select {
			case <-subCtx.Done():
				return
			default:
				msg, err := pubsub.ReceiveMessage(subCtx)
				if err != nil {
					return
				}

				fmt.Printf("📨 Real-time notification received: %s\n", msg.Payload[:50]+"...")

				mu.Lock()
				receivedNotifications = append(receivedNotifications, msg.Payload)
				mu.Unlock()
			}
		}
	}()

	// Give subscriber time to start
	time.Sleep(100 * time.Millisecond)

	// Publish real-time notifications
	realtimeNotifications := []map[string]interface{}{
		{
			"type":      "message",
			"title":     "New Message",
			"content":   "You have a new message",
			"timestamp": time.Now().Unix(),
		},
		{
			"type":      "system",
			"title":     "System Update",
			"content":   "System will be updated tonight",
			"timestamp": time.Now().Unix(),
		},
	}

	for _, notif := range realtimeNotifications {
		err := client.Publish(ctx, "notifications:user:123", notif)
		if err != nil {
			log.Fatalf("❌ Publish failed: %v", err)
		}
		fmt.Printf("✅ Published real-time notification: %s\n", notif["title"])
		time.Sleep(10 * time.Millisecond)
	}

	wg.Wait()
	fmt.Printf("✅ Received %d real-time notifications\n", len(receivedNotifications))
}

func testNotificationScenario(ctx context.Context, client *redis.Client) {
	fmt.Println("\n🎯 Simulating complete notification workflow...")

	// 1. User logs in - cache session
	userSession := map[string]interface{}{
		"user_id":    456,
		"marina_id":  789,
		"online":     true,
		"login_time": time.Now().Unix(),
	}

	err := client.SetCache(ctx, "session:user:456", userSession, 2*time.Hour)
	if err != nil {
		log.Fatalf("❌ Session cache failed: %v", err)
	}
	fmt.Println("1. ✅ User 456 logged in - session cached")

	// 2. New message arrives - store in task DB
	messageNotification := map[string]interface{}{
		"id":         "msg_12345",
		"user_id":    456,
		"type":       "message",
		"title":      "New Message from Marina",
		"content":    "Your slip rental payment is due tomorrow",
		"sender":     "Marina Bay Office",
		"created_at": time.Now().Unix(),
		"read":       false,
	}

	err = client.SetTask(ctx, "notification:msg_12345", messageNotification, 30*24*time.Hour)
	if err != nil {
		log.Fatalf("❌ Notification storage failed: %v", err)
	}
	fmt.Println("2. ✅ Message notification stored in task DB")

	// 3. Add to delivery queue
	err = client.EnqueueTask(ctx, "notifications_pending", messageNotification)
	if err != nil {
		log.Fatalf("❌ Queue add failed: %v", err)
	}
	fmt.Println("3. ✅ Added to notification delivery queue")

	// 4. Check if user is online (from cache)
	var session map[string]interface{}
	err = client.GetCacheJSON(ctx, "session:user:456", &session)
	if err != nil {
		fmt.Println("4. ⚠️  User offline - will deliver later")
	} else {
		if online, ok := session["online"].(bool); ok && online {
			fmt.Println("4. ✅ User is online - can deliver real-time notification")

			// 5. Send real-time notification via pub/sub
			err = client.Publish(ctx, "notifications:user:456", messageNotification)
			if err != nil {
				log.Fatalf("❌ Real-time delivery failed: %v", err)
			}
			fmt.Println("5. ✅ Real-time notification sent via pub/sub")
		}
	}

	// 6. Mark as delivered and update delivery status
	deliveryStatus := map[string]interface{}{
		"notification_id": "msg_12345",
		"user_id":         456,
		"status":          "delivered",
		"delivered_at":    time.Now().Unix(),
		"delivery_method": "realtime",
	}

	err = client.SetTask(ctx, "delivery:msg_12345", deliveryStatus, 7*24*time.Hour)
	if err != nil {
		log.Fatalf("❌ Delivery status failed: %v", err)
	}
	fmt.Println("6. ✅ Delivery status tracked in task DB")

	// 7. Update rate limiting (prevent spam)
	currentTime := time.Now().Unix()
	err = client.SetCache(ctx, fmt.Sprintf("rate_limit:user:456:%d", currentTime/60), 1, 60*time.Second)
	if err != nil {
		log.Fatalf("❌ Rate limiting failed: %v", err)
	}
	fmt.Println("7. ✅ Rate limiting updated in cache DB")

	fmt.Println("\n🏆 Complete notification workflow successful!")
	fmt.Println("   Cache DB (0): Session data, rate limiting, real-time connections")
	fmt.Println("   Task DB (1): Persistent notifications, delivery status, queues")
}
