package middleware

import (
	"net/http"
	"strings"

	"github.com/dockworks/dm-web-backend/internal/config"
	"github.com/dockworks/dm-web-backend/internal/responses"
	"github.com/labstack/echo/v4"
)

// APIKeyMiddleware handles API key authentication
type APIKeyMiddleware struct {
	apiKey string
}

// NewAPIKeyMiddleware creates a new API key middleware instance
func NewAPIKeyMiddleware(cfg *config.Config) *APIKeyMiddleware {
	return &APIKeyMiddleware{
		apiKey: cfg.Auth.DMEApiKey,
	}
}

// ValidateAPIKey returns a middleware function that validates the x-api-key header
func (m *APIKeyMiddleware) ValidateAPIKey() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Skip validation if API key is not configured
			if m.apiKey == "" {
				return next(c)
			}

			// Get the API key from the x-api-key header
			apiKey := c.Request().Header.Get("x-api-key")

			// Check if header is missing
			if apiKey == "" {
				return responses.NewErrorResponse(
					http.StatusUnauthorized,
					"API key is required. Please provide a valid API key in the x-api-key header",
				).JSON(c)
			}

			// Validate the API key
			if !strings.EqualFold(strings.TrimSpace(apiKey), strings.TrimSpace(m.apiKey)) {
				return responses.NewErrorResponse(
					http.StatusUnauthorized,
					"Invalid API key provided",
				).JSON(c)
			}

			// API key is valid, continue to next handler
			return next(c)
		}
	}
}

// RequireAPIKey is a convenience function that creates and returns the middleware
func RequireAPIKey(cfg *config.Config) echo.MiddlewareFunc {
	middleware := NewAPIKeyMiddleware(cfg)
	return middleware.ValidateAPIKey()
}
