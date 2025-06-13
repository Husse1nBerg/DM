package guard

import (
	"encoding/json"
)

// BooleanModules represents the boolean-based module structure
type BooleanModules map[string]bool

// ModuleBuilder helps build boolean marina modules easily
type ModuleBuilder struct {
	modules BooleanModules
}

// NewModuleBuilder creates a new module builder
func NewModuleBuilder() *ModuleBuilder {
	return &ModuleBuilder{
		modules: make(BooleanModules),
	}
}

// Enable enables a specific module
func (mb *ModuleBuilder) Enable(module string) *ModuleBuilder {
	mb.modules[module] = true
	return mb
}

// Disable explicitly disables a specific module
func (mb *ModuleBuilder) Disable(module string) *ModuleBuilder {
	mb.modules[module] = false
	return mb
}

// EnableAll enables all defined modules
func (mb *ModuleBuilder) EnableAll() *ModuleBuilder {
	for module := range ModuleObjects {
		mb.modules[module] = true
	}
	return mb
}

// EnableCore enables core marina modules (customers and payments)
func (mb *ModuleBuilder) EnableCore() *ModuleBuilder {
	return mb.Enable("customerVessels").
		Enable("payments")
}

// EnableBasic enables basic marina modules (core + marina management)
func (mb *ModuleBuilder) EnableBasic() *ModuleBuilder {
	return mb.EnableCore().
		Enable("marinaManagement")
}

// EnableStandard enables standard marina modules (basic + services)
func (mb *ModuleBuilder) EnableStandard() *ModuleBuilder {
	return mb.EnableBasic().
		Enable("serviceManagement")
}

// EnableFull enables all modules except POS (for standard marinas)
func (mb *ModuleBuilder) EnableFull() *ModuleBuilder {
	return mb.EnableStandard().
		Enable("inventoryManagement").
		Enable("salesManagement")
}

// EnableMultiple enables multiple modules at once
func (mb *ModuleBuilder) EnableMultiple(modules ...string) *ModuleBuilder {
	for _, module := range modules {
		mb.Enable(module)
	}
	return mb
}

// Build returns the final modules map
func (mb *ModuleBuilder) Build() BooleanModules {
	return mb.modules
}

// ToJSON converts modules to JSON for database storage
func (mb *ModuleBuilder) ToJSON() ([]byte, error) {
	return json.Marshal(mb.modules)
}

// IsEnabled checks if a module is enabled
func (bm BooleanModules) IsEnabled(module string) bool {
	return bm[module]
}

// GetEnabledModules returns a list of all enabled modules
func (bm BooleanModules) GetEnabledModules() []string {
	var enabled []string
	for module, isEnabled := range bm {
		if isEnabled {
			enabled = append(enabled, module)
		}
	}
	return enabled
}

// GetAllAvailableObjects returns all objects available based on enabled modules
func (bm BooleanModules) GetAllAvailableObjects() []string {
	var objects []string
	objectSet := make(map[string]bool) // To avoid duplicates

	for module, isEnabled := range bm {
		if isEnabled {
			if moduleConfig, exists := ModuleObjects[module]; exists {
				for _, object := range moduleConfig.Objects {
					if !objectSet[object] {
						objects = append(objects, object)
						objectSet[object] = true
					}
				}
			}
		}
	}

	return objects
}

// HasObjectAccess checks if an object can be accessed based on enabled modules
func (bm BooleanModules) HasObjectAccess(object string) bool {
	// Find which module contains this object
	moduleName, exists := GetObjectModule(object)
	if !exists {
		return false
	}

	// Check if that module is enabled
	return bm.IsEnabled(moduleName)
}

// Validate checks if all modules in the configuration are valid (defined in the system)
func (bm BooleanModules) Validate() []string {
	var invalidModules []string
	for module := range bm {
		if _, exists := ModuleObjects[module]; !exists {
			invalidModules = append(invalidModules, module)
		}
	}
	return invalidModules
}
