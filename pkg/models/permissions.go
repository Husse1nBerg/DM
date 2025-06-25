package models

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Permissions represents the role-based access control permissions using boolean map structure
type Permissions map[string]bool

// ToBytes converts the Permissions map to a JSON byte array for database storage
func (p Permissions) ToBytes() ([]byte, error) {
	return json.Marshal(p)
}

// FromBytes populates the Permissions map from a JSON byte array
func (p *Permissions) FromBytes(data []byte) error {
	if data == nil {
		*p = make(Permissions)
		return nil
	}
	return json.Unmarshal(data, p)
}

// String returns a string representation of the Permissions for debugging
func (p Permissions) String() string {
	data, err := json.Marshal(p)
	if err != nil {
		return fmt.Sprintf("Error marshaling Permissions: %v", err)
	}
	return string(data)
}

// HasPermission checks if a specific object.action permission is granted
func (p Permissions) HasPermission(object, action string) bool {
	key := fmt.Sprintf("%s.%s", object, action)
	granted, exists := p[key]
	return exists && granted
}

// GetObjectPermissions returns all permissions for a specific object
func (p Permissions) GetObjectPermissions(object string) map[string]bool {
	objectPerms := make(map[string]bool)
	prefix := object + "."

	for key, value := range p {
		if strings.HasPrefix(key, prefix) {
			action := strings.TrimPrefix(key, prefix)
			objectPerms[action] = value
		}
	}
	return objectPerms
}

// GetAllowedActions returns all allowed actions for a specific object
func (p Permissions) GetAllowedActions(object string) []string {
	var allowed []string
	prefix := object + "."

	for key, value := range p {
		if strings.HasPrefix(key, prefix) && value {
			action := strings.TrimPrefix(key, prefix)
			allowed = append(allowed, action)
		}
	}
	return allowed
}

// GetAllowedObjects returns all objects that have at least one allowed action
func (p Permissions) GetAllowedObjects() []string {
	objectSet := make(map[string]bool)

	for key, value := range p {
		if value && strings.Contains(key, ".") {
			parts := strings.SplitN(key, ".", 2)
			if len(parts) == 2 {
				objectSet[parts[0]] = true
			}
		}
	}

	var objects []string
	for object := range objectSet {
		objects = append(objects, object)
	}
	return objects
}

// Grant adds or updates a permission
func (p Permissions) Grant(object, action string) {
	key := fmt.Sprintf("%s.%s", object, action)
	p[key] = true
}

// Revoke removes a permission or sets it to false
func (p Permissions) Revoke(object, action string) {
	key := fmt.Sprintf("%s.%s", object, action)
	p[key] = false
}

// DefaultAdminPermissions returns a Permissions map with full administrative access
func DefaultAdminPermissions() Permissions {
	return Permissions{
		"users.read":             true,
		"users.write":            true,
		"users.delete":           true,
		"organizations.read":     true,
		"organizations.write":    true,
		"organizations.delete":   true,
		"marinas.read":           true,
		"marinas.write":          true,
		"marinas.delete":         true,
		"customers.read":         true,
		"customers.write":        true,
		"customers.delete":       true,
		"vessels.read":           true,
		"vessels.write":          true,
		"vessels.delete":         true,
		"roles.read":             true,
		"roles.write":            true,
		"roles.delete":           true,
		"settings.read":          true,
		"settings.write":         true,
		"esign_templates.read":   true,
		"esign_templates.create": true,
		"esign_templates.write":  true,
		"esign_templates.delete": true,
		"esign_documents.read":   true,
		"esign_documents.create": true,
		"esign_documents.write":  true,
		"esign_documents.delete": true,
	}
}

// DefaultUserPermissions returns a Permissions map with limited user-level access
func DefaultUserPermissions() Permissions {
	return Permissions{
		"users.read":           true,
		"users.write":          false,
		"users.delete":         false,
		"organizations.read":   true,
		"organizations.write":  false,
		"organizations.delete": false,
		"marinas.read":         true,
		"marinas.write":        false,
		"marinas.delete":       false,
		"customers.read":       true,
		"customers.write":      false,
		"customers.delete":     false,
		"vessels.read":         true,
		"vessels.write":        false,
		"vessels.delete":       false,
		"roles.read":           true,
		"roles.write":          false,
		"roles.delete":         false,
		"settings.read":        true,
		"settings.write":       false,
	}
}

// ReadOnlyPermissions returns a Permissions map with only read access
func ReadOnlyPermissions() Permissions {
	return Permissions{
		"users.read":         true,
		"organizations.read": true,
		"marinas.read":       true,
		"customers.read":     true,
		"vessels.read":       true,
		"roles.read":         true,
		"settings.read":      true,
	}
}
