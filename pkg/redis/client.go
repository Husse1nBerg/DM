package redis

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dockworks/dm-web-backend/internal/config"
	"github.com/dockworks/dm-web-backend/pkg/logger"
	"github.com/redis/go-redis/v9"
)

// Client wraps the Redis client with our specific functionality
type Client struct {
	cache  *redis.Client // DB 0 - for caching, sessions, rate limiting
	tasks  *redis.Client // DB 1 - for task queues, notifications
	config config.RedisConfig
	logger *logger.Logger
}

// NewClient creates a new Redis client instance with separate cache and task databases
func NewClient(cfg config.RedisConfig, logger *logger.Logger) *Client {
	// Base Redis options
	baseOptions := &redis.Options{
		Addr:     cfg.Addr(),
		Username: cfg.Username,
		Password: cfg.Password,
		DB:       0, // Always use DB 0 for cluster compatibility
	}

	// Configure TLS if enabled
	if cfg.TLSEnabled {
		baseOptions.TLSConfig = &tls.Config{
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: false,
		}
	}

	// Cache database (DB 0) - for fast, temporary data
	cacheOptions := *baseOptions
	cacheClient := redis.NewClient(&cacheOptions)

	// Task database - use same connection as cache for cluster compatibility
	// We'll differentiate using key prefixes instead of separate databases
	taskOptions := *baseOptions
	taskClient := redis.NewClient(&taskOptions)

	logger.Zap.Infow("Redis client initialized",
		"host", cfg.Host,
		"port", cfg.Port,
		"tls_enabled", cfg.TLSEnabled,
		"cluster_mode", "compatible")

	return &Client{
		cache:  cacheClient,
		tasks:  taskClient,
		config: cfg,
		logger: logger,
	}
}

// Cache returns the cache database client (DB 0)
func (c *Client) Cache() *redis.Client {
	return c.cache
}

// Tasks returns the task database client (DB 1)
func (c *Client) Tasks() *redis.Client {
	return c.tasks
}

// Ping tests the connection to both Redis databases
func (c *Client) Ping(ctx context.Context) error {
	// Test cache database
	_, err := c.cache.Ping(ctx).Result()
	if err != nil {
		c.logger.Zap.Error("Redis cache database ping failed", err)
		return fmt.Errorf("redis cache ping failed: %w", err)
	}

	// Test task database
	_, err = c.tasks.Ping(ctx).Result()
	if err != nil {
		c.logger.Zap.Error("Redis task database ping failed", err)
		return fmt.Errorf("redis task ping failed: %w", err)
	}

	return nil
}

// === CACHE OPERATIONS (DB 0) ===

// SetCache stores a key-value pair in the cache database with optional expiration
func (c *Client) SetCache(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.setToDatabase(ctx, c.cache, "cache", key, value, expiration)
}

// GetCache retrieves a value by key from the cache database
func (c *Client) GetCache(ctx context.Context, key string) (string, error) {
	return c.getFromDatabase(ctx, c.cache, "cache", key)
}

// GetCacheJSON retrieves a value and unmarshals it from JSON from cache database
func (c *Client) GetCacheJSON(ctx context.Context, key string, dest interface{}) error {
	val, err := c.GetCache(ctx, key)
	if err != nil {
		return err
	}

	err = json.Unmarshal([]byte(val), dest)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON from cache: %w", err)
	}

	return nil
}

// DeleteCache removes keys from the cache database
func (c *Client) DeleteCache(ctx context.Context, keys ...string) error {
	return c.deleteFromDatabase(ctx, c.cache, "cache", keys...)
}

// ExistsCache checks if a key exists in the cache database
func (c *Client) ExistsCache(ctx context.Context, key string) (bool, error) {
	return c.existsInDatabase(ctx, c.cache, "cache", key)
}

// TTLCache returns the time to live for a key in cache database
func (c *Client) TTLCache(ctx context.Context, key string) (time.Duration, error) {
	return c.ttlInDatabase(ctx, c.cache, "cache", key)
}

// === TASK OPERATIONS (DB 1) ===

// SetTask stores a key-value pair in the task database with optional expiration
func (c *Client) SetTask(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.setToDatabase(ctx, c.tasks, "task", key, value, expiration)
}

// GetTask retrieves a value by key from the task database
func (c *Client) GetTask(ctx context.Context, key string) (string, error) {
	return c.getFromDatabase(ctx, c.tasks, "task", key)
}

// GetTaskJSON retrieves a value and unmarshals it from JSON from task database
func (c *Client) GetTaskJSON(ctx context.Context, key string, dest interface{}) error {
	val, err := c.GetTask(ctx, key)
	if err != nil {
		return err
	}

	err = json.Unmarshal([]byte(val), dest)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON from task database: %w", err)
	}

	return nil
}

// DeleteTask removes keys from the task database
func (c *Client) DeleteTask(ctx context.Context, keys ...string) error {
	return c.deleteFromDatabase(ctx, c.tasks, "task", keys...)
}

// ExistsTask checks if a key exists in the task database
func (c *Client) ExistsTask(ctx context.Context, key string) (bool, error) {
	return c.existsInDatabase(ctx, c.tasks, "task", key)
}

// TTLTask returns the time to live for a key in task database
func (c *Client) TTLTask(ctx context.Context, key string) (time.Duration, error) {
	return c.ttlInDatabase(ctx, c.tasks, "task", key)
}

// === QUEUE OPERATIONS (DB 1) ===

// EnqueueTask adds a task to a queue (LPUSH)
func (c *Client) EnqueueTask(ctx context.Context, queue string, task interface{}) error {
	queueKey := c.config.KeyPrefix + "task:queue:" + queue

	// JSON encode the task if it's not a string
	var taskData string
	switch v := task.(type) {
	case string:
		taskData = v
	default:
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return fmt.Errorf("failed to marshal task to JSON: %w", err)
		}
		taskData = string(jsonBytes)
	}

	err := c.tasks.LPush(ctx, queueKey, taskData).Err()
	if err != nil {
		c.logger.Zap.Errorw("Redis LPUSH failed", "queue", queueKey, "error", err)
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	c.logger.Zap.Debugw("Task enqueued", "queue", queueKey)
	return nil
}

// DequeueTask removes and returns a task from a queue (RPOP)
func (c *Client) DequeueTask(ctx context.Context, queue string) (string, error) {
	queueKey := c.config.KeyPrefix + "task:queue:" + queue

	val, err := c.tasks.RPop(ctx, queueKey).Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("queue %s is empty", queueKey)
		}
		c.logger.Zap.Errorw("Redis RPOP failed", "queue", queueKey, "error", err)
		return "", fmt.Errorf("failed to dequeue task: %w", err)
	}

	c.logger.Zap.Debugw("Task dequeued", "queue", queueKey)
	return val, nil
}

// QueueLength returns the length of a queue
func (c *Client) QueueLength(ctx context.Context, queue string) (int64, error) {
	queueKey := c.config.KeyPrefix + "task:queue:" + queue

	length, err := c.tasks.LLen(ctx, queueKey).Result()
	if err != nil {
		c.logger.Zap.Errorw("Redis LLEN failed", "queue", queueKey, "error", err)
		return 0, fmt.Errorf("failed to get queue length: %w", err)
	}

	return length, nil
}

// === PUB/SUB OPERATIONS (DB 0 - Cache for real-time) ===

// Publish publishes a message to a channel for pub/sub
func (c *Client) Publish(ctx context.Context, channel string, message interface{}) error {
	// JSON encode the message if it's not a string
	var msg string
	switch v := message.(type) {
	case string:
		msg = v
	default:
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return fmt.Errorf("failed to marshal message to JSON: %w", err)
		}
		msg = string(jsonBytes)
	}

	err := c.cache.Publish(ctx, channel, msg).Err()
	if err != nil {
		c.logger.Zap.Errorw("Redis PUBLISH failed", "channel", channel, "error", err)
		return fmt.Errorf("failed to publish to channel %s: %w", channel, err)
	}

	c.logger.Zap.Debugw("Redis PUBLISH successful", "channel", channel)
	return nil
}

// Subscribe creates a subscription to channels
func (c *Client) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	return c.cache.Subscribe(ctx, channels...)
}

// === BACKWARD COMPATIBILITY METHODS ===

// Set stores a key-value pair with optional expiration (defaults to cache DB)
func (c *Client) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.SetCache(ctx, key, value, expiration)
}

// Get retrieves a value by key (defaults to cache DB)
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	return c.GetCache(ctx, key)
}

// GetJSON retrieves a value and unmarshals it from JSON (defaults to cache DB)
func (c *Client) GetJSON(ctx context.Context, key string, dest interface{}) error {
	return c.GetCacheJSON(ctx, key, dest)
}

// Delete removes a key (defaults to cache DB)
func (c *Client) Delete(ctx context.Context, keys ...string) error {
	return c.DeleteCache(ctx, keys...)
}

// Exists checks if a key exists (defaults to cache DB)
func (c *Client) Exists(ctx context.Context, key string) (bool, error) {
	return c.ExistsCache(ctx, key)
}

// SetEX stores a key-value pair with expiration in seconds (defaults to cache DB)
func (c *Client) SetEX(ctx context.Context, key string, value interface{}, seconds int64) error {
	return c.SetCache(ctx, key, value, time.Duration(seconds)*time.Second)
}

// TTL returns the time to live for a key (defaults to cache DB)
func (c *Client) TTL(ctx context.Context, key string) (time.Duration, error) {
	return c.TTLCache(ctx, key)
}

// === INTERNAL HELPER METHODS ===

func (c *Client) setToDatabase(ctx context.Context, db *redis.Client, dbName, key string, value interface{}, expiration time.Duration) error {
	// Create different prefixes for cache vs task operations to simulate separate databases
	var prefixedKey string
	if dbName == "cache" {
		prefixedKey = c.config.KeyPrefix + "cache:" + key
	} else {
		prefixedKey = c.config.KeyPrefix + "task:" + key
	}

	// Handle different value types
	var val interface{}
	switch v := value.(type) {
	case string, []byte, int, int64, float64, bool:
		val = v
	default:
		// JSON encode complex types
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return fmt.Errorf("failed to marshal value to JSON: %w", err)
		}
		val = jsonBytes
	}

	err := db.Set(ctx, prefixedKey, val, expiration).Err()
	if err != nil {
		c.logger.Zap.Errorw("Redis SET failed", "database", dbName, "key", prefixedKey, "error", err)
		return fmt.Errorf("failed to set key %s in %s database: %w", prefixedKey, dbName, err)
	}

	c.logger.Zap.Debugw("Redis SET successful", "database", dbName, "key", prefixedKey)
	return nil
}

func (c *Client) getFromDatabase(ctx context.Context, db *redis.Client, dbName, key string) (string, error) {
	// Create different prefixes for cache vs task operations
	var prefixedKey string
	if dbName == "cache" {
		prefixedKey = c.config.KeyPrefix + "cache:" + key
	} else {
		prefixedKey = c.config.KeyPrefix + "task:" + key
	}

	val, err := db.Get(ctx, prefixedKey).Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("key %s not found in %s database", prefixedKey, dbName)
		}
		c.logger.Zap.Errorw("Redis GET failed", "database", dbName, "key", prefixedKey, "error", err)
		return "", fmt.Errorf("failed to get key %s from %s database: %w", prefixedKey, dbName, err)
	}

	c.logger.Zap.Debugw("Redis GET successful", "database", dbName, "key", prefixedKey)
	return val, nil
}

func (c *Client) deleteFromDatabase(ctx context.Context, db *redis.Client, dbName string, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}

	// Add appropriate prefixes to all keys
	prefixedKeys := make([]string, len(keys))
	for i, key := range keys {
		if dbName == "cache" {
			prefixedKeys[i] = c.config.KeyPrefix + "cache:" + key
		} else {
			prefixedKeys[i] = c.config.KeyPrefix + "task:" + key
		}
	}

	deleted, err := db.Del(ctx, prefixedKeys...).Result()
	if err != nil {
		c.logger.Zap.Errorw("Redis DELETE failed", "database", dbName, "keys", prefixedKeys, "error", err)
		return fmt.Errorf("failed to delete keys from %s database: %w", dbName, err)
	}

	c.logger.Zap.Debugw("Redis DELETE successful", "database", dbName, "keys", prefixedKeys, "deleted_count", deleted)
	return nil
}

func (c *Client) existsInDatabase(ctx context.Context, db *redis.Client, dbName, key string) (bool, error) {
	// Create different prefixes for cache vs task operations
	var prefixedKey string
	if dbName == "cache" {
		prefixedKey = c.config.KeyPrefix + "cache:" + key
	} else {
		prefixedKey = c.config.KeyPrefix + "task:" + key
	}

	count, err := db.Exists(ctx, prefixedKey).Result()
	if err != nil {
		c.logger.Zap.Errorw("Redis EXISTS failed", "database", dbName, "key", prefixedKey, "error", err)
		return false, fmt.Errorf("failed to check key existence %s in %s database: %w", prefixedKey, dbName, err)
	}

	return count > 0, nil
}

func (c *Client) ttlInDatabase(ctx context.Context, db *redis.Client, dbName, key string) (time.Duration, error) {
	// Create different prefixes for cache vs task operations
	var prefixedKey string
	if dbName == "cache" {
		prefixedKey = c.config.KeyPrefix + "cache:" + key
	} else {
		prefixedKey = c.config.KeyPrefix + "task:" + key
	}

	ttl, err := db.TTL(ctx, prefixedKey).Result()
	if err != nil {
		c.logger.Zap.Errorw("Redis TTL failed", "database", dbName, "key", prefixedKey, "error", err)
		return 0, fmt.Errorf("failed to get TTL for key %s in %s database: %w", prefixedKey, dbName, err)
	}

	return ttl, nil
}

// Close closes both Redis connections
func (c *Client) Close() error {
	var err1, err2 error

	err1 = c.cache.Close()
	err2 = c.tasks.Close()

	if err1 != nil {
		return err1
	}
	return err2
}

// GetCacheClient returns the underlying cache redis client for advanced operations
func (c *Client) GetCacheClient() *redis.Client {
	return c.cache
}

// GetTaskClient returns the underlying task redis client for advanced operations
func (c *Client) GetTaskClient() *redis.Client {
	return c.tasks
}
