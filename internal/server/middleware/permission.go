package middleware

import (
	"net/http"

	"github.com/dockworks/dm-web-backend/internal/db"
	"github.com/dockworks/dm-web-backend/internal/guard"
	"github.com/dockworks/dm-web-backend/internal/responses"
	"github.com/dockworks/dm-web-backend/pkg/token"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

// RoutePermission defines the permission requirements for a route
type RoutePermission struct {
	Object string `json:"object"` // The object being accessed (e.g., "customers", "vessels")
	Action string `json:"action"` // The action being performed (e.g., "read", "write", "delete")
}

// PermissionMiddleware handles automatic permission validation for routes
type PermissionMiddleware struct {
	permissionService *guard.PermissionService
	routeMap          map[string]RoutePermission
}

// NewPermissionMiddleware creates a new permission middleware instance
func NewPermissionMiddleware(permissionService *guard.PermissionService) *PermissionMiddleware {
	return &PermissionMiddleware{
		permissionService: permissionService,
		routeMap:          buildRoutePermissionMap(),
	}
}

// InitializePermissionMiddleware creates and configures the permission middleware
func InitializePermissionMiddleware(database *db.Queries) (*PermissionMiddleware, error) {
	// Create permission service
	permissionService, err := guard.NewPermissionService(database)
	if err != nil {
		return nil, err
	}

	// Create permission middleware
	permissionMiddleware := NewPermissionMiddleware(permissionService)

	// You can customize the middleware here if needed
	// For example:
	// permissionMiddleware.WithCustomPermission("GET /api/v1/special-route", "special_object", "read")
	// permissionMiddleware.SkipPermission("GET /api/v1/public-route")

	return permissionMiddleware, nil
}

// RequirePermission returns a middleware function that validates permissions for the current route
func (pm *PermissionMiddleware) RequirePermission() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get route information - use the route pattern, not the actual path
			// This handles parameterized routes like /api/v1/customers/:id
			route := c.Request().Method + " " + c.Path()
			// Check if this route requires permission validation
			permission, requiresPermission := pm.routeMap[route]
			if !requiresPermission {
				// Route doesn't require permission validation, continue
				return next(c)
			}

			// Get user context from JWT token
			userToken := c.Get("user").(*jwt.Token)
			if userToken == nil {
				return responses.NewErrorResponse(http.StatusUnauthorized, "Authentication required").JSON(c)
			}

			claims := userToken.Claims.(*token.JwtCustomClaims)
			userID := claims.ID.String()
			marinaID := claims.MarinaId.String()

			// Check permission
			hasAccess, err := pm.permissionService.CanAccess(c.Request().Context(), userID, marinaID, permission.Object, permission.Action)
			if err != nil {
				return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to check permissions: "+err.Error()).JSON(c)
			}

			if !hasAccess {
				return responses.NewErrorResponse(http.StatusForbidden, "Insufficient permissions to access this resource").JSON(c)
			}

			// Permission granted, continue to handler
			return next(c)
		}
	}
}

// buildRoutePermissionMap creates the mapping between routes and required permissions
func buildRoutePermissionMap() map[string]RoutePermission {
	routeMap := make(map[string]RoutePermission)

	// User routes - using "users" object from core module
	routeMap["GET /api/v1/user/profile"] = RoutePermission{Object: "profile", Action: "read"}
	routeMap["GET /api/v1/user/list"] = RoutePermission{Object: "users", Action: "read"}
	routeMap["POST /api/v1/user"] = RoutePermission{Object: "users", Action: "create"}
	routeMap["GET /api/v1/user/:userId"] = RoutePermission{Object: "users", Action: "read"}
	routeMap["PUT /api/v1/user/:userId"] = RoutePermission{Object: "users", Action: "write"}
	routeMap["DELETE /api/v1/user/:userId"] = RoutePermission{Object: "users", Action: "delete"}
	routeMap["POST /api/v1/user/reset-password"] = RoutePermission{Object: "profile", Action: "write"}
	routeMap["POST /api/v1/user/customer-portal"] = RoutePermission{Object: "users", Action: "create"}
	routeMap["GET /api/v1/user/customer-portal/:customerId"] = RoutePermission{Object: "users", Action: "read"}
	routeMap["POST /api/v1/user/invite"] = RoutePermission{Object: "users", Action: "create"}
	routeMap["GET /api/v1/user/role/:roleId"] = RoutePermission{Object: "users", Action: "read"}
	routeMap["GET /api/v1/user/organization/:organizationId"] = RoutePermission{Object: "users", Action: "read"}
	routeMap["GET /api/v1/user/marina/:marinaId"] = RoutePermission{Object: "users", Action: "read"}
	routeMap["GET /api/v1/user/marina/:marinaId/assigned"] = RoutePermission{Object: "users", Action: "read"}
	routeMap["POST /api/v1/user/marina/assign"] = RoutePermission{Object: "users", Action: "write"}
	routeMap["POST /api/v1/user/marina/unassign"] = RoutePermission{Object: "users", Action: "write"}

	// Role routes - using "roles" object from core module
	routeMap["GET /api/v1/role/list"] = RoutePermission{Object: "roles", Action: "read"}
	routeMap["POST /api/v1/role"] = RoutePermission{Object: "roles", Action: "create"}
	routeMap["GET /api/v1/role/:roleId"] = RoutePermission{Object: "roles", Action: "read"}
	routeMap["PUT /api/v1/role/:roleId"] = RoutePermission{Object: "roles", Action: "write"}
	routeMap["DELETE /api/v1/role/:roleId"] = RoutePermission{Object: "roles", Action: "delete"}
	routeMap["GET /api/v1/role/name"] = RoutePermission{Object: "roles", Action: "read"}

	// Organization routes - using "organizations" object from core module
	routeMap["POST /api/v1/organizations"] = RoutePermission{Object: "organizations", Action: "create"}
	routeMap["GET /api/v1/organizations"] = RoutePermission{Object: "organizations", Action: "read"}
	routeMap["GET /api/v1/organizations/by-email"] = RoutePermission{Object: "organizations", Action: "read"}
	routeMap["GET /api/v1/organizations/:id"] = RoutePermission{Object: "organizations", Action: "read"}
	routeMap["GET /api/v1/organizations/:id/with-address"] = RoutePermission{Object: "organizations", Action: "read"}
	routeMap["PUT /api/v1/organizations/:id"] = RoutePermission{Object: "organizations", Action: "write"}
	routeMap["PUT /api/v1/organizations/:id/with-address"] = RoutePermission{Object: "organizations", Action: "write"}
	routeMap["DELETE /api/v1/organizations/:id"] = RoutePermission{Object: "organizations", Action: "delete"}

	// Marina routes - using "marinas" object from core module
	routeMap["POST /api/v1/marinas"] = RoutePermission{Object: "marinas", Action: "create"}
	routeMap["GET /api/v1/marinas"] = RoutePermission{Object: "marinas", Action: "read"}
	routeMap["GET /api/v1/marinas/by-email"] = RoutePermission{Object: "marinas", Action: "read"}
	routeMap["GET /api/v1/marinas/organization/:organizationId"] = RoutePermission{Object: "marinas", Action: "read"}
	routeMap["GET /api/v1/marinas/user/:userId"] = RoutePermission{Object: "marinas", Action: "read"}
	routeMap["GET /api/v1/marinas/user"] = RoutePermission{Object: "marinas", Action: "read"}
	routeMap["GET /api/v1/marinas/:id"] = RoutePermission{Object: "marinas", Action: "read"}
	routeMap["GET /api/v1/marinas/:id/with-address"] = RoutePermission{Object: "marinas", Action: "read"}
	routeMap["PUT /api/v1/marinas/:id"] = RoutePermission{Object: "marinas", Action: "write"}
	routeMap["PUT /api/v1/marinas/:id/with-address"] = RoutePermission{Object: "marinas", Action: "write"}
	routeMap["DELETE /api/v1/marinas/:id"] = RoutePermission{Object: "marinas", Action: "delete"}
	routeMap["GET /api/v1/marinas/:id/contacts"] = RoutePermission{Object: "marinas", Action: "read"}
	routeMap["POST /api/v1/marinas/:id/contacts"] = RoutePermission{Object: "marinas", Action: "create"}
	routeMap["PUT /api/v1/marinas/:id/contacts/:contactId"] = RoutePermission{Object: "marinas", Action: "write"}
	routeMap["DELETE /api/v1/marinas/:id/contacts/:contactId"] = RoutePermission{Object: "marinas", Action: "delete"}

	// Address routes - using "addresses" object from core module
	routeMap["POST /api/v1/addresses"] = RoutePermission{Object: "addresses", Action: "create"}
	routeMap["GET /api/v1/addresses/:id"] = RoutePermission{Object: "addresses", Action: "read"}
	routeMap["PUT /api/v1/addresses/:id"] = RoutePermission{Object: "addresses", Action: "write"}

	// DME routes - using "equipment" from inventory management
	routeMap["POST /api/v1/dme/credentials"] = RoutePermission{Object: "admin", Action: "create"}
	routeMap["GET /api/v1/dme/credentials/organization/:organizationId"] = RoutePermission{Object: "admin", Action: "read"}
	routeMap["PUT /api/v1/dme/credentials/organization/:organizationId"] = RoutePermission{Object: "admin", Action: "write"}
	routeMap["DELETE /api/v1/dme/credentials/organization/:organizationId"] = RoutePermission{Object: "admin", Action: "delete"}
	routeMap["POST /api/v1/dme/sysids"] = RoutePermission{Object: "admin", Action: "create"}
	routeMap["GET /api/v1/dme/sysids"] = RoutePermission{Object: "admin", Action: "read"}
	routeMap["GET /api/v1/dme/sysids/:id"] = RoutePermission{Object: "admin", Action: "read"}
	routeMap["PUT /api/v1/dme/sysids/:id"] = RoutePermission{Object: "admin", Action: "write"}
	routeMap["DELETE /api/v1/dme/sysids/:id"] = RoutePermission{Object: "admin", Action: "delete"}
	routeMap["GET /api/v1/dme/sysids/system/:systemId"] = RoutePermission{Object: "admin", Action: "read"}
	routeMap["GET /api/v1/dme/sysids/organization/:organizationId"] = RoutePermission{Object: "admin", Action: "read"}
	routeMap["GET /api/v1/dme/sysids/marina/:marinaId"] = RoutePermission{Object: "admin", Action: "read"}
	routeMap["PATCH /api/v1/dme/sysids/:id/link"] = RoutePermission{Object: "admin", Action: "write"}
	routeMap["PATCH /api/v1/dme/sysids/:id/unlink"] = RoutePermission{Object: "admin", Action: "write"}

	// Customer routes
	routeMap["GET /api/v1/customers/list"] = RoutePermission{Object: "customers", Action: "read"}
	routeMap["GET /api/v1/customers/list-short"] = RoutePermission{Object: "customers", Action: "read"}
	routeMap["GET /api/v1/customers/retrieve"] = RoutePermission{Object: "customers", Action: "read"}
	routeMap["GET /api/v1/customers/search"] = RoutePermission{Object: "customers", Action: "read"}
	routeMap["POST /api/v1/customers/create"] = RoutePermission{Object: "customers", Action: "create"}
	routeMap["POST /api/v1/customers/update"] = RoutePermission{Object: "customers", Action: "write"}
	routeMap["GET /api/v1/customers/settings"] = RoutePermission{Object: "customers", Action: "read"}
	routeMap["POST /api/v1/customers/settings"] = RoutePermission{Object: "customers", Action: "write"}

	// Vessel/Boat routes (mapping boats to vessels object)
	routeMap["GET /api/v1/boats/list"] = RoutePermission{Object: "vessels", Action: "read"}
	routeMap["GET /api/v1/boats/retrieve"] = RoutePermission{Object: "vessels", Action: "read"}
	routeMap["GET /api/v1/boats/customer"] = RoutePermission{Object: "vessels", Action: "read"}
	routeMap["GET /api/v1/boats/search"] = RoutePermission{Object: "vessels", Action: "read"}
	routeMap["POST /api/v1/boats/create"] = RoutePermission{Object: "vessels", Action: "create"}
	routeMap["POST /api/v1/boats/update"] = RoutePermission{Object: "vessels", Action: "write"}

	// Work Order routes (service management)
	routeMap["GET /api/v1/work-orders/list"] = RoutePermission{Object: "work_orders", Action: "read"}
	routeMap["GET /api/v1/work-orders/retrieve"] = RoutePermission{Object: "work_orders", Action: "read"}
	routeMap["GET /api/v1/work-orders/search"] = RoutePermission{Object: "work_orders", Action: "read"}
	routeMap["GET /api/v1/work-orders/customer"] = RoutePermission{Object: "work_orders", Action: "read"}
	routeMap["GET /api/v1/work-orders/operations"] = RoutePermission{Object: "work_orders", Action: "read"}
	routeMap["GET /api/v1/work-orders/completed"] = RoutePermission{Object: "work_orders", Action: "read"}
	routeMap["POST /api/v1/work-orders/create"] = RoutePermission{Object: "work_orders", Action: "create"}
	routeMap["POST /api/v1/work-orders/update"] = RoutePermission{Object: "work_orders", Action: "write"}
	routeMap["POST /api/v1/work-orders/create-from-estimate"] = RoutePermission{Object: "work_orders", Action: "create"}
	routeMap["POST /api/v1/work-orders/delete-operation"] = RoutePermission{Object: "work_orders", Action: "delete"}

	// Gallery routes (vessel and marina management)
	routeMap["POST /api/v1/gallery/marina"] = RoutePermission{Object: "marina_gallery", Action: "write"}
	routeMap["GET /api/v1/gallery/marina/:marinaId"] = RoutePermission{Object: "marina_gallery", Action: "read"}
	routeMap["GET /api/v1/gallery/marina/item/:id"] = RoutePermission{Object: "marina_gallery", Action: "read"}
	routeMap["PUT /api/v1/gallery/marina/item/:id"] = RoutePermission{Object: "marina_gallery", Action: "write"}
	routeMap["DELETE /api/v1/gallery/marina/item/:id"] = RoutePermission{Object: "marina_gallery", Action: "delete"}

	routeMap["POST /api/v1/gallery/boat"] = RoutePermission{Object: "boat_gallery", Action: "write"}
	routeMap["GET /api/v1/gallery/boat/:boatId"] = RoutePermission{Object: "boat_gallery", Action: "read"}
	routeMap["GET /api/v1/gallery/boat/item/:id"] = RoutePermission{Object: "boat_gallery", Action: "read"}
	routeMap["PUT /api/v1/gallery/boat/item/:id"] = RoutePermission{Object: "boat_gallery", Action: "write"}
	routeMap["DELETE /api/v1/gallery/boat/item/:id"] = RoutePermission{Object: "boat_gallery", Action: "delete"}

	// Document routes
	routeMap["POST /api/v1/documents/customer"] = RoutePermission{Object: "documents", Action: "write"}
	routeMap["GET /api/v1/documents/customer"] = RoutePermission{Object: "documents", Action: "read"}
	routeMap["POST /api/v1/documents/boat"] = RoutePermission{Object: "documents", Action: "write"}
	routeMap["GET /api/v1/documents/boat"] = RoutePermission{Object: "documents", Action: "read"}
	routeMap["POST /api/v1/documents/user"] = RoutePermission{Object: "documents", Action: "write"}
	routeMap["GET /api/v1/documents/user"] = RoutePermission{Object: "documents", Action: "read"}
	routeMap["GET /api/v1/documents/:id"] = RoutePermission{Object: "documents", Action: "read"}
	routeMap["DELETE /api/v1/documents/:id"] = RoutePermission{Object: "documents", Action: "delete"}

	// Message routes (communications)
	routeMap["POST /api/v1/message/customer"] = RoutePermission{Object: "messages", Action: "write"}
	routeMap["GET /api/v1/message/customer"] = RoutePermission{Object: "messages", Action: "read"}
	routeMap["PUT /api/v1/message/customer"] = RoutePermission{Object: "messages", Action: "write"}
	routeMap["DELETE /api/v1/message/customer"] = RoutePermission{Object: "messages", Action: "delete"}
	routeMap["POST /api/v1/message/marina"] = RoutePermission{Object: "messages", Action: "write"}
	routeMap["GET /api/v1/message/marina"] = RoutePermission{Object: "messages", Action: "read"}
	routeMap["PUT /api/v1/message/marina"] = RoutePermission{Object: "messages", Action: "write"}
	routeMap["DELETE /api/v1/message/marina"] = RoutePermission{Object: "messages", Action: "delete"}
	routeMap["GET /api/v1/message/get"] = RoutePermission{Object: "messages", Action: "read"}

	// Email and SMS routes (communications)
	routeMap["POST /api/v1/email/send-html"] = RoutePermission{Object: "admin", Action: "write"}
	routeMap["POST /api/v1/email/send-template"] = RoutePermission{Object: "admin", Action: "write"}
	routeMap["POST /api/v1/sms/send"] = RoutePermission{Object: "admin", Action: "write"}

	// Permission test routes (for testing the permission system)
	routeMap["POST /api/v1/test/permissions"] = RoutePermission{Object: "admin", Action: "read"}
	routeMap["GET /api/v1/test/permissions/user"] = RoutePermission{Object: "admin", Action: "read"}
	routeMap["GET /api/v1/test/permissions/routes"] = RoutePermission{Object: "admin", Action: "read"}

	return routeMap
}

// WithCustomPermission allows overriding the default route mapping for specific routes
func (pm *PermissionMiddleware) WithCustomPermission(route string, object string, action string) *PermissionMiddleware {
	pm.routeMap[route] = RoutePermission{Object: object, Action: action}
	return pm
}

// SkipPermission allows skipping permission validation for specific routes
func (pm *PermissionMiddleware) SkipPermission(route string) *PermissionMiddleware {
	delete(pm.routeMap, route)
	return pm
}

// GetRoutePermissions returns all configured route permissions (useful for debugging)
func (pm *PermissionMiddleware) GetRoutePermissions() map[string]RoutePermission {
	return pm.routeMap
}
