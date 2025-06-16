package guard

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/casbin/casbin/v2/model"
	"github.com/dockworks/dm-web-backend/internal/db"
	"github.com/google/uuid"
)

// MarinaCasbinAdapter implements the persist.Adapter interface for Casbin
// It reads policies from existing database tables without storing duplicate data
type MarinaCasbinAdapter struct {
	db *db.Queries
}

// NewMarinaCasbinAdapter creates a new adapter instance
func NewMarinaCasbinAdapter(database *db.Queries) *MarinaCasbinAdapter {
	return &MarinaCasbinAdapter{
		db: database,
	}
}

// LoadPolicy loads policies from the database for the specified model
// This is called by Casbin to load all policies
func (a *MarinaCasbinAdapter) LoadPolicy(model model.Model) error {
	// For our per-request approach, we don't need to load all policies
	// This method is required by the interface but we'll handle policy loading
	// in the LoadPoliciesForUser method instead
	return nil
}

// LoadPoliciesForUser loads policies for a specific user in a specific marina
// This is our custom method for per-request policy building
func (a *MarinaCasbinAdapter) LoadPoliciesForUser(ctx context.Context, userID, marinaID string) ([][]string, error) {
	var policies [][]string

	// Convert string IDs to UUIDs
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	marinaUUID, err := uuid.Parse(marinaID)
	if err != nil {
		return nil, fmt.Errorf("invalid marina ID: %w", err)
	}

	// Get user's role in the specified marina
	userRoleID, err := a.db.GetUserRoleInMarina(ctx, db.GetUserRoleInMarinaParams{
		ID:       userUUID,
		MarinaID: marinaUUID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get user role in marina: %w", err)
	}

	// Get role permissions
	role, err := a.db.GetRoleByID(ctx, userRoleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	// Parse role permissions from JSON (boolean format only)
	var rolePermissions map[string]bool
	if err := json.Unmarshal(role.Permissions, &rolePermissions); err != nil {
		return nil, fmt.Errorf("failed to parse role permissions: %w", err)
	}

	// Get marina enabled modules
	marina, err := a.db.GetMarinaByID(ctx, marinaUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get marina: %w", err)
	}

	// Parse marina modules (boolean format only)
	var marinaModules map[string]bool
	if err := json.Unmarshal(marina.Modules, &marinaModules); err != nil {
		return nil, fmt.Errorf("failed to parse marina modules: %w", err)
	}

	// Build policies from boolean permissions
	for permissionKey, isAllowed := range rolePermissions {
		if isAllowed {
			// Parse "object.action" format
			parts := strings.Split(permissionKey, ".")
			if len(parts) == 2 {
				object, action := parts[0], parts[1]

				// Check if the object belongs to an enabled module
				if objectModule, exists := GetObjectModule(object); exists {
					if marinaModules[objectModule] || objectModule == "core" {
						// Create policy: userID, marinaID, object, action
						policy := []string{userID, marinaID, object, action}
						policies = append(policies, policy)
					}
				}
			}
		}
	}

	// Note: We don't need role-based policies since we're creating direct user policies above
	// The role information is already incorporated into the direct policies

	return policies, nil
}

// SavePolicy saves policies to the database
// Since we use existing tables as source of truth, this is not implemented
func (a *MarinaCasbinAdapter) SavePolicy(model model.Model) error {
	return fmt.Errorf("SavePolicy not implemented - use existing database tables")
}

// AddPolicy adds a policy rule to the database
// Since we use existing tables as source of truth, this is not implemented
func (a *MarinaCasbinAdapter) AddPolicy(sec string, ptype string, rule []string) error {
	return fmt.Errorf("AddPolicy not implemented - use existing database tables")
}

// RemovePolicy removes a policy rule from the database
// Since we use existing tables as source of truth, this is not implemented
func (a *MarinaCasbinAdapter) RemovePolicy(sec string, ptype string, rule []string) error {
	return fmt.Errorf("RemovePolicy not implemented - use existing database tables")
}

// RemoveFilteredPolicy removes policy rules that match the filter
// Since we use existing tables as source of truth, this is not implemented
func (a *MarinaCasbinAdapter) RemoveFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) error {
	return fmt.Errorf("RemoveFilteredPolicy not implemented - use existing database tables")
}
