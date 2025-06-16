package models

import (
	"encoding/json"
	"fmt"
)

// Modules represents the feature access control modules for users
type Modules struct {
	CustomerVessels     bool `json:"customerVessels"`
	Payments            bool `json:"payments"`
	ServiceManagement   bool `json:"serviceManagement"`
	InventoryManagement bool `json:"inventoryManagement"`
	MarinaManagement    bool `json:"marinaManagement"`
	POS                 bool `json:"pos"`
	SalesManagement     bool `json:"salesManagement"`
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
		CustomerVessels:     true,
		ServiceManagement:   false,
		Payments:            false,
		InventoryManagement: false,
		MarinaManagement:    false,
		POS:                 false,
		SalesManagement:     false,
	}
}
