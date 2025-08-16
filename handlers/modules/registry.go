package modules

// Module represents a reusable UI module
type Module interface {
	GetTemplateName() string
}

// ModuleRegistry manages available modules
type ModuleRegistry struct {
	modules map[string]Module
}

var registry *ModuleRegistry

// GetRegistry returns the singleton module registry
func GetRegistry() *ModuleRegistry {
	if registry == nil {
		registry = &ModuleRegistry{
			modules: make(map[string]Module),
		}
	}
	return registry
}

// Register adds a module to the registry
func (mr *ModuleRegistry) Register(name string, module Module) {
	mr.modules[name] = module
}

// Get retrieves a module by name
func (mr *ModuleRegistry) Get(name string) (Module, bool) {
	module, exists := mr.modules[name]
	return module, exists
}