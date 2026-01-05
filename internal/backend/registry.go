package backend

import (
	"context"
	"fmt"
	"sync"
)

// Registry manages available backends
type Registry struct {
	mu       sync.RWMutex
	backends map[string]Backend
	default_ string
}

// NewRegistry creates a new backend registry
func NewRegistry() *Registry {
	return &Registry{
		backends: make(map[string]Backend),
	}
}

// Register adds a backend to the registry
func (r *Registry) Register(backend Backend) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := backend.Name()
	if _, exists := r.backends[name]; exists {
		return fmt.Errorf("backend already registered: %s", name)
	}

	r.backends[name] = backend
	return nil
}

// Get retrieves a backend by name
func (r *Registry) Get(name string) (Backend, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	backend, ok := r.backends[name]
	return backend, ok
}

// SetDefault sets the default backend
func (r *Registry) SetDefault(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.backends[name]; !exists {
		return fmt.Errorf("backend not found: %s", name)
	}

	r.default_ = name
	return nil
}

// Default returns the default backend
func (r *Registry) Default() (Backend, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.default_ == "" {
		// Return the first available backend
		for _, b := range r.backends {
			return b, nil
		}
		return nil, fmt.Errorf("no backends registered")
	}

	backend, ok := r.backends[r.default_]
	if !ok {
		return nil, fmt.Errorf("default backend not found: %s", r.default_)
	}

	return backend, nil
}

// List returns all registered backends
func (r *Registry) List() []Backend {
	r.mu.RLock()
	defer r.mu.RUnlock()

	backends := make([]Backend, 0, len(r.backends))
	for _, b := range r.backends {
		backends = append(backends, b)
	}
	return backends
}

// GetAvailable returns the first available backend
func (r *Registry) GetAvailable(ctx context.Context) (Backend, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Try default first
	if r.default_ != "" {
		if b, ok := r.backends[r.default_]; ok {
			if avail, _ := b.IsAvailable(ctx); avail {
				return b, nil
			}
		}
	}

	// Try others
	for _, b := range r.backends {
		if avail, _ := b.IsAvailable(ctx); avail {
			return b, nil
		}
	}

	return nil, fmt.Errorf("no available backends")
}
