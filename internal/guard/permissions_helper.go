package guard

import (
	"encoding/json"
	"fmt"
)

// BooleanPermissions represents the new efficient boolean-based permission structure
type BooleanPermissions map[string]bool

// PermissionBuilder helps build boolean permissions easily
type PermissionBuilder struct {
	permissions BooleanPermissions
}

// NewPermissionBuilder creates a new permission builder
func NewPermissionBuilder() *PermissionBuilder {
	return &PermissionBuilder{
		permissions: make(BooleanPermissions),
	}
}

// Allow grants permission for an object and action
func (pb *PermissionBuilder) Allow(object, action string) *PermissionBuilder {
	key := fmt.Sprintf("%s.%s", object, action)
	pb.permissions[key] = true
	return pb
}

// Deny explicitly denies permission for an object and action
func (pb *PermissionBuilder) Deny(object, action string) *PermissionBuilder {
	key := fmt.Sprintf("%s.%s", object, action)
	pb.permissions[key] = false
	return pb
}

// AllowAll grants all CRUD permissions for an object
func (pb *PermissionBuilder) AllowAll(object string) *PermissionBuilder {
	return pb.Allow(object, "create").
		Allow(object, "read").
		Allow(object, "write").
		Allow(object, "delete")
}

// AllowRead grants only read permission for an object
func (pb *PermissionBuilder) AllowRead(object string) *PermissionBuilder {
	return pb.Allow(object, "read")
}

// AllowReadWrite grants read and write permissions for an object
func (pb *PermissionBuilder) AllowReadWrite(object string) *PermissionBuilder {
	return pb.Allow(object, "read").
		Allow(object, "write")
}

// AllowModule grants specific permissions for all objects in a module
func (pb *PermissionBuilder) AllowModule(module string, actions ...string) *PermissionBuilder {
	if moduleConfig, exists := ModuleObjects[module]; exists {
		for _, object := range moduleConfig.Objects {
			for _, action := range actions {
				pb.Allow(object, action)
			}
		}
	}
	return pb
}

// AllowModuleAll grants all CRUD permissions for all objects in a module
func (pb *PermissionBuilder) AllowModuleAll(module string) *PermissionBuilder {
	return pb.AllowModule(module, "create", "read", "write", "delete")
}

// AllowModuleRead grants read-only access for all objects in a module
func (pb *PermissionBuilder) AllowModuleRead(module string) *PermissionBuilder {
	return pb.AllowModule(module, "read")
}

// AllowModuleReadWrite grants read and write access for all objects in a module
func (pb *PermissionBuilder) AllowModuleReadWrite(module string) *PermissionBuilder {
	return pb.AllowModule(module, "read", "write")
}

// Build returns the final permissions map
func (pb *PermissionBuilder) Build() BooleanPermissions {
	return pb.permissions
}

// ToJSON converts permissions to JSON for database storage
func (pb *PermissionBuilder) ToJSON() ([]byte, error) {
	return json.Marshal(pb.permissions)
}

// FromJSON loads permissions from JSON
func FromJSON(jsonData []byte) (BooleanPermissions, error) {
	var permissions BooleanPermissions
	err := json.Unmarshal(jsonData, &permissions)
	return permissions, err
}

// HasPermission checks if a permission exists and is true
func (bp BooleanPermissions) HasPermission(object, action string) bool {
	key := fmt.Sprintf("%s.%s", object, action)
	return bp[key]
}

// GetObjectPermissions returns all permissions for a specific object
func (bp BooleanPermissions) GetObjectPermissions(object string) map[string]bool {
	result := make(map[string]bool)
	prefix := object + "."

	for key, value := range bp {
		if len(key) > len(prefix) && key[:len(prefix)] == prefix {
			action := key[len(prefix):]
			result[action] = value
		}
	}

	return result
}

// GetAllowedActions returns all actions that are allowed for an object
func (bp BooleanPermissions) GetAllowedActions(object string) []string {
	var actions []string
	objectPerms := bp.GetObjectPermissions(object)

	for action, allowed := range objectPerms {
		if allowed {
			actions = append(actions, action)
		}
	}

	return actions
}

// GetAllowedObjects returns all objects that have at least one allowed action
func (bp BooleanPermissions) GetAllowedObjects() []string {
	objects := make(map[string]bool)

	for key, allowed := range bp {
		if allowed {
			// Find the dot separator and extract object name
			for i, char := range key {
				if char == '.' {
					object := key[:i]
					objects[object] = true
					break
				}
			}
		}
	}

	var result []string
	for object := range objects {
		result = append(result, object)
	}

	return result
}
