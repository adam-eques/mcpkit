package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"sync"

	"github.com/adam-eques/mcpkit/mcp"
)

// Registry is a concurrency-safe collection of tools keyed by name.
type Registry struct {
	mu     sync.RWMutex
	byName map[string]Handler
}

// NewRegistry returns an empty registry.
func NewRegistry() *Registry {
	return &Registry{byName: make(map[string]Handler)}
}

// Register adds h to the registry. It fails if the tool's name is empty or
// already registered.
func (r *Registry) Register(h Handler) error {
	name := h.Definition().Name
	if name == "" {
		return fmt.Errorf("tools: tool has an empty name")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.byName[name]; exists {
		return fmt.Errorf("tools: tool %q already registered", name)
	}
	r.byName[name] = h
	return nil
}

// MustRegister is like Register but panics on error, for use at startup.
func (r *Registry) MustRegister(handlers ...Handler) {
	for _, h := range handlers {
		if err := r.Register(h); err != nil {
			panic(err)
		}
	}
}

// Get returns the named tool.
func (r *Registry) Get(name string) (Handler, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	h, ok := r.byName[name]
	return h, ok
}

// Len returns the number of registered tools.
func (r *Registry) Len() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.byName)
}

// List returns every tool definition sorted by name for deterministic output.
func (r *Registry) List() []mcp.Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]mcp.Tool, 0, len(r.byName))
	for _, h := range r.byName {
		out = append(out, h.Definition())
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// Call dispatches to the named tool. It returns ErrToolNotFound when the tool is
// unknown so the server can map it to an appropriate protocol error.
func (r *Registry) Call(ctx context.Context, name string, args json.RawMessage) (*mcp.CallToolResult, error) {
	h, ok := r.Get(name)
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrToolNotFound, name)
	}
	return h.Call(ctx, args)
}
