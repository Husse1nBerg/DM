package responses

import "time"

// RedisSetResponse represents the response for a Redis SET operation
type RedisSetResponse struct {
	Success bool   `json:"success"`
	Key     string `json:"key"`
	Message string `json:"message"`
}

// RedisGetResponse represents the response for a Redis GET operation
type RedisGetResponse struct {
	Success bool        `json:"success"`
	Key     string      `json:"key"`
	Value   interface{} `json:"value,omitempty"`
	Message string      `json:"message,omitempty"`
	TTL     *string     `json:"ttl,omitempty"` // time to live
}

// RedisDeleteResponse represents the response for a Redis DELETE operation
type RedisDeleteResponse struct {
	Success bool     `json:"success"`
	Keys    []string `json:"keys"`
	Message string   `json:"message"`
}

// RedisPingResponse represents the response for a Redis PING operation
type RedisPingResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Latency string `json:"latency,omitempty"`
}

// RedisInfoResponse represents general Redis information
type RedisInfoResponse struct {
	Success    bool   `json:"success"`
	Connected  bool   `json:"connected"`
	ServerInfo string `json:"server_info,omitempty"`
	Message    string `json:"message,omitempty"`
	Latency    string `json:"latency,omitempty"`
	KeyPrefix  string `json:"key_prefix"`
}

// NewRedisSetResponse creates a new successful Redis SET response
func NewRedisSetResponse(key string) *RedisSetResponse {
	return &RedisSetResponse{
		Success: true,
		Key:     key,
		Message: "Key set successfully",
	}
}

// NewRedisGetResponse creates a new successful Redis GET response
func NewRedisGetResponse(key string, value interface{}, ttl *time.Duration) *RedisGetResponse {
	response := &RedisGetResponse{
		Success: true,
		Key:     key,
		Value:   value,
		Message: "Key retrieved successfully",
	}

	if ttl != nil && *ttl > 0 {
		ttlStr := ttl.String()
		response.TTL = &ttlStr
	}

	return response
}

// NewRedisDeleteResponse creates a new successful Redis DELETE response
func NewRedisDeleteResponse(keys []string) *RedisDeleteResponse {
	return &RedisDeleteResponse{
		Success: true,
		Keys:    keys,
		Message: "Keys deleted successfully",
	}
}

// NewRedisPingResponse creates a new successful Redis PING response
func NewRedisPingResponse(latency time.Duration) *RedisPingResponse {
	return &RedisPingResponse{
		Success: true,
		Message: "Redis connection is healthy",
		Latency: latency.String(),
	}
}

// NewRedisInfoResponse creates a new Redis info response
func NewRedisInfoResponse(connected bool, serverInfo string, latency time.Duration, keyPrefix string) *RedisInfoResponse {
	return &RedisInfoResponse{
		Success:    true,
		Connected:  connected,
		ServerInfo: serverInfo,
		Message:    "Redis information retrieved successfully",
		Latency:    latency.String(),
		KeyPrefix:  keyPrefix,
	}
}
