package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	baseURL = "http://localhost:8080/api/v1"
	timeout = 30 * time.Second
)

type TestConfig struct {
	AuthToken  string
	UserID     string
	OrgID      string
	MarinaID   string
	CustomerID string
	BaseURL    string
}

type NotificationResponse struct {
	ID             uuid.UUID              `json:"id"`
	UserID         uuid.UUID              `json:"userId"`
	OrganizationID uuid.UUID              `json:"organizationId"`
	MarinaID       uuid.UUID              `json:"marinaId"`
	Type           string                 `json:"type"`
	Title          string                 `json:"title"`
	Content        string                 `json:"content"`
	Data           map[string]interface{} `json:"data,omitempty"`
	Read           bool                   `json:"read"`
	Priority       string                 `json:"priority"`
	CreatedAt      time.Time              `json:"createdAt"`
	UpdatedAt      time.Time              `json:"updatedAt"`
	ReadAt         *time.Time             `json:"readAt,omitempty"`
}

type UnreadCountResponse struct {
	Count int64 `json:"count"`
}

type SSEEvent struct {
	Event string                 `json:"event"`
	Data  map[string]interface{} `json:"data"`
}

func main() {
	log.Println("🚀 Starting comprehensive notification system test...")

	// Get configuration from environment or use defaults
	config := &TestConfig{
		AuthToken:  getEnv("TEST_AUTH_TOKEN", ""),
		UserID:     getEnv("TEST_USER_ID", ""),
		OrgID:      getEnv("TEST_ORG_ID", ""),
		MarinaID:   getEnv("TEST_MARINA_ID", ""),
		CustomerID: getEnv("TEST_CUSTOMER_ID", ""),
		BaseURL:    getEnv("TEST_BASE_URL", baseURL),
	}

	if config.AuthToken == "" {
		log.Println("⚠️  No AUTH_TOKEN provided. Please set TEST_AUTH_TOKEN environment variable.")
		log.Println("   You can get a token by logging in via the /auth/login endpoint")
		return
	}

	// Run all tests
	runNotificationTests(config)
}

func runNotificationTests(config *TestConfig) {
	log.Println("\n📋 Running notification system tests...")

	// Test 1: Basic notification endpoints
	log.Println("\n1️⃣  Testing basic notification endpoints...")
	testBasicNotificationEndpoints(config)

	// Test 2: Message notification flow
	log.Println("\n2️⃣  Testing message notification flow...")
	testMessageNotificationFlow(config)

	// Test 3: Invite notification flow
	log.Println("\n3️⃣  Testing invite notification flow...")
	testInviteNotificationFlow(config)

	// Test 4: SSE real-time streaming
	log.Println("\n4️⃣  Testing SSE real-time streaming...")
	testSSEStream(config)

	// Test 5: Mark as read functionality
	log.Println("\n5️⃣  Testing mark as read functionality...")
	testMarkAsReadFlow(config)

	log.Println("\n🎉 All notification tests completed!")
}

func testBasicNotificationEndpoints(config *TestConfig) {
	// Test unread count
	log.Println("   📊 Testing unread count endpoint...")
	unreadCount := getUnreadCount(config)
	log.Printf("   ✅ Current unread count: %d", unreadCount)

	// Test list notifications
	log.Println("   📋 Testing list notifications endpoint...")
	notifications := listNotifications(config, false)
	log.Printf("   ✅ Retrieved %d notifications", len(notifications))

	// Test list unread notifications
	log.Println("   📋 Testing list unread notifications endpoint...")
	unreadNotifications := listNotifications(config, true)
	log.Printf("   ✅ Retrieved %d unread notifications", len(unreadNotifications))

	// Test notifications by type
	log.Println("   🔍 Testing notifications by type endpoint...")
	messageNotifications := getNotificationsByType(config, "message")
	log.Printf("   ✅ Retrieved %d message notifications", len(messageNotifications))
}

func testMessageNotificationFlow(config *TestConfig) {
	if config.CustomerID == "" {
		log.Println("   ⚠️  Skipping message test - no customer ID provided")
		return
	}

	log.Println("   📨 Creating test message...")

	// Create a customer-to-marina message (should trigger notifications for marina staff)
	messageData := map[string]interface{}{
		"marinaId":   config.MarinaID,
		"customerId": config.CustomerID,
		"body":       "Test message from customer - this should create notifications for marina staff",
		"sender":     "Test Customer",
		"recipient":  "Marina Staff",
		"contact":    "test@example.com",
		"pinned":     false,
	}

	response := makeRequest(config, "POST", "/message/customer", messageData)
	if response != nil {
		log.Println("   ✅ Message created successfully")

		// Wait a moment for notifications to be processed
		time.Sleep(2 * time.Second)

		// Check if notifications were created
		newUnreadCount := getUnreadCount(config)
		log.Printf("   📊 Unread count after message: %d", newUnreadCount)
	}
}

func testInviteNotificationFlow(config *TestConfig) {
	log.Println("   👥 Creating test user invitation...")

	// Create a user invitation (should trigger notifications for marina staff)
	inviteData := map[string]interface{}{
		"firstName":      "Test",
		"lastName":       "User",
		"email":          fmt.Sprintf("test-user-%d@example.com", time.Now().Unix()),
		"phone":          "+1234567890",
		"title":          "Test User",
		"username":       fmt.Sprintf("testuser%d", time.Now().Unix()),
		"organizationId": config.OrgID,
		"marinaId":       config.MarinaID,
		"roleId":         "550e8400-e29b-41d4-a716-446655440000", // You may need to adjust this
		"isActive":       true,
		"isSuperuser":    false,
	}

	response := makeRequest(config, "POST", "/user/invite", inviteData)
	if response != nil {
		log.Println("   ✅ User invitation created successfully")

		// Wait a moment for notifications to be processed
		time.Sleep(2 * time.Second)

		// Check if notifications were created
		newUnreadCount := getUnreadCount(config)
		log.Printf("   📊 Unread count after invite: %d", newUnreadCount)
	}
}

func testSSEStream(config *TestConfig) {
	log.Println("   🌊 Testing SSE stream...")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	var eventsReceived []SSEEvent
	var mu sync.Mutex

	// Start SSE connection
	wg.Add(1)
	go func() {
		defer wg.Done()

		req, err := http.NewRequestWithContext(ctx, "GET", config.BaseURL+"/notifications/stream", nil)
		if err != nil {
			log.Printf("   ❌ Failed to create SSE request: %v", err)
			return
		}

		req.Header.Set("Authorization", "Bearer "+config.AuthToken)
		req.Header.Set("Accept", "text/event-stream")
		req.Header.Set("Cache-Control", "no-cache")

		client := &http.Client{Timeout: 20 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("   ❌ Failed to connect to SSE: %v", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Printf("   ❌ SSE connection failed with status: %d", resp.StatusCode)
			return
		}

		log.Println("   ✅ SSE connection established")

		// Read SSE events
		buffer := make([]byte, 4096)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				n, err := resp.Body.Read(buffer)
				if err != nil {
					if err != io.EOF {
						log.Printf("   ⚠️  SSE read error: %v", err)
					}
					return
				}

				data := string(buffer[:n])
				events := parseSSEData(data)

				mu.Lock()
				eventsReceived = append(eventsReceived, events...)
				for _, event := range events {
					log.Printf("   📡 SSE Event received: %s", event.Event)
				}
				mu.Unlock()
			}
		}
	}()

	// Wait a moment for connection to establish
	time.Sleep(2 * time.Second)

	// Trigger some notifications while SSE is connected
	log.Println("   🔄 Triggering notifications while SSE is active...")

	// Create a test message to trigger notifications
	if config.CustomerID != "" {
		messageData := map[string]interface{}{
			"marinaId":   config.MarinaID,
			"customerId": config.CustomerID,
			"body":       "SSE test message - should appear in real-time stream",
			"sender":     "SSE Test",
			"recipient":  "Marina Staff",
			"contact":    "sse-test@example.com",
			"pinned":     false,
		}
		makeRequest(config, "POST", "/message/customer", messageData)
	}

	// Wait for events to be processed
	time.Sleep(5 * time.Second)
	cancel() // Close the SSE connection
	wg.Wait()

	mu.Lock()
	eventCount := len(eventsReceived)
	mu.Unlock()

	log.Printf("   ✅ SSE test completed. Received %d events", eventCount)
}

func testMarkAsReadFlow(config *TestConfig) {
	log.Println("   ✅ Testing mark as read functionality...")

	// Get current notifications
	notifications := listNotifications(config, true) // Get unread only
	if len(notifications) == 0 {
		log.Println("   ⚠️  No unread notifications to test mark as read")
		return
	}

	// Mark first notification as read
	firstNotification := notifications[0]
	url := fmt.Sprintf("/notifications/%s/read", firstNotification.ID.String())

	response := makeRequest(config, "PUT", url, nil)
	if response != nil {
		log.Println("   ✅ Notification marked as read successfully")

		// Check unread count decreased
		newUnreadCount := getUnreadCount(config)
		log.Printf("   📊 Unread count after marking as read: %d", newUnreadCount)
	}

	// Test mark all as read
	log.Println("   ✅ Testing mark all as read...")
	response = makeRequest(config, "PUT", "/notifications/mark-all-read", nil)
	if response != nil {
		log.Println("   ✅ All notifications marked as read successfully")

		finalUnreadCount := getUnreadCount(config)
		log.Printf("   📊 Final unread count: %d", finalUnreadCount)
	}
}

// Helper functions

func getUnreadCount(config *TestConfig) int64 {
	response := makeRequest(config, "GET", "/notifications/unread-count", nil)
	if response == nil {
		return 0
	}

	var result struct {
		Data UnreadCountResponse `json:"data"`
	}

	if err := json.Unmarshal(response, &result); err != nil {
		log.Printf("Failed to parse unread count response: %v", err)
		return 0
	}

	return result.Data.Count
}

func listNotifications(config *TestConfig, unreadOnly bool) []NotificationResponse {
	url := "/notifications/list?pageSize=50"
	if unreadOnly {
		url += "&unreadOnly=true"
	}

	response := makeRequest(config, "GET", url, nil)
	if response == nil {
		return nil
	}

	var result struct {
		Data []NotificationResponse `json:"data"`
	}

	if err := json.Unmarshal(response, &result); err != nil {
		log.Printf("Failed to parse notifications response: %v", err)
		return nil
	}

	return result.Data
}

func getNotificationsByType(config *TestConfig, notificationType string) []NotificationResponse {
	url := fmt.Sprintf("/notifications/type/%s?pageSize=50", notificationType)

	response := makeRequest(config, "GET", url, nil)
	if response == nil {
		return nil
	}

	var result struct {
		Data []NotificationResponse `json:"data"`
	}

	if err := json.Unmarshal(response, &result); err != nil {
		log.Printf("Failed to parse notifications by type response: %v", err)
		return nil
	}

	return result.Data
}

func makeRequest(config *TestConfig, method, endpoint string, data interface{}) []byte {
	var body io.Reader

	if data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			log.Printf("Failed to marshal request data: %v", err)
			return nil
		}
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, config.BaseURL+endpoint, body)
	if err != nil {
		log.Printf("Failed to create request: %v", err)
		return nil
	}

	req.Header.Set("Authorization", "Bearer "+config.AuthToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Request failed: %v", err)
		return nil
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read response: %v", err)
		return nil
	}

	if resp.StatusCode >= 400 {
		log.Printf("Request failed with status %d: %s", resp.StatusCode, string(responseBody))
		return nil
	}

	return responseBody
}

func parseSSEData(data string) []SSEEvent {
	var events []SSEEvent
	lines := strings.Split(data, "\n")

	var currentEvent string
	var currentData string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			// End of event
			if currentEvent != "" && currentData != "" {
				var eventData map[string]interface{}
				if err := json.Unmarshal([]byte(currentData), &eventData); err == nil {
					events = append(events, SSEEvent{
						Event: currentEvent,
						Data:  eventData,
					})
				}
			}
			currentEvent = ""
			currentData = ""
		} else if strings.HasPrefix(line, "event: ") {
			currentEvent = strings.TrimPrefix(line, "event: ")
		} else if strings.HasPrefix(line, "data: ") {
			currentData = strings.TrimPrefix(line, "data: ")
		}
	}

	return events
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
