package requests

// RedisSetRequest represents a request to set a key-value pair in Redis
type RedisSetRequest struct {
	Key        string      `json:"key" validate:"required"`
	Value      interface{} `json:"value" validate:"required"`
	Expiration *int64      `json:"expiration,omitempty"` // expiration in seconds, optional
}

// RedisGetRequest represents a request to get a value from Redis
type RedisGetRequest struct {
	Key string `json:"key" validate:"required"`
}

// RedisDeleteRequest represents a request to delete keys from Redis
type RedisDeleteRequest struct {
	Keys []string `json:"keys" validate:"required,min=1"`
}
