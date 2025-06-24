# Redis Dual Database Implementation

This directory contains a comprehensive test for the enhanced Redis client with dual database support.

## Overview

The marina management system now uses Redis/Valkey with **two separate databases** to optimize performance and organize data logically:

- **Cache Database (DB 0)**: Fast, temporary data
- **Task Database (DB 1)**: Persistent tasks and queues

## Architecture

### Cache Database (DB 0)
**Purpose**: High-speed temporary data and real-time operations
- User sessions (`session:user:{id}`)
- Rate limiting counters (`rate_limit:user:{id}:{minute}`)
- Real-time pub/sub messaging
- Temporary authentication tokens
- Online user status

### Task Database (DB 1)  
**Purpose**: Persistent notifications and background processing
- Notification storage (`notification:{id}`)
- Delivery queues (`queue:notifications_pending`, `queue:notifications_urgent`)
- Delivery status tracking (`delivery:{notification_id}`)
- Task retry mechanisms
- Historical notification data

## Key Features

### ✅ Dual Database Support
- Separate connections to DB 0 (cache) and DB 1 (tasks)
- Automatic database selection based on operation type
- Backward compatibility with existing single-database methods

### ✅ Queue Operations
```go
// Enqueue a notification for processing
client.EnqueueTask(ctx, "notifications_pending", notification)

// Process notifications from queue
message, err := client.DequeueTask(ctx, "notifications_pending")

// Check queue length
length, err := client.QueueLength(ctx, "notifications_pending")
```

### ✅ Cache Operations
```go
// Store user session in cache
client.SetCache(ctx, "session:user:123", sessionData, 2*time.Hour)

// Get session data
var session map[string]interface{}
client.GetCacheJSON(ctx, "session:user:123", &session)

// Rate limiting
client.SetCache(ctx, "rate_limit:user:123", 1, 60*time.Second)
```

### ✅ Task Operations
```go
// Store persistent notification
client.SetTask(ctx, "notification:msg_123", notification, 30*24*time.Hour)

// Track delivery status
client.SetTask(ctx, "delivery:msg_123", deliveryStatus, 7*24*time.Hour)
```

### ✅ Pub/Sub for Real-time
```go
// Publish real-time notification
client.Publish(ctx, "notifications:user:123", notification)

// Subscribe to notifications
pubsub := client.Subscribe(ctx, "notifications:user:123")
```

## Configuration

### Environment Variables
```bash
# Local Development
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_USERNAME=valkey-local
REDIS_PASSWORD=password
REDIS_DB=0
REDIS_TASK_DB=1
REDIS_KEY_PREFIX=dm:

# AWS Production
REDIS_HOST=dmweb-ai-myyysj.serverless.use1.cache.amazonaws.com
REDIS_USERNAME=valkey-dmweb-dev
REDIS_PASSWORD=V@lKeDMw3B-AIDev
```

### Docker Setup
```yaml
# docker-compose.dev.yml
valkey:
  image: valkey/valkey:8.0-alpine
  ports:
    - "6379:6379"
  environment:
    - VALKEY_PASSWORD=password
  command: >
    valkey-server
    --requirepass password
    --user valkey-local on >password allcommands allkeys
```

## Testing

Run the comprehensive test suite:
```bash
go run scripts/redis_dual_db_test/main.go
```

### Test Coverage
1. **Connection Test**: Verifies both databases are accessible
2. **Cache Operations**: Session management, rate limiting, TTL
3. **Task Operations**: Notification storage, delivery tracking
4. **Queue Operations**: Enqueue, dequeue, length checking
5. **Pub/Sub Operations**: Real-time message publishing/subscribing
6. **Complete Workflow**: End-to-end notification delivery simulation

## API Endpoints

### Test Dual Database Connection
```
GET /api/v1/redis/dual-databases/ping
```

### Original Test Endpoints
```
GET /api/v1/redis/ping
GET /api/v1/redis/test
POST /api/v1/redis/test-operations
```

## Benefits

### 🚀 Performance
- **Cache DB**: Optimized for speed with short TTLs
- **Task DB**: Optimized for persistence and reliability
- Separate connections prevent blocking operations

### 🔧 Scalability  
- Independent scaling of cache vs task operations
- Queue-based processing for notifications
- Retry mechanisms for failed deliveries

### 🛡️ Reliability
- Persistent storage for critical notification data
- Separate failure domains for cache vs tasks
- Comprehensive error handling and logging

## Notification Workflow

1. **User Login** → Session cached in DB 0
2. **New Message** → Notification stored in DB 1
3. **Queue Processing** → Added to delivery queue in DB 1
4. **Online Check** → User status from cache in DB 0
5. **Real-time Delivery** → Pub/sub message via DB 0
6. **Status Tracking** → Delivery status stored in DB 1
7. **Rate Limiting** → Spam prevention via cache in DB 0

## Monitoring

The implementation includes comprehensive logging for:
- Database connection status
- Operation latencies
- Queue lengths and processing
- Pub/sub message delivery
- Error rates and retry attempts

All operations are logged with structured logging using Zap for easy monitoring and debugging. 