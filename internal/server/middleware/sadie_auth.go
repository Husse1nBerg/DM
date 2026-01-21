package middleware

import (
	"strings"

	"github.com/dockworks/dm-web-backend/internal/config"
	"github.com/dockworks/dm-web-backend/internal/responses"
	"github.com/labstack/echo/v4"
)

// SadieAuthMiddleware handles SADIE core secret authentication
type SadieAuthMiddleware struct {
	secret string
}

// NewSadieAuthMiddleware creates a new SADIE auth middleware instance
func NewSadieAuthMiddleware(cfg *config.Config) *SadieAuthMiddleware {
	return &SadieAuthMiddleware{
		secret: cfg.Sadie.ClientSecret,
	}
}

// ValidateSadieSecret returns a middleware function that validates the x-sadie-core-secret header
func (m *SadieAuthMiddleware) ValidateSadieSecret() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Skip validation if secret is not configured
			if m.secret == "" {
				return next(c)
			}

			// Check if the path is a SADIE route (either /sadie/* or root-level SADIE routes)
			path := c.Request().URL.Path
			isSadieRoute := strings.HasPrefix(path, "/sadie") ||
				path == "/getAvailableSlips" ||
				path == "/makeReservation" ||
				path == "/getAssistantPhoneNumber" ||
				path == "/getPhoneNumbers" ||
				path == "/updateAgentWebhook"

			if !isSadieRoute {
				return next(c)
			}

			// Get the secret from the x-sadie-core-secret header
			clientSecret := c.Request().Header.Get("x-sadie-core-secret")

			// Check if header is missing or invalid
			if clientSecret == "" || !strings.EqualFold(strings.TrimSpace(clientSecret), strings.TrimSpace(m.secret)) {
				return responses.NewSadieErrorResponse(
					"Authentication failed",
					"Access to the API requires valid credentials",
					[]string{
						"Apologize for the inconvenience",
						"Explain that there's a system authentication issue",
						"Transfer the call to technical support",
					},
					map[string]interface{}{"retryAllowed": false},
				).JSON(c)
			}

			// Secret is valid, continue to next handler
			return next(c)
		}
	}
}

// RequireSadieAuth is a convenience function that creates and returns the middleware
func RequireSadieAuth(cfg *config.Config) echo.MiddlewareFunc {
	middleware := NewSadieAuthMiddleware(cfg)
	return middleware.ValidateSadieSecret()
}
