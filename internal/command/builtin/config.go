package builtin

import (
	"context"
	"fmt"
	"strings"

	"github.com/scmd/scmd/internal/command"
	"github.com/scmd/scmd/internal/config"
)

// ConfigCommand implements /config
type ConfigCommand struct{}

// NewConfigCommand creates a new config command
func NewConfigCommand() *ConfigCommand {
	return &ConfigCommand{}
}

// Name returns the command name
func (c *ConfigCommand) Name() string { return "config" }

// Aliases returns command aliases
func (c *ConfigCommand) Aliases() []string { return []string{"cfg"} }

// Description returns the command description
func (c *ConfigCommand) Description() string { return "View or modify configuration" }

// Usage returns usage information
func (c *ConfigCommand) Usage() string { return "/config [key] [value]" }

// Category returns the command category
func (c *ConfigCommand) Category() command.Category { return command.CategoryConfig }

// RequiresBackend returns false
func (c *ConfigCommand) RequiresBackend() bool { return false }

// Examples returns example usages
func (c *ConfigCommand) Examples() []string {
	return []string{
		"/config",
		"/config backends.default",
		"/config ui.colors true",
	}
}

// Validate validates arguments
func (c *ConfigCommand) Validate(_ *command.Args) error {
	return nil
}

// Execute runs the config command
func (c *ConfigCommand) Execute(
	_ context.Context,
	args *command.Args,
	execCtx *command.ExecContext,
) (*command.Result, error) {
	// No args - show all config
	if len(args.Positional) == 0 {
		return c.showAllConfig(execCtx)
	}

	key := args.Positional[0]

	// One arg - show specific config
	if len(args.Positional) == 1 {
		return c.showConfig(key, execCtx)
	}

	// Two args - set config (not implemented yet)
	value := args.Positional[1]
	return c.setConfig(key, value, execCtx)
}

func (c *ConfigCommand) showAllConfig(execCtx *command.ExecContext) (*command.Result, error) {
	cfg := execCtx.Config

	var sb strings.Builder
	sb.WriteString("scmd configuration:\n\n")

	sb.WriteString(fmt.Sprintf("  version: %s\n", cfg.Version))
	sb.WriteString("\n  backends:\n")
	sb.WriteString(fmt.Sprintf("    default: %s\n", cfg.Backends.Default))
	sb.WriteString(fmt.Sprintf("    local.model: %s\n", cfg.Backends.Local.Model))
	sb.WriteString(fmt.Sprintf("    local.context_length: %d\n", cfg.Backends.Local.ContextLength))

	sb.WriteString("\n  ui:\n")
	sb.WriteString(fmt.Sprintf("    streaming: %t\n", cfg.UI.Streaming))
	sb.WriteString(fmt.Sprintf("    colors: %t\n", cfg.UI.Colors))
	sb.WriteString(fmt.Sprintf("    verbose: %t\n", cfg.UI.Verbose))

	sb.WriteString("\n  models:\n")
	sb.WriteString(fmt.Sprintf("    directory: %s\n", cfg.Models.Directory))
	sb.WriteString(fmt.Sprintf("    auto_download: %t\n", cfg.Models.AutoDownload))

	sb.WriteString(fmt.Sprintf("\nConfig file: %s\n", config.ConfigPath()))

	execCtx.UI.Write(sb.String())
	return &command.Result{Success: true}, nil
}

func (c *ConfigCommand) showConfig(key string, execCtx *command.ExecContext) (*command.Result, error) {
	cfg := execCtx.Config

	// Try string
	if v := cfg.GetString(key); v != "" {
		execCtx.UI.WriteLine(fmt.Sprintf("%s = %s", key, v))
		return &command.Result{Success: true}, nil
	}

	// Try bool
	v := cfg.GetBool(key)
	execCtx.UI.WriteLine(fmt.Sprintf("%s = %t", key, v))
	return &command.Result{Success: true}, nil
}

func (c *ConfigCommand) setConfig(key, value string, execCtx *command.ExecContext) (*command.Result, error) {
	// Parse the value based on key type
	cfg := execCtx.Config

	// Map of valid keys to their types
	validKeys := map[string]string{
		"backends.default":              "string",
		"backends.local.model":          "string",
		"backends.local.context_length": "int",
		"ui.streaming":                  "bool",
		"ui.colors":                     "bool",
		"ui.verbose":                    "bool",
		"models.auto_download":          "bool",
	}

	keyType, valid := validKeys[key]
	if !valid {
		return &command.Result{
			Success: false,
			Error:   fmt.Sprintf("unknown configuration key: %s", key),
			Suggestions: []string{
				"Use 'scmd config' to see all available keys",
			},
		}, nil
	}

	// Update the config based on type
	switch keyType {
	case "string":
		if err := cfg.Set(key, value); err != nil {
			return &command.Result{
				Success: false,
				Error:   fmt.Sprintf("failed to set %s: %v", key, err),
			}, nil
		}
	case "bool":
		boolVal := value == "true" || value == "1" || value == "yes"
		if err := cfg.Set(key, boolVal); err != nil {
			return &command.Result{
				Success: false,
				Error:   fmt.Sprintf("failed to set %s: %v", key, err),
			}, nil
		}
	case "int":
		// Parse int value
		var intVal int
		if _, err := fmt.Sscanf(value, "%d", &intVal); err != nil {
			return &command.Result{
				Success: false,
				Error:   fmt.Sprintf("invalid integer value for %s: %s", key, value),
			}, nil
		}
		if err := cfg.Set(key, intVal); err != nil {
			return &command.Result{
				Success: false,
				Error:   fmt.Sprintf("failed to set %s: %v", key, err),
			}, nil
		}
	}

	// Save the config
	if err := config.Save(cfg); err != nil {
		return &command.Result{
			Success: false,
			Error:   fmt.Sprintf("failed to save config: %v", err),
		}, nil
	}

	execCtx.UI.WriteLine(fmt.Sprintf("Set %s = %s", key, value))
	execCtx.UI.WriteLine(fmt.Sprintf("Configuration saved to %s", config.ConfigPath()))
	return &command.Result{Success: true}, nil
}
