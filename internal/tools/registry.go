package tools

import (
	"fmt"
	"sync"

	"github.com/tmc/langchaingo/llms"
)

// Registry manages tool registration and lookup
type Registry struct {
	mu    sync.RWMutex
	tools map[string]Tool
}

// NewRegistry creates a new tool registry
func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]Tool),
	}
}

// Register adds a tool to the registry
func (r *Registry) Register(tool Tool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tools[tool.Name()] = tool
}

// Get retrieves a tool by name
func (r *Registry) Get(name string) (Tool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tool, ok := r.tools[name]
	if !ok {
		return nil, fmt.Errorf("unknown tool: %s", name)
	}
	return tool, nil
}

// Has returns true if a tool is registered
func (r *Registry) Has(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.tools[name]
	return ok
}

// All returns all registered tools
func (r *Registry) All() []Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tools := make([]Tool, 0, len(r.tools))
	for _, tool := range r.tools {
		tools = append(tools, tool)
	}
	return tools
}

// Names returns all registered tool names
func (r *Registry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.tools))
	for name := range r.tools {
		names = append(names, name)
	}
	return names
}

// ToLangchainTools converts all registered tools to langchaingo format
func (r *Registry) ToLangchainTools() []llms.Tool {
	tools := r.All()
	return ToLangchainTools(tools)
}

// FilterByNames returns only tools with the given names
func (r *Registry) FilterByNames(names []string) []Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	nameSet := make(map[string]bool)
	for _, name := range names {
		nameSet[name] = true
	}

	var tools []Tool
	for name, tool := range r.tools {
		if nameSet[name] {
			tools = append(tools, tool)
		}
	}
	return tools
}

// DefaultRegistry creates a registry with all default tools
func DefaultRegistry() *Registry {
	r := NewRegistry()
	r.Register(NewBashTool())
	r.Register(NewFileReadTool())
	r.Register(NewFileWriteTool())
	return r
}
