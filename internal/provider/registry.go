package provider

import (
	"fmt"
	"sync"
)

// Factory is a function that creates a new provider instance
type Factory func() Provider

// Registry manages provider registration and creation
type Registry struct {
	mu        sync.RWMutex
	factories map[string]Factory
}

// globalRegistry is the default registry
var globalRegistry = NewRegistry()

// NewRegistry creates a new provider registry
func NewRegistry() *Registry {
	return &Registry{
		factories: make(map[string]Factory),
	}
}

// Register registers a provider factory with the registry
func (r *Registry) Register(name string, factory Factory) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.factories[name] = factory
}

// Get creates a new provider instance by name
func (r *Registry) Get(name string) (Provider, error) {
	r.mu.RLock()
	factory, ok := r.factories[name]
	r.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("unknown provider: %s", name)
	}

	return factory(), nil
}

// List returns all registered provider names
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.factories))
	for name := range r.factories {
		names = append(names, name)
	}
	return names
}

// Has returns true if a provider is registered
func (r *Registry) Has(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.factories[name]
	return ok
}

// Register registers a provider factory with the global registry
func Register(name string, factory Factory) {
	globalRegistry.Register(name, factory)
}

// Get creates a new provider instance from the global registry
func Get(name string) (Provider, error) {
	return globalRegistry.Get(name)
}

// List returns all registered provider names from the global registry
func List() []string {
	return globalRegistry.List()
}

// Has returns true if a provider is registered in the global registry
func Has(name string) bool {
	return globalRegistry.Has(name)
}

// GetAndInitialize gets a provider and initializes it with the given config
func GetAndInitialize(name string, cfg Config) (Provider, error) {
	p, err := Get(name)
	if err != nil {
		return nil, err
	}

	if err := p.Initialize(cfg); err != nil {
		return nil, fmt.Errorf("failed to initialize provider %s: %w", name, err)
	}

	return p, nil
}

// init registers all built-in providers
func init() {
	Register("openai", func() Provider { return &OpenAIProvider{} })
	Register("gemini", func() Provider { return &GeminiProvider{} })
	Register("deepseek", func() Provider { return &DeepseekProvider{} })
	Register("kimi", func() Provider { return &KimiProvider{} })
	Register("glm", func() Provider { return &GLMProvider{} })
}
