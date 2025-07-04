package guard

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/dockworks/dm-web-backend/internal/db"
	"github.com/google/uuid"
)

// PermissionService handles authorization logic combining Casbin RBAC with marina module validation
type PermissionService struct {
	model   model.Model
	adapter *MarinaCasbinAdapter
}

// NewPermissionService creates a new permission service instance
func NewPermissionService(database *db.Queries) (*PermissionService, error) {
	// Create Casbin model from string
	modelText := `
[request_definition]
r = sub, dom, obj, act

[policy_definition]
p = sub, dom, obj, act

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = r.sub == p.sub && r.dom == p.dom && r.obj == p.obj && r.act == p.act
`

	m, err := model.NewModelFromString(modelText)
	if err != nil {
		return nil, fmt.Errorf("failed to create casbin model: %w", err)
	}

	adapter := NewMarinaCasbinAdapter(database)

	return &PermissionService{
		model:   m,
		adapter: adapter,
	}, nil
}

// CanAccess checks if a user can perform an action on an object in a specific marina
// This is the main method for permission checking
func (s *PermissionService) CanAccess(ctx context.Context, userID, marinaID, object, action string) (bool, error) {
	isAdmin, err := s.IsAdmin(ctx, userID, marinaID)
	if err != nil {
		return false, fmt.Errorf("failed to check if user is admin: %w", err)
	}
	if isAdmin {
		return true, nil
	}
	// Step 1: Validate that the object is defined in our module system
	if !IsValidObject(object) {
		return false, fmt.Errorf("unknown object: %s", object)
	}

	// Step 2: Check if the marina has the module enabled for this object
	hasModule, err := s.marinaHasModuleForObject(ctx, marinaID, object)
	if err != nil {
		return false, fmt.Errorf("failed to check marina modules: %w", err)
	}
	if !hasModule {
		return false, nil // Marina doesn't have the required module enabled
	}

	// Step 3: Load policies for this specific user and marina
	policies, err := s.adapter.LoadPoliciesForUser(ctx, userID, marinaID)
	if err != nil {
		return false, fmt.Errorf("failed to load user policies: %w", err)
	}

	// Step 4: Create a per-request enforcer to avoid race conditions
	enforcer, err := casbin.NewEnforcer(s.model)
	if err != nil {
		return false, fmt.Errorf("failed to create per-request enforcer: %w", err)
	}

	// Step 5: Load policies into the per-request enforcer
	for _, policy := range policies {
		if len(policy) == 4 {
			// Regular policy: sub, dom, obj, act
			enforcer.AddPolicy(policy[0], policy[1], policy[2], policy[3])
		}
		// Note: We no longer handle role policies since we only use direct user policies
	}

	// Step 6: Use Casbin to check the permission
	allowed, err := enforcer.Enforce(userID, marinaID, object, action)
	if err != nil {
		return false, fmt.Errorf("failed to enforce permission: %w", err)
	}

	return allowed, nil
}

// marinaHasModuleForObject checks if the marina has the module enabled for the given object
func (s *PermissionService) marinaHasModuleForObject(ctx context.Context, marinaID, object string) (bool, error) {
	// Find which module contains this object
	moduleName, exists := GetObjectModule(object)
	if !exists {
		return false, fmt.Errorf("object %s is not defined in any module", object)
	}

	// Check if marina has this module enabled
	return s.MarinaHasModule(ctx, marinaID, moduleName)
}

// GetMarinaEnabledModules returns all enabled modules for a marina
func (s *PermissionService) GetMarinaEnabledModules(ctx context.Context, marinaID string) ([]string, error) {
	// Convert marinaID to UUID
	marinaUUID, err := uuid.Parse(marinaID)
	if err != nil {
		return nil, fmt.Errorf("invalid marina ID: %w", err)
	}

	// Get marina from database
	marina, err := s.adapter.db.GetMarinaByID(ctx, marinaUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get marina: %w", err)
	}

	// Parse marina modules (boolean format only)
	var marinaModules map[string]bool
	if err := json.Unmarshal(marina.Modules, &marinaModules); err != nil {
		return nil, fmt.Errorf("failed to parse marina modules: %w", err)
	}

	// Start with core module (always enabled)
	var enabledModules []string
	enabledModules = append(enabledModules, "core")

	// Add other enabled modules
	for module, enabled := range marinaModules {
		if enabled {
			enabledModules = append(enabledModules, module)
		}
	}

	return enabledModules, nil
}

// GetUserPermissions returns all permissions for a user in a specific marina
func (s *PermissionService) GetUserPermissions(ctx context.Context, userID, marinaID string) ([][]string, error) {
	return s.adapter.LoadPoliciesForUser(ctx, userID, marinaID)
}

// Helper methods for common permission patterns

// CanRead checks if user can read an object
func (s *PermissionService) CanRead(ctx context.Context, userID, marinaID, object string) (bool, error) {
	return s.CanAccess(ctx, userID, marinaID, object, "read")
}

// CanWrite checks if user can write/update an object
func (s *PermissionService) CanWrite(ctx context.Context, userID, marinaID, object string) (bool, error) {
	return s.CanAccess(ctx, userID, marinaID, object, "write")
}

// CanDelete checks if user can delete an object
func (s *PermissionService) CanDelete(ctx context.Context, userID, marinaID, object string) (bool, error) {
	return s.CanAccess(ctx, userID, marinaID, object, "delete")
}

// CanCreate checks if user can create an object
func (s *PermissionService) CanCreate(ctx context.Context, userID, marinaID, object string) (bool, error) {
	return s.CanAccess(ctx, userID, marinaID, object, "create")
}

// MarinaHasModule checks if a marina has a specific module enabled
func (s *PermissionService) MarinaHasModule(ctx context.Context, marinaID, moduleName string) (bool, error) {
	// Core module is always available for all marinas
	if moduleName == "core" {
		return true, nil
	}

	// Convert marinaID to UUID
	marinaUUID, err := uuid.Parse(marinaID)
	if err != nil {
		return false, fmt.Errorf("invalid marina ID: %w", err)
	}

	// Get marina from database
	marina, err := s.adapter.db.GetMarinaByID(ctx, marinaUUID)
	if err != nil {
		return false, fmt.Errorf("failed to get marina: %w", err)
	}

	// Parse marina modules (boolean format only)
	var marinaModules map[string]bool
	if err := json.Unmarshal(marina.Modules, &marinaModules); err != nil {
		return false, fmt.Errorf("failed to parse marina modules: %w", err)
	}

	// Check if the specific module is enabled
	return marinaModules[moduleName], nil
}

func (s *PermissionService) IsAdmin(ctx context.Context, userID string, marinaID string) (bool, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return false, fmt.Errorf("invalid user ID: %w", err)
	}
	user, err := s.adapter.db.GetUserByID(ctx, userUUID)
	if err != nil {
		return false, fmt.Errorf("failed to get user: %w", err)
	}
	role, err := s.adapter.db.GetRoleByID(ctx, user.RoleID)
	if err != nil {
		return false, fmt.Errorf("failed to get role: %w", err)
	}
	if role.Name == "superuser" {
		return true, nil
	}

	return false, nil
}
