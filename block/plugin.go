package block

import (
	"fmt"
	"sync"
)

// Plugin represents an extension that registers custom block types.
// Plugins are the primary extension mechanism for Fluui — they allow
// third-party packages to add new Block types to the Registry.
type Plugin interface {
	// Name returns a unique plugin identifier.
	Name() string

	// Init registers the plugin's block types with the given Registry.
	// Called once when the plugin is loaded by PluginManager.Load().
	Init(r *Registry) error
}

// PluginManager manages loaded plugins and their shared Registry.
// It provides a clean lifecycle: create manager → load plugins → use registry.
type PluginManager struct {
	mu       sync.RWMutex
	plugins  []Plugin
	registry *Registry
}

// NewPluginManager creates a PluginManager wrapping the given Registry.
// If reg is nil, a new Registry is created (without default block types).
func NewPluginManager(r *Registry) *PluginManager {
	if r == nil {
		r = NewRegistry()
	}
	return &PluginManager{
		registry: r,
	}
}

// Load registers a single plugin by calling its Init method.
// Returns an error if Init fails or if a plugin with the same Name is already loaded.
func (pm *PluginManager) Load(p Plugin) error {
	if p == nil {
		return fmt.Errorf("plugin is nil")
	}

	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Check for duplicate plugin name.
	for _, existing := range pm.plugins {
		if existing.Name() == p.Name() {
			return fmt.Errorf("plugin %q already loaded", p.Name())
		}
	}

	// Initialize the plugin.
	if err := p.Init(pm.registry); err != nil {
		return fmt.Errorf("plugin %q init failed: %w", p.Name(), err)
	}

	pm.plugins = append(pm.plugins, p)
	return nil
}

// LoadAll loads multiple plugins in order.
// Stops and returns the error on the first failure.
func (pm *PluginManager) LoadAll(plugs []Plugin) error {
	for _, p := range plugs {
		if err := pm.Load(p); err != nil {
			return err
		}
	}
	return nil
}

// Plugins returns a copy of the loaded plugin list.
func (pm *PluginManager) Plugins() []Plugin {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	result := make([]Plugin, len(pm.plugins))
	copy(result, pm.plugins)
	return result
}

// Registry returns the shared Registry used by all plugins.
func (pm *PluginManager) Registry() *Registry {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.registry
}

// HasPlugin returns true if a plugin with the given name is loaded.
func (pm *PluginManager) HasPlugin(name string) bool {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	for _, p := range pm.plugins {
		if p.Name() == name {
			return true
		}
	}
	return false
}

// Count returns the number of loaded plugins.
func (pm *PluginManager) Count() int {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return len(pm.plugins)
}
