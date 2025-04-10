package models

import (
	"encoding/json"
	"fmt"
)

// Modules represents the feature access control modules for users
type Modules struct {
	CustomerVesselsRead      bool `json:"customerVesselsRead"`
	CustomerVesselsWrite     bool `json:"customerVesselsWrite"`
	PaymentsRead             bool `json:"paymentsRead"`
	PaymentsWrite            bool `json:"paymentsWrite"`
	ServiceManagementRead    bool `json:"serviceManagementRead"`
	ServiceManagementWrite   bool `json:"serviceManagementWrite"`
	InventoryManagementRead  bool `json:"inventoryManagementRead"`
	InventoryManagementWrite bool `json:"inventoryManagementWrite"`
	MarinaManagementRead     bool `json:"marinaManagementRead"`
	MarinaManagementWrite    bool `json:"marinaManagementWrite"`
	POSRead                  bool `json:"posRead"`
	POSWrite                 bool `json:"posWrite"`
	SalesManagementRead      bool `json:"salesManagementRead"`
	SalesManagementWrite     bool `json:"salesManagementWrite"`
}

// ToBytes converts the Modules struct to a JSON byte array for database storage
func (m *Modules) ToBytes() ([]byte, error) {
	return json.Marshal(m)
}

// FromBytes populates the Modules struct from a JSON byte array
func (m *Modules) FromBytes(data []byte) error {
	if data == nil {
		return nil
	}
	return json.Unmarshal(data, m)
}

// String returns a string representation of the Modules for debugging
func (m *Modules) String() string {
	data, err := json.Marshal(m)
	if err != nil {
		return fmt.Sprintf("Error marshaling Modules: %v", err)
	}
	return string(data)
}

// DefaultModules returns a Modules struct with all access enabled
func DefaultModules() *Modules {
	return &Modules{
		CustomerVesselsRead:      true,
		CustomerVesselsWrite:     true,
		PaymentsRead:             true,
		PaymentsWrite:            true,
		ServiceManagementRead:    true,
		ServiceManagementWrite:   true,
		InventoryManagementRead:  true,
		InventoryManagementWrite: true,
		MarinaManagementRead:     true,
		MarinaManagementWrite:    true,
		POSRead:                  true,
		POSWrite:                 true,
		SalesManagementRead:      true,
		SalesManagementWrite:     true,
	}
}

// ReadOnlyModules returns a Modules struct with only read access enabled
func ReadOnlyModules() *Modules {
	return &Modules{
		CustomerVesselsRead:      true,
		CustomerVesselsWrite:     false,
		PaymentsRead:             true,
		PaymentsWrite:            false,
		ServiceManagementRead:    true,
		ServiceManagementWrite:   false,
		InventoryManagementRead:  true,
		InventoryManagementWrite: false,
		MarinaManagementRead:     true,
		MarinaManagementWrite:    false,
		POSRead:                  true,
		POSWrite:                 false,
		SalesManagementRead:      true,
		SalesManagementWrite:     false,
	}
}
