package models

import (
	"encoding/json"
	"fmt"
)

// Permissions represents the role-based access control permissions structure
type Permissions struct {
	// User permissions
	ReadUsers   bool `json:"readUsers" example:"true"`
	WriteUsers  bool `json:"writeUsers" example:"true"`
	DeleteUsers bool `json:"deleteUsers" example:"false"`

	// Organization permissions
	ReadOrganizations   bool `json:"readOrganizations" example:"true"`
	WriteOrganizations  bool `json:"writeOrganizations" example:"true"`
	DeleteOrganizations bool `json:"deleteOrganizations" example:"false"`

	// Marina permissions
	ReadMarinas   bool `json:"readMarinas" example:"true"`
	WriteMarinas  bool `json:"writeMarinas" example:"true"`
	DeleteMarinas bool `json:"deleteMarinas" example:"false"`

	// Role permissions
	ReadRoles   bool `json:"readRoles" example:"true"`
	WriteRoles  bool `json:"writeRoles" example:"true"`
	DeleteRoles bool `json:"deleteRoles" example:"false"`

	// Settings permissions
	ReadSettings  bool `json:"readSettings" example:"true"`
	WriteSettings bool `json:"writeSettings" example:"true"`
}

// ToBytes converts the Permissions struct to a JSON byte array for database storage
func (p *Permissions) ToBytes() ([]byte, error) {
	return json.Marshal(p)
}

// FromBytes populates the Permissions struct from a JSON byte array
func (p *Permissions) FromBytes(data []byte) error {
	if data == nil {
		return nil
	}
	return json.Unmarshal(data, p)
}

// String returns a string representation of the Permissions for debugging
func (p *Permissions) String() string {
	data, err := json.Marshal(p)
	if err != nil {
		return fmt.Sprintf("Error marshaling Permissions: %v", err)
	}
	return string(data)
}

// DefaultAdminPermissions returns a Permissions struct with full administrative access
func DefaultAdminPermissions() *Permissions {
	return &Permissions{
		ReadUsers:           true,
		WriteUsers:          true,
		DeleteUsers:         true,
		ReadOrganizations:   true,
		WriteOrganizations:  true,
		DeleteOrganizations: true,
		ReadMarinas:         true,
		WriteMarinas:        true,
		DeleteMarinas:       true,
		ReadRoles:           true,
		WriteRoles:          true,
		DeleteRoles:         true,
		ReadSettings:        true,
		WriteSettings:       true,
	}
}

// DefaultUserPermissions returns a Permissions struct with limited user-level access
func DefaultUserPermissions() *Permissions {
	return &Permissions{
		ReadUsers:           true,
		WriteUsers:          false,
		DeleteUsers:         false,
		ReadOrganizations:   true,
		WriteOrganizations:  false,
		DeleteOrganizations: false,
		ReadMarinas:         true,
		WriteMarinas:        false,
		DeleteMarinas:       false,
		ReadRoles:           true,
		WriteRoles:          false,
		DeleteRoles:         false,
		ReadSettings:        true,
		WriteSettings:       false,
	}
}
