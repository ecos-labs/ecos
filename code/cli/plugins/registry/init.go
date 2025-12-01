package registry

import (
	"fmt"

	"github.com/ecos-labs/ecos/code/cli/plugins/types"
)

// InitPluginRegistry holds factory functions for init plugins
var InitPluginRegistry = make(map[string]types.PluginFactory)

// RegisterInitPlugin allows plugins to self-register
func RegisterInitPlugin(name string, factory types.PluginFactory) {
	InitPluginRegistry[name] = factory
}

// LoadInitPlugin loads an init plugin instance from the registry
func LoadInitPlugin(dataSource string, force bool, outputPath string) (types.InitPlugin, error) {
	factory, ok := InitPluginRegistry[dataSource]
	if !ok {
		return nil, fmt.Errorf("unsupported data source: %s", dataSource)
	}

	// Create plugin with empty config - plugin will fill it during RunInteractiveSetup
	return factory(force, outputPath)
}

// DestroyPluginFactory is a function type that creates a new DestroyPlugin instance.
type DestroyPluginFactory func() types.DestroyPlugin

var destroyPluginFactories = map[string]DestroyPluginFactory{}

// RegisterDestroyPlugin registers a destroy plugin factory with the given name.
func RegisterDestroyPlugin(name string, factory DestroyPluginFactory) {
	destroyPluginFactories[name] = factory
}

// LoadDestroyPlugin loads a destroy plugin by name from the registry.
func LoadDestroyPlugin(name string) (types.DestroyPlugin, error) {
	factory, ok := destroyPluginFactories[name]
	if !ok {
		return nil, fmt.Errorf("destroy plugin '%s' not found", name)
	}
	return factory(), nil
}
