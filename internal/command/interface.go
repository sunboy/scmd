// Package command provides the command system for scmd
package command

import (
	"context"

	"github.com/scmd/scmd/internal/backend"
	"github.com/scmd/scmd/internal/config"
)

// Command defines the interface for all scmd commands
type Command interface {
	// Metadata
	Name() string
	Aliases() []string
	Description() string
	Usage() string
	Examples() []string
	Category() Category

	// Execution
	Execute(ctx context.Context, args *Args, execCtx *ExecContext) (*Result, error)

	// Validation
	Validate(args *Args) error

	// Requirements
	RequiresBackend() bool
}

// Category classifies commands
type Category string

const (
	CategoryCore   Category = "core"
	CategoryCode   Category = "code"
	CategoryGit    Category = "git"
	CategoryConfig Category = "config"
	CategoryPlugin Category = "plugin"
)

// Args represents parsed command arguments
type Args struct {
	Positional []string
	Flags      map[string]bool
	Options    map[string]string
	Raw        string
}

// NewArgs creates a new Args instance
func NewArgs() *Args {
	return &Args{
		Positional: []string{},
		Flags:      make(map[string]bool),
		Options:    make(map[string]string),
	}
}

// HasFlag returns true if a flag is set
func (a *Args) HasFlag(name string) bool {
	return a.Flags[name]
}

// GetOption returns an option value
func (a *Args) GetOption(name string) string {
	return a.Options[name]
}

// GetOptionOrDefault returns an option value or a default
func (a *Args) GetOptionOrDefault(name, def string) string {
	if v, ok := a.Options[name]; ok {
		return v
	}
	return def
}

// Result represents command execution result
type Result struct {
	Success     bool
	Output      string
	Error       string
	Suggestions []string
	ExitCode    int
}

// NewResult creates a successful result
func NewResult(output string) *Result {
	return &Result{
		Success:  true,
		Output:   output,
		ExitCode: 0,
	}
}

// NewErrorResult creates an error result
func NewErrorResult(err string, suggestions ...string) *Result {
	return &Result{
		Success:     false,
		Error:       err,
		Suggestions: suggestions,
		ExitCode:    1,
	}
}

// ExecContext provides execution dependencies
type ExecContext struct {
	Config   *config.Config
	Backend  backend.Backend
	UI       UI
	Registry *Registry // Command registry for composition
	DataDir  string    // Data directory for plugin loading
}

// UI interface for user interaction
type UI interface {
	Write(s string)
	WriteLine(s string)
	WriteError(s string)
	Confirm(prompt string) bool
	Spinner(message string) func()
}
