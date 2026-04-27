package core

import (
	"fmt"
	"sort"
	"sync"
)

var (
	registry   = make(map[string]Plugin)
	registryMu sync.RWMutex
)

// Register adds a plugin to the registry. Should be called from init()
// in plugin packages. Panics if a plugin with the same name is already registered.
func Register(p Plugin) {
	registryMu.Lock()
	defer registryMu.Unlock()
	if _, exists := registry[p.Name()]; exists {
		panic(fmt.Sprintf("ctx: plugin %q already registered", p.Name()))
	}
	registry[p.Name()] = p
}

// All returns all registered plugins, sorted by name.
func All() []Plugin {
	registryMu.RLock()
	defer registryMu.RUnlock()
	plugins := make([]Plugin, 0, len(registry))
	for _, p := range registry {
		plugins = append(plugins, p)
	}
	sort.Slice(plugins, func(i, j int) bool {
		return plugins[i].Name() < plugins[j].Name()
	})
	return plugins
}

// Get returns a plugin by name, or nil if not found.
func Get(name string) Plugin {
	registryMu.RLock()
	defer registryMu.RUnlock()
	return registry[name]
}
