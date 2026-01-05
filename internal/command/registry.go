package command

import (
	"fmt"
	"sort"
	"sync"
)

// Registry manages all available commands
type Registry struct {
	mu       sync.RWMutex
	commands map[string]Command
	aliases  map[string]string
}

// NewRegistry creates a new command registry
func NewRegistry() *Registry {
	return &Registry{
		commands: make(map[string]Command),
		aliases:  make(map[string]string),
	}
}

// Register adds a command to the registry
func (r *Registry) Register(cmd Command) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := cmd.Name()
	if _, exists := r.commands[name]; exists {
		return fmt.Errorf("command already registered: %s", name)
	}

	r.commands[name] = cmd

	for _, alias := range cmd.Aliases() {
		if existing, exists := r.aliases[alias]; exists {
			return fmt.Errorf("alias %s already used by %s", alias, existing)
		}
		r.aliases[alias] = name
	}

	return nil
}

// Get retrieves a command by name or alias
func (r *Registry) Get(nameOrAlias string) (Command, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if cmd, ok := r.commands[nameOrAlias]; ok {
		return cmd, true
	}

	if name, ok := r.aliases[nameOrAlias]; ok {
		return r.commands[name], true
	}

	return nil, false
}

// List returns all registered commands
func (r *Registry) List() []Command {
	r.mu.RLock()
	defer r.mu.RUnlock()

	cmds := make([]Command, 0, len(r.commands))
	for _, cmd := range r.commands {
		cmds = append(cmds, cmd)
	}

	// Sort by name
	sort.Slice(cmds, func(i, j int) bool {
		return cmds[i].Name() < cmds[j].Name()
	})

	return cmds
}

// ListByCategory returns commands filtered by category
func (r *Registry) ListByCategory(cat Category) []Command {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var cmds []Command
	for _, cmd := range r.commands {
		if cmd.Category() == cat {
			cmds = append(cmds, cmd)
		}
	}

	// Sort by name
	sort.Slice(cmds, func(i, j int) bool {
		return cmds[i].Name() < cmds[j].Name()
	})

	return cmds
}

// Names returns all command names
func (r *Registry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.commands))
	for name := range r.commands {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Count returns the number of registered commands
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.commands)
}
