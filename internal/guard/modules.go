package guard

// ModuleObjects defines which objects (resources) belong to each marina module.
// This configuration determines what permissions are needed based on enabled marina modules.
type ModuleConfig struct {
	Objects []string `json:"objects"`
}

// ModuleObjects maps marina module names to their contained objects/resources
// These names match the JSON fields in pkg/models/modules.go
var ModuleObjects = map[string]ModuleConfig{
	// Core module - always available for all users (not stored in marina modules)
	"core": {
		Objects: []string{"profile", "users", "organizations", "marinas", "addresses", "roles", "marina_gallery"},
	},
	"customerVessels": {
		Objects: []string{"customers", "vessels", "messages", "documents", "boat_gallery"},
	},
	"serviceManagement": {
		Objects: []string{"work_orders"},
	},
	"payments": {
		Objects: []string{},
	},
	"inventoryManagement": {
		Objects: []string{},
	},
	"marinaManagement": {
		Objects: []string{},
	},
	"pos": {
		Objects: []string{},
	},
	"salesManagement": {
		Objects: []string{},
	},
}

// GetObjectModule returns the module name that contains the given object
func GetObjectModule(object string) (string, bool) {
	for moduleName, config := range ModuleObjects {
		for _, obj := range config.Objects {
			if obj == object {
				return moduleName, true
			}
		}
	}
	return "", false
}

// GetModuleObjects returns all objects for a given module
func GetModuleObjects(module string) ([]string, bool) {
	config, exists := ModuleObjects[module]
	if !exists {
		return nil, false
	}
	return config.Objects, true
}

// IsValidObject checks if an object is defined in any module
func IsValidObject(object string) bool {
	_, exists := GetObjectModule(object)
	return exists
}
