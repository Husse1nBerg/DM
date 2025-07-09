package handlers

import (
	"net/http"

	"github.com/dockworks/dm-web-backend/internal/guard"
	"github.com/dockworks/dm-web-backend/internal/responses"
	s "github.com/dockworks/dm-web-backend/internal/server"
	"github.com/dockworks/dm-web-backend/internal/server/middleware"
	"github.com/dockworks/dm-web-backend/pkg/token"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

type PermissionTestHandler struct {
	server *s.Server
}

func NewPermissionTestHandler(server *s.Server) *PermissionTestHandler {
	return &PermissionTestHandler{server: server}
}

// PermissionTestRequest represents the request for testing permissions
type PermissionTestRequest struct {
	Object string `json:"object" validate:"required" example:"customers"`
	Action string `json:"action" validate:"required" example:"read"`
}

// PermissionTestResponse represents the response for permission testing
type PermissionTestResponse struct {
	UserID     string `json:"user_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	MarinaID   string `json:"marina_id" example:"123e4567-e89b-12d3-a456-426614174001"`
	Object     string `json:"object" example:"customers"`
	Action     string `json:"action" example:"read"`
	HasAccess  bool   `json:"has_access" example:"true"`
	ModuleInfo struct {
		RequiredModule string `json:"required_module" example:"customerVessels"`
		ModuleEnabled  bool   `json:"module_enabled" example:"true"`
	} `json:"module_info"`
	Message string `json:"message" example:"Access granted"`
}

// TestPermission is a test endpoint to check user permissions
// @Summary Test user permissions
// @Description Test if the current user has permission to perform an action on an object
// @Tags permissions
// @Accept json
// @Produce json
// @Param request body PermissionTestRequest true "Permission test request"
// @Success 200 {object} PermissionTestResponse "Permission check result"
// @Failure 400 {object} responses.BaseResponse "Bad request"
// @Failure 401 {object} responses.BaseResponse "Unauthorized"
// @Failure 500 {object} responses.BaseResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /test/permissions [post]
func (h *PermissionTestHandler) TestPermission(c echo.Context) error {
	ctx := c.Request().Context()

	permissionService, err := guard.NewPermissionService(h.server.DB.Queries())
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to initialize permission service").JSON(c)
	}

	// Parse and validate request
	req := new(PermissionTestRequest)
	if err := c.Bind(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}
	if err := c.Validate(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	// Get user context from JWT token
	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(*token.JwtCustomClaims)
	userID := claims.ID.String()

	// Fetch the user's current marina_id from the database
	queries := h.server.DB.Queries()
	user, err := queries.GetUserByID(ctx, claims.ID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to load user: "+err.Error()).JSON(c)
	}
	marinaID := user.MarinaID.String()

	// Check permission
	hasAccess, err := permissionService.CanAccess(ctx, userID, marinaID, req.Object, req.Action)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to check permission: "+err.Error()).JSON(c)
	}

	// Get additional info for debugging
	requiredModule, moduleExists := guard.GetObjectModule(req.Object)
	var moduleEnabled bool
	var message string

	if !moduleExists {
		message = "Object not found in any module"
	} else {
		moduleEnabled, err = permissionService.MarinaHasModule(ctx, marinaID, requiredModule)
		if err != nil {
			return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to check marina modules: "+err.Error()).JSON(c)
		}

		if !moduleEnabled {
			message = "Marina module not enabled"
		} else if hasAccess {
			message = "Access granted"
		} else {
			message = "Access denied - insufficient permissions"
		}
	}

	// Build response
	response := PermissionTestResponse{
		UserID:    userID,
		MarinaID:  marinaID,
		Object:    req.Object,
		Action:    req.Action,
		HasAccess: hasAccess,
		ModuleInfo: struct {
			RequiredModule string `json:"required_module" example:"customerVessels"`
			ModuleEnabled  bool   `json:"module_enabled" example:"true"`
		}{
			RequiredModule: requiredModule,
			ModuleEnabled:  moduleEnabled,
		},
		Message: message,
	}

	return responses.NewSuccessResponse(response).JSON(c)
}

// TestPermissionBatch is a test endpoint to check multiple permissions at once
// @Summary Test multiple user permissions
// @Description Test if the current user has permissions for multiple object-action combinations
// @Tags permissions
// @Accept json
// @Produce json
// @Param requests body []PermissionTestRequest true "Array of permission test requests"
// @Success 200 {object} []PermissionTestResponse "Permission check results"
// @Failure 400 {object} responses.BaseResponse "Bad request"
// @Failure 401 {object} responses.BaseResponse "Unauthorized"
// @Failure 500 {object} responses.BaseResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /test/permissions/batch [post]
func (h *PermissionTestHandler) TestPermissionBatch(c echo.Context) error {
	ctx := c.Request().Context()

	// Parse and validate request
	var requests []PermissionTestRequest
	if err := c.Bind(&requests); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid request format").JSON(c)
	}

	if len(requests) == 0 {
		return responses.NewErrorResponse(http.StatusBadRequest, "At least one permission test is required").JSON(c)
	}

	// Get user context from JWT token
	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(*token.JwtCustomClaims)
	userID := claims.ID.String()

	// Fetch the user's current marina_id from the database
	queries := h.server.DB.Queries()
	user, err := queries.GetUserByID(ctx, claims.ID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to load user: "+err.Error()).JSON(c)
	}
	marinaID := user.MarinaID.String()

	permissionService, err := guard.NewPermissionService(h.server.DB.Queries())
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to initialize permission service").JSON(c)
	}

	var results []PermissionTestResponse

	// Process each permission test
	for _, req := range requests {
		// Check permission
		hasAccess, err := permissionService.CanAccess(ctx, userID, marinaID, req.Object, req.Action)
		if err != nil {
			return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to check permission: "+err.Error()).JSON(c)
		}

		// Get additional info
		requiredModule, moduleExists := guard.GetObjectModule(req.Object)
		var moduleEnabled bool
		var message string

		if !moduleExists {
			message = "Object not found in any module"
		} else {
			moduleEnabled, err = permissionService.MarinaHasModule(ctx, marinaID, requiredModule)
			if err != nil {
				return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to check marina modules: "+err.Error()).JSON(c)
			}

			if !moduleEnabled {
				message = "Marina module not enabled"
			} else if hasAccess {
				message = "Access granted"
			} else {
				message = "Access denied - insufficient permissions"
			}
		}

		// Build response for this test
		result := PermissionTestResponse{
			UserID:    userID,
			MarinaID:  marinaID,
			Object:    req.Object,
			Action:    req.Action,
			HasAccess: hasAccess,
			ModuleInfo: struct {
				RequiredModule string `json:"required_module" example:"customerVessels"`
				ModuleEnabled  bool   `json:"module_enabled" example:"true"`
			}{
				RequiredModule: requiredModule,
				ModuleEnabled:  moduleEnabled,
			},
			Message: message,
		}

		results = append(results, result)
	}

	return responses.NewSuccessResponse(results).JSON(c)
}

// GetUserPermissions returns all permissions for the current user
// @Summary Get current user permissions
// @Description Get all permissions for the currently logged in user in the current marina
// @Tags permissions
// @Produce json
// @Success 200 {object} map[string]interface{} "User permissions and available objects"
// @Failure 401 {object} responses.BaseResponse "Unauthorized"
// @Failure 500 {object} responses.BaseResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /test/permissions/user [get]
func (h *PermissionTestHandler) GetUserPermissions(c echo.Context) error {
	ctx := c.Request().Context()

	// Get user context from JWT token
	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(*token.JwtCustomClaims)
	userID := claims.ID.String()

	// Fetch the user's current marina_id from the database
	queries := h.server.DB.Queries()
	user, err := queries.GetUserByID(ctx, claims.ID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to load user: "+err.Error()).JSON(c)
	}
	marinaID := user.MarinaID.String()

	permissionService, err := guard.NewPermissionService(h.server.DB.Queries())
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to initialize permission service").JSON(c)
	}

	// Get user permissions
	permissions, err := permissionService.GetUserPermissions(ctx, userID, marinaID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get user permissions: "+err.Error()).JSON(c)
	}

	// Get enabled marina modules
	enabledModules, err := permissionService.GetMarinaEnabledModules(ctx, marinaID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get marina modules: "+err.Error()).JSON(c)
	}

	// Get available objects based on enabled modules
	var availableObjects []string
	for _, module := range enabledModules {
		if objects, exists := guard.GetModuleObjects(module); exists {
			availableObjects = append(availableObjects, objects...)
		}
	}

	response := map[string]interface{}{
		"user_id":           userID,
		"marina_id":         marinaID,
		"permissions":       permissions,
		"enabled_modules":   enabledModules,
		"available_objects": availableObjects,
	}

	return responses.NewSuccessResponse(response).JSON(c)
}

// GetRoutePermissions returns all configured route permissions for debugging
// @Summary Get route permission mappings
// @Description Get all configured route-to-permission mappings for debugging
// @Tags permissions
// @Produce json
// @Success 200 {object} map[string]interface{} "Route permission mappings"
// @Failure 401 {object} responses.BaseResponse "Unauthorized"
// @Failure 500 {object} responses.BaseResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /test/permissions/routes [get]
func (h *PermissionTestHandler) GetRoutePermissions(c echo.Context) error {
	// This would need to be passed from the middleware, but for now let's create a new instance
	permissionService, err := guard.NewPermissionService(h.server.DB.Queries())
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to initialize permission service").JSON(c)
	}

	// Create middleware instance to get route mappings
	permissionMiddleware := middleware.NewPermissionMiddleware(permissionService)
	routePermissions := permissionMiddleware.GetRoutePermissions()

	response := map[string]interface{}{
		"total_routes": len(routePermissions),
		"routes":       routePermissions,
	}

	return responses.NewSuccessResponse(response).JSON(c)
}
