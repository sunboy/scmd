// Package tools provides built-in tools for LLM tool calling
package tools

import (
	"context"

	"github.com/scmd/scmd/internal/backend"
)

// Tool defines the interface for executable tools
type Tool interface {
	// Name returns the tool name
	Name() string

	// Description returns what the tool does
	Description() string

	// Parameters returns the tool's parameter schema
	Parameters() map[string]backend.ToolParameter

	// Execute runs the tool with given parameters
	Execute(ctx context.Context, params map[string]interface{}) (*Result, error)

	// RequiresConfirmation returns true if tool needs user confirmation
	RequiresConfirmation() bool
}

// Result is the result of a tool execution
type Result struct {
	Success bool
	Output  string
	Error   string
}

// Registry manages available tools
type Registry struct {
	tools     map[string]Tool
	enabled   map[string]bool
	confirmUI ConfirmUI
}

// ConfirmUI handles user confirmation prompts
type ConfirmUI interface {
	Confirm(message string) bool
}

// NewRegistry creates a new tool registry
func NewRegistry(confirmUI ConfirmUI) *Registry {
	return &Registry{
		tools:     make(map[string]Tool),
		enabled:   make(map[string]bool),
		confirmUI: confirmUI,
	}
}

// Register adds a tool to the registry
func (r *Registry) Register(tool Tool) {
	r.tools[tool.Name()] = tool
	r.enabled[tool.Name()] = true
}

// Get retrieves a tool by name
func (r *Registry) Get(name string) (Tool, bool) {
	tool, ok := r.tools[name]
	if !ok || !r.enabled[name] {
		return nil, false
	}
	return tool, true
}

// Enable enables a tool
func (r *Registry) Enable(name string) {
	if _, ok := r.tools[name]; ok {
		r.enabled[name] = true
	}
}

// Disable disables a tool
func (r *Registry) Disable(name string) {
	r.enabled[name] = false
}

// List returns all registered tool names
func (r *Registry) List() []string {
	names := make([]string, 0, len(r.tools))
	for name := range r.tools {
		if r.enabled[name] {
			names = append(names, name)
		}
	}
	return names
}

// ToBackendTools converts tools to backend.ToolDefinition format
func (r *Registry) ToBackendTools() []backend.ToolDefinition {
	tools := make([]backend.ToolDefinition, 0, len(r.tools))
	for name, tool := range r.tools {
		if !r.enabled[name] {
			continue
		}
		tools = append(tools, backend.ToolDefinition{
			Name:        tool.Name(),
			Description: tool.Description(),
			Parameters:  tool.Parameters(),
		})
	}
	return tools
}

// Execute executes a tool by name with given parameters
func (r *Registry) Execute(ctx context.Context, name string, params map[string]interface{}) (*Result, error) {
	tool, ok := r.Get(name)
	if !ok {
		return &Result{
			Success: false,
			Error:   "tool not found or disabled: " + name,
		}, nil
	}

	// Check if confirmation is required
	if tool.RequiresConfirmation() && r.confirmUI != nil {
		if !r.confirmUI.Confirm("Execute " + name + "?") {
			return &Result{
				Success: false,
				Error:   "user cancelled operation",
			}, nil
		}
	}

	return tool.Execute(ctx, params)
}
