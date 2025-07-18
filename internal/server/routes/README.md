# Routes Organization

This directory contains route registration functions organized by feature/domain. Each route file focuses on a specific area of the application.

## Implemented Route Files

- `auth_routes.go` - Authentication and password recovery routes
- `user_routes.go` - User management and user-marina assignments
- `organization_routes.go` - Organization CRUD operations
- `marina_routes.go` - Marina management and marina-specific contacts
- `notification_routes.go` - Notification management (read-only API)
- `admin_routes.go` - Administrative functions
- `work_order_routes.go` - Work order management and operations
- `dme_routes.go` - DME system integration (credentials, system IDs, external API)
- `customer_routes.go` - Customer management and customer intake
- `role_routes.go` - Role management and permissions
- `criteria_routes.go` - Criteria management and search
- `gallery_routes.go` - Marina and vessel gallery management
- `document_routes.go` - Document upload and management (customer, boat, user)
- `esign_routes.go` - E-signature workflows (templates, documents, submissions)
- `message_routes.go` - Messaging system (customer and marina messages)
- `marina_usage_history_routes.go` - Marina usage tracking and analytics
- `invite_routes.go` - User invitation system (public and protected)
- `address_routes.go` - Address management and CRUD operations
- `email_routes.go` - Email sending services (HTML and template)
- `sms_routes.go` - SMS sending services
- `boat_routes.go` - Boat/vessel management and operations
- `plan_routes.go` - Plan management (notes/messages, storage, document plans)
- `test_routes.go` - Testing and development routes (permission and Redis tests)

## Pattern

Each route file follows this pattern:

```go
package routes

import (
	s "github.com/dockworks/dm-web-backend/internal/server"
	h "github.com/dockworks/dm-web-backend/internal/server/handlers"
	"github.com/labstack/echo/v4"
)

// RegisterFeatureRoutes registers all feature-related routes
func RegisterFeatureRoutes(server *s.Server, permissionProtected *echo.Group) {
	featureHandler := h.NewFeatureHandler(server)

	// Feature routes
	feature := permissionProtected.Group("/feature")
	feature.GET("/list", featureHandler.ListHandler)
	feature.POST("", featureHandler.CreateHandler)
	// ... other routes
}
```

## Route Organization Complete! 🎉

All major route groups have been successfully organized into dedicated files. The routing system is now fully modularized with each business domain having its own focused route file.

### Summary of Achievement
- **23 dedicated route files** covering all business domains
- **Main routes.go reduced** from 290+ lines to ~150 lines
- **100% route coverage** - no inline route definitions remaining
- **Clean architecture** with single-responsibility principle
- **Enhanced maintainability** and team collaboration

## Usage in Main Routes File

In `routes.go`, simply call the registration functions:

```go
// Feature routes
RegisterAuthRoutes(s, base, userHandler)
RegisterInviteRoutes(s, base, permissionProtected)
RegisterEsignRoutes(s, base, permissionProtected)
RegisterUserRoutes(s, permissionProtected)
RegisterCustomerRoutes(s, base, permissionProtected)
RegisterDMERoutes(s, base, permissionProtected, esignHandler)
RegisterRoleRoutes(s, permissionProtected)
RegisterOrganizationRoutes(s, permissionProtected)
RegisterMarinaRoutes(s, permissionProtected)
RegisterCriteriaRoutes(s, permissionProtected)
RegisterAddressRoutes(s, permissionProtected)
RegisterEmailRoutes(s, s.Config, permissionProtected)
RegisterSMSRoutes(s, s.Config, permissionProtected)
RegisterBoatRoutes(s, permissionProtected)
RegisterGalleryRoutes(s, permissionProtected)
RegisterDocumentRoutes(s, permissionProtected)
RegisterWorkOrderRoutes(s, permissionProtected)
RegisterMessageRoutes(s, permissionProtected)
RegisterMarinaUsageHistoryRoutes(s, protected)
RegisterPlanRoutes(s, protected)
RegisterTestRoutes(s, permissionProtected)
RegisterNotificationRoutes(s, permissionProtected)
RegisterAdminRoutes(s, permissionProtected)
```

## Benefits

- **Maintainability**: Each feature's routes are isolated
- **Readability**: Easier to find and understand routes for a specific feature
- **Scalability**: New features can be added without modifying existing route files
- **Testing**: Individual route groups can be tested in isolation
- **Team Development**: Multiple developers can work on different route files simultaneously

## Guidelines

1. **Single Responsibility**: Each file should handle routes for one business domain
2. **Consistent Naming**: Use `Register{Feature}Routes` as the function name
3. **Handler Creation**: Create handlers within the registration function
4. **Group Organization**: Use route groups to organize related endpoints
5. **Documentation**: Add comments explaining the purpose of each route group 