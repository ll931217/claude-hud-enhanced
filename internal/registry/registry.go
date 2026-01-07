package registry

import (
	"fmt"
	"sync"
)

// SectionFactory is a function that creates a Section instance from configuration
type SectionFactory func(config interface{}) (Section, error)

// SectionRegistry manages registration and creation of section types
type SectionRegistry struct {
	mu       sync.RWMutex
	factories map[string]SectionFactory
}

// global registry instance
var defaultRegistry = &SectionRegistry{
	factories: make(map[string]SectionFactory),
}

// Register registers a new section type with the given name and factory function
func (r *SectionRegistry) Register(name string, factory SectionFactory) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if factory == nil {
		panic(fmt.Sprintf("cannot register nil factory for section: %s", name))
	}

	r.factories[name] = factory
}

// Create creates a new section instance of the specified type with the given configuration
func (r *SectionRegistry) Create(name string, config interface{}) (Section, error) {
	r.mu.RLock()
	factory, exists := r.factories[name]
	r.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("section type not registered: %s", name)
	}

	return factory(config)
}

// List returns a list of all registered section type names
func (r *SectionRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.factories))
	for name := range r.factories {
		names = append(names, name)
	}

	return names
}

// Register registers a section type with the default registry
func Register(name string, factory SectionFactory) {
	defaultRegistry.Register(name, factory)
}

// Create creates a section instance using the default registry
func Create(name string, config interface{}) (Section, error) {
	return defaultRegistry.Create(name, config)
}

// List returns all registered section types from the default registry
func List() []string {
	return defaultRegistry.List()
}

// DefaultRegistry returns the default registry instance
func DefaultRegistry() *SectionRegistry {
	return defaultRegistry
}
