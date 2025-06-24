package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/dockworks/dm-web-backend/internal/requests"
	"github.com/dockworks/dm-web-backend/internal/responses"
	s "github.com/dockworks/dm-web-backend/internal/server"
	"github.com/labstack/echo/v4"
)

type RedisHandler struct {
	server *s.Server
}

func NewRedisHandler(server *s.Server) *RedisHandler {
	return &RedisHandler{server: server}
}

// TestRedisConnection
//
//	@Summary		Test Redis connection
//	@Description	Test the connection to Redis/Valkey and get basic information
//	@ID				redis-test-connection
//	@Tags			test
//	@Produce		json
//	@Success		200	{object}	responses.RedisInfoResponse	"Redis connection information"
//	@Failure		500	{object}	responses.Error				"Server error"
//	@Router			/api/v1/test/redis/connection [get]
func (h *RedisHandler) TestRedisConnection(c echo.Context) error {
	start := time.Now()

	// Test the connection
	err := h.server.Redis.Ping(c.Request().Context())
	latency := time.Since(start)

	if err != nil {
		h.server.Logger.Zap.Errorw("Redis connection test failed", "error", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Redis connection failed: "+err.Error()).JSON(c)
	}

	// Get some basic info about the Redis instance
	serverInfo := "Redis/Valkey is connected"

	response := responses.NewRedisInfoResponse(
		true,
		serverInfo,
		latency,
		h.server.Config.Redis.KeyPrefix,
		h.server.Config.Redis.MainDB,
	)

	return c.JSON(http.StatusOK, response)
}

// TestRedisOperations
//
//	@Summary		Test Redis operations
//	@Description	Test basic Redis operations (set, get, delete) with a sample key
//	@ID				redis-test-operations
//	@Tags			test
//	@Accept			json
//	@Produce		json
//	@Param			params	body		requests.RedisSetRequest	true	"Redis test data"
//	@Success		200		{object}	responses.RedisGetResponse	"Test completed successfully"
//	@Failure		400		{object}	responses.Error				"Validation error"
//	@Failure		500		{object}	responses.Error				"Server error"
//	@Router			/api/v1/test/redis/operations [post]
func (h *RedisHandler) TestRedisOperations(c echo.Context) error {
	var req requests.RedisSetRequest
	if err := c.Bind(&req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid request format").JSON(c)
	}

	if err := c.Validate(&req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err.Error()).JSON(c)
	}

	ctx := c.Request().Context()

	// Determine expiration
	var expiration time.Duration
	if req.Expiration != nil {
		expiration = time.Duration(*req.Expiration) * time.Second
	}

	// Step 1: Set the key-value pair
	err := h.server.Redis.Set(ctx, req.Key, req.Value, expiration)
	if err != nil {
		h.server.Logger.Zap.Errorw("Redis SET operation failed", "key", req.Key, "error", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to set key: "+err.Error()).JSON(c)
	}

	// Step 2: Get the value back
	value, err := h.server.Redis.Get(ctx, req.Key)
	if err != nil {
		h.server.Logger.Zap.Errorw("Redis GET operation failed", "key", req.Key, "error", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get key: "+err.Error()).JSON(c)
	}

	// Step 3: Get TTL information
	var ttl *time.Duration
	if expiration > 0 {
		ttlValue, err := h.server.Redis.TTL(ctx, req.Key)
		if err == nil && ttlValue > 0 {
			ttl = &ttlValue
		}
	}

	// Step 4: Try to parse the value back to its original type if it was JSON
	var parsedValue interface{}
	if err := json.Unmarshal([]byte(value), &parsedValue); err != nil {
		// If it's not JSON, return as string
		parsedValue = value
	}

	response := responses.NewRedisGetResponse(req.Key, parsedValue, ttl)
	response.Message = "Redis operations test completed successfully (SET -> GET)"

	h.server.Logger.Zap.Infow("Redis operations test completed",
		"key", req.Key,
		"has_ttl", ttl != nil,
		"value_type", fmt.Sprintf("%T", parsedValue))

	return c.JSON(http.StatusOK, response)
}

// Ping
//
//	@Summary		Ping Redis
//	@Description	Simple ping test for Redis connection
//	@ID				redis-ping
//	@Tags			test
//	@Produce		json
//	@Success		200	{object}	responses.RedisPingResponse	"Ping successful"
//	@Failure		500	{object}	responses.Error				"Server error"
//	@Router			/api/v1/test/redis/ping [get]
func (h *RedisHandler) Ping(c echo.Context) error {
	start := time.Now()

	err := h.server.Redis.Ping(c.Request().Context())
	latency := time.Since(start)

	if err != nil {
		h.server.Logger.Zap.Errorw("Redis ping failed", "error", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Redis ping failed: "+err.Error()).JSON(c)
	}

	response := responses.NewRedisPingResponse(latency)
	return c.JSON(http.StatusOK, response)
}

// TestDualDatabaseConnection tests both cache and task databases
// @Summary Test dual database connection
// @Description Tests connection to both cache (DB 0) and task (DB 1) databases
// @ID redis-test-dual-connection
// @Tags test
// @Accept json
// @Produce json
// @Success 200 {object} responses.RedisInfoResponse "Dual database connection status"
// @Failure 500 {object} responses.Error "Server error"
// @Router /api/v1/test/redis/dual-databases [get]
func (h *RedisHandler) TestDualDatabaseConnection(c echo.Context) error {
	ctx := c.Request().Context()
	start := time.Now()

	err := h.server.Redis.Ping(ctx)
	if err != nil {
		h.server.Logger.Zap.Errorw("Redis dual database ping failed", "error", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Redis connection failed: "+err.Error()).JSON(c)
	}

	latency := time.Since(start)

	response := responses.NewRedisInfoResponse(
		true,
		"Both cache and task databases are connected",
		latency,
		h.server.Config.Redis.KeyPrefix,
		h.server.Config.Redis.MainDB,
	)

	h.server.Logger.Zap.Debugw("Dual database connection test completed",
		"cache_db", h.server.Redis.Cache().Options().DB,
		"task_db", h.server.Redis.Tasks().Options().DB,
		"latency", latency)

	return c.JSON(http.StatusOK, response)
}

// TestCacheOperations tests cache database operations (DB 0)
// @Summary Test cache operations
// @Description Test cache database operations (DB 0) - set, get, delete, TTL
// @ID redis-test-cache-operations
// @Tags test
// @Accept json
// @Produce json
// @Param params body requests.RedisSetRequest true "Cache test data"
// @Success 200 {object} responses.RedisGetResponse "Cache operations completed successfully"
// @Failure 400 {object} responses.Error "Validation error"
// @Failure 500 {object} responses.Error "Server error"
// @Router /api/v1/test/redis/cache/operations [post]
func (h *RedisHandler) TestCacheOperations(c echo.Context) error {
	var req requests.RedisSetRequest
	if err := c.Bind(&req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid request format").JSON(c)
	}

	if err := c.Validate(&req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err.Error()).JSON(c)
	}

	ctx := c.Request().Context()

	// Determine expiration
	var expiration time.Duration
	if req.Expiration != nil {
		expiration = time.Duration(*req.Expiration) * time.Second
	}

	// Step 1: Set in cache database (DB 0)
	err := h.server.Redis.SetCache(ctx, req.Key, req.Value, expiration)
	if err != nil {
		h.server.Logger.Zap.Errorw("Cache SET operation failed", "key", req.Key, "error", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to set cache key: "+err.Error()).JSON(c)
	}

	// Step 2: Get from cache database
	var retrievedValue interface{}
	err = h.server.Redis.GetCacheJSON(ctx, req.Key, &retrievedValue)
	if err != nil {
		// Try as string if JSON parsing fails
		stringValue, stringErr := h.server.Redis.GetCache(ctx, req.Key)
		if stringErr != nil {
			h.server.Logger.Zap.Errorw("Cache GET operation failed", "key", req.Key, "error", err)
			return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get cache key: "+err.Error()).JSON(c)
		}
		retrievedValue = stringValue
	}

	// Step 3: Check TTL
	var ttl *time.Duration
	if expiration > 0 {
		ttlValue, err := h.server.Redis.TTLCache(ctx, req.Key)
		if err == nil && ttlValue > 0 {
			ttl = &ttlValue
		}
	}

	// Step 4: Test exists operation
	exists, err := h.server.Redis.ExistsCache(ctx, req.Key)
	if err != nil {
		h.server.Logger.Zap.Errorw("Cache EXISTS operation failed", "key", req.Key, "error", err)
	}

	response := responses.NewRedisGetResponse(req.Key, retrievedValue, ttl)
	response.Message = fmt.Sprintf("Cache operations completed successfully (DB 0) - exists: %t", exists)

	h.server.Logger.Zap.Infow("Cache operations test completed",
		"key", req.Key,
		"database", "cache (DB 0)",
		"exists", exists,
		"has_ttl", ttl != nil,
		"value_type", fmt.Sprintf("%T", retrievedValue))

	return c.JSON(http.StatusOK, response)
}

// TestTaskOperations tests task database operations (DB 1)
// @Summary Test task operations
// @Description Test task database operations (DB 1) - set, get, delete, TTL
// @ID redis-test-task-operations
// @Tags test
// @Accept json
// @Produce json
// @Param params body requests.RedisSetRequest true "Task test data"
// @Success 200 {object} responses.RedisGetResponse "Task operations completed successfully"
// @Failure 400 {object} responses.Error "Validation error"
// @Failure 500 {object} responses.Error "Server error"
// @Router /api/v1/test/redis/task/operations [post]
func (h *RedisHandler) TestTaskOperations(c echo.Context) error {
	var req requests.RedisSetRequest
	if err := c.Bind(&req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid request format").JSON(c)
	}

	if err := c.Validate(&req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err.Error()).JSON(c)
	}

	ctx := c.Request().Context()

	// Determine expiration
	var expiration time.Duration
	if req.Expiration != nil {
		expiration = time.Duration(*req.Expiration) * time.Second
	}

	// Step 1: Set in task database (DB 1)
	err := h.server.Redis.SetTask(ctx, req.Key, req.Value, expiration)
	if err != nil {
		h.server.Logger.Zap.Errorw("Task SET operation failed", "key", req.Key, "error", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to set task key: "+err.Error()).JSON(c)
	}

	// Step 2: Get from task database
	var retrievedValue interface{}
	err = h.server.Redis.GetTaskJSON(ctx, req.Key, &retrievedValue)
	if err != nil {
		// Try as string if JSON parsing fails
		stringValue, stringErr := h.server.Redis.GetTask(ctx, req.Key)
		if stringErr != nil {
			h.server.Logger.Zap.Errorw("Task GET operation failed", "key", req.Key, "error", err)
			return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get task key: "+err.Error()).JSON(c)
		}
		retrievedValue = stringValue
	}

	// Step 3: Check TTL
	var ttl *time.Duration
	if expiration > 0 {
		ttlValue, err := h.server.Redis.TTLTask(ctx, req.Key)
		if err == nil && ttlValue > 0 {
			ttl = &ttlValue
		}
	}

	// Step 4: Test exists operation
	exists, err := h.server.Redis.ExistsTask(ctx, req.Key)
	if err != nil {
		h.server.Logger.Zap.Errorw("Task EXISTS operation failed", "key", req.Key, "error", err)
	}

	response := responses.NewRedisGetResponse(req.Key, retrievedValue, ttl)
	response.Message = fmt.Sprintf("Task operations completed successfully (DB 1) - exists: %t", exists)

	h.server.Logger.Zap.Infow("Task operations test completed",
		"key", req.Key,
		"database", "task (DB 1)",
		"exists", exists,
		"has_ttl", ttl != nil,
		"value_type", fmt.Sprintf("%T", retrievedValue))

	return c.JSON(http.StatusOK, response)
}

// DeleteCacheKey deletes a key from cache database (DB 0)
// @Summary Delete cache key
// @Description Delete a key from cache database (DB 0)
// @ID redis-delete-cache-key
// @Tags test
// @Accept json
// @Produce json
// @Param key path string true "Key to delete"
// @Success 200 {object} responses.BaseResponse "Key deleted successfully"
// @Failure 500 {object} responses.Error "Server error"
// @Router /api/v1/test/redis/cache/key/{key} [delete]
func (h *RedisHandler) DeleteCacheKey(c echo.Context) error {
	key := c.Param("key")
	ctx := c.Request().Context()

	err := h.server.Redis.DeleteCache(ctx, key)
	if err != nil {
		h.server.Logger.Zap.Errorw("Cache DELETE operation failed", "key", key, "error", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to delete cache key: "+err.Error()).JSON(c)
	}

	response := responses.NewMessageResponse(http.StatusOK, "Cache key deleted successfully from DB 0")
	h.server.Logger.Zap.Infow("Cache key deleted", "key", key, "database", "cache (DB 0)")

	return response.JSON(c)
}

// DeleteTaskKey deletes a key from task database (DB 1)
// @Summary Delete task key
// @Description Delete a key from task database (DB 1)
// @ID redis-delete-task-key
// @Tags test
// @Accept json
// @Produce json
// @Param key path string true "Key to delete"
// @Success 200 {object} responses.BaseResponse "Key deleted successfully"
// @Failure 500 {object} responses.Error "Server error"
// @Router /api/v1/test/redis/task/key/{key} [delete]
func (h *RedisHandler) DeleteTaskKey(c echo.Context) error {
	key := c.Param("key")
	ctx := c.Request().Context()

	err := h.server.Redis.DeleteTask(ctx, key)
	if err != nil {
		h.server.Logger.Zap.Errorw("Task DELETE operation failed", "key", key, "error", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to delete task key: "+err.Error()).JSON(c)
	}

	response := responses.NewMessageResponse(http.StatusOK, "Task key deleted successfully from DB 1")
	h.server.Logger.Zap.Infow("Task key deleted", "key", key, "database", "task (DB 1)")

	return response.JSON(c)
}

// TestQueueOperations tests queue operations in task database (DB 1)
// @Summary Test queue operations
// @Description Test queue operations (enqueue, dequeue, length) in task database (DB 1)
// @ID redis-test-queue-operations
// @Tags test
// @Accept json
// @Produce json
// @Param params body requests.RedisSetRequest true "Queue test data"
// @Success 200 {object} responses.RedisGetResponse "Queue operations completed successfully"
// @Failure 400 {object} responses.Error "Validation error"
// @Failure 500 {object} responses.Error "Server error"
// @Router /api/v1/test/redis/queue/operations [post]
func (h *RedisHandler) TestQueueOperations(c echo.Context) error {
	var req requests.RedisSetRequest
	if err := c.Bind(&req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid request format").JSON(c)
	}

	if err := c.Validate(&req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err.Error()).JSON(c)
	}

	ctx := c.Request().Context()
	queueName := req.Key // Use the key as queue name

	// Step 1: Enqueue the value
	err := h.server.Redis.EnqueueTask(ctx, queueName, req.Value)
	if err != nil {
		h.server.Logger.Zap.Errorw("Queue ENQUEUE operation failed", "queue", queueName, "error", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to enqueue: "+err.Error()).JSON(c)
	}

	// Step 2: Check queue length
	length, err := h.server.Redis.QueueLength(ctx, queueName)
	if err != nil {
		h.server.Logger.Zap.Errorw("Queue LENGTH operation failed", "queue", queueName, "error", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get queue length: "+err.Error()).JSON(c)
	}

	// Step 3: Dequeue the value
	dequeuedValue, err := h.server.Redis.DequeueTask(ctx, queueName)
	if err != nil {
		h.server.Logger.Zap.Errorw("Queue DEQUEUE operation failed", "queue", queueName, "error", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to dequeue: "+err.Error()).JSON(c)
	}

	// Parse the dequeued value
	var parsedValue interface{}
	if err := json.Unmarshal([]byte(dequeuedValue), &parsedValue); err != nil {
		parsedValue = dequeuedValue
	}

	response := responses.NewRedisGetResponse(queueName, parsedValue, nil)
	response.Message = fmt.Sprintf("Queue operations completed successfully (DB 1) - queue length was: %d", length)

	h.server.Logger.Zap.Infow("Queue operations test completed",
		"queue", queueName,
		"database", "task (DB 1)",
		"queue_length", length,
		"value_type", fmt.Sprintf("%T", parsedValue))

	return c.JSON(http.StatusOK, response)
}

// GetCacheKey retrieves a key from cache database (DB 0)
// @Summary Get cache key
// @Description Get a key from cache database (DB 0)
// @ID redis-get-cache-key
// @Tags test
// @Accept json
// @Produce json
// @Param key path string true "Key to retrieve"
// @Success 200 {object} responses.RedisGetResponse "Key retrieved successfully"
// @Failure 404 {object} responses.Error "Key not found"
// @Failure 500 {object} responses.Error "Server error"
// @Router /api/v1/test/redis/cache/key/{key} [get]
func (h *RedisHandler) GetCacheKey(c echo.Context) error {
	key := c.Param("key")
	ctx := c.Request().Context()

	// Check if key exists
	exists, err := h.server.Redis.ExistsCache(ctx, key)
	if err != nil {
		h.server.Logger.Zap.Errorw("Cache EXISTS check failed", "key", key, "error", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to check cache key existence: "+err.Error()).JSON(c)
	}

	if !exists {
		return responses.NewErrorResponse(http.StatusNotFound, "Cache key not found").JSON(c)
	}

	// Try to get as JSON first
	var retrievedValue interface{}
	err = h.server.Redis.GetCacheJSON(ctx, key, &retrievedValue)
	if err != nil {
		// Try as string if JSON parsing fails
		stringValue, stringErr := h.server.Redis.GetCache(ctx, key)
		if stringErr != nil {
			h.server.Logger.Zap.Errorw("Cache GET operation failed", "key", key, "error", err)
			return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get cache key: "+err.Error()).JSON(c)
		}
		retrievedValue = stringValue
	}

	// Get TTL information
	var ttl *time.Duration
	ttlValue, err := h.server.Redis.TTLCache(ctx, key)
	if err == nil && ttlValue > 0 {
		ttl = &ttlValue
	}

	response := responses.NewRedisGetResponse(key, retrievedValue, ttl)
	response.Message = "Cache key retrieved successfully from DB 0"

	h.server.Logger.Zap.Infow("Cache key retrieved",
		"key", key,
		"database", "cache (DB 0)",
		"has_ttl", ttl != nil,
		"value_type", fmt.Sprintf("%T", retrievedValue))

	return c.JSON(http.StatusOK, response)
}

// GetTaskKey retrieves a key from task database (DB 1)
// @Summary Get task key
// @Description Get a key from task database (DB 1)
// @ID redis-get-task-key
// @Tags test
// @Accept json
// @Produce json
// @Param key path string true "Key to retrieve"
// @Success 200 {object} responses.RedisGetResponse "Key retrieved successfully"
// @Failure 404 {object} responses.Error "Key not found"
// @Failure 500 {object} responses.Error "Server error"
// @Router /api/v1/test/redis/task/key/{key} [get]
func (h *RedisHandler) GetTaskKey(c echo.Context) error {
	key := c.Param("key")
	ctx := c.Request().Context()

	// Check if key exists
	exists, err := h.server.Redis.ExistsTask(ctx, key)
	if err != nil {
		h.server.Logger.Zap.Errorw("Task EXISTS check failed", "key", key, "error", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to check task key existence: "+err.Error()).JSON(c)
	}

	if !exists {
		return responses.NewErrorResponse(http.StatusNotFound, "Task key not found").JSON(c)
	}

	// Try to get as JSON first
	var retrievedValue interface{}
	err = h.server.Redis.GetTaskJSON(ctx, key, &retrievedValue)
	if err != nil {
		// Try as string if JSON parsing fails
		stringValue, stringErr := h.server.Redis.GetTask(ctx, key)
		if stringErr != nil {
			h.server.Logger.Zap.Errorw("Task GET operation failed", "key", key, "error", err)
			return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get task key: "+err.Error()).JSON(c)
		}
		retrievedValue = stringValue
	}

	// Get TTL information
	var ttl *time.Duration
	ttlValue, err := h.server.Redis.TTLTask(ctx, key)
	if err == nil && ttlValue > 0 {
		ttl = &ttlValue
	}

	response := responses.NewRedisGetResponse(key, retrievedValue, ttl)
	response.Message = "Task key retrieved successfully from DB 1"

	h.server.Logger.Zap.Infow("Task key retrieved",
		"key", key,
		"database", "task (DB 1)",
		"has_ttl", ttl != nil,
		"value_type", fmt.Sprintf("%T", retrievedValue))

	return c.JSON(http.StatusOK, response)
}

// GetQueueLength retrieves the length of a queue from task database (DB 1)
// @Summary Get queue length
// @Description Get the length of a queue from task database (DB 1)
// @ID redis-get-queue-length
// @Tags test
// @Accept json
// @Produce json
// @Param queue path string true "Queue name"
// @Success 200 {object} responses.RedisGetResponse "Queue length retrieved successfully"
// @Failure 500 {object} responses.Error "Server error"
// @Router /api/v1/test/redis/queue/{queue}/length [get]
func (h *RedisHandler) GetQueueLength(c echo.Context) error {
	queueName := c.Param("queue")
	ctx := c.Request().Context()

	length, err := h.server.Redis.QueueLength(ctx, queueName)
	if err != nil {
		h.server.Logger.Zap.Errorw("Queue LENGTH operation failed", "queue", queueName, "error", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get queue length: "+err.Error()).JSON(c)
	}

	response := responses.NewRedisGetResponse(queueName, length, nil)
	response.Message = fmt.Sprintf("Queue length retrieved successfully from DB 1: %d items", length)

	h.server.Logger.Zap.Infow("Queue length retrieved",
		"queue", queueName,
		"database", "task (DB 1)",
		"length", length)

	return c.JSON(http.StatusOK, response)
}
