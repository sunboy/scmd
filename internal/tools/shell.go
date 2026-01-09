package tools

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/scmd/scmd/internal/backend"
	"github.com/scmd/scmd/internal/preview"
)

// ShellTool executes shell commands
type ShellTool struct {
	allowedCommands map[string]bool
	timeout         time.Duration
	confirmUI       ConfirmUI
}

// NewShellTool creates a new shell tool
func NewShellTool(confirmUI ConfirmUI) *ShellTool {
	return &ShellTool{
		allowedCommands: getDefaultAllowedCommands(),
		timeout:         30 * time.Second,
		confirmUI:       confirmUI,
	}
}

// Name returns the tool name
func (t *ShellTool) Name() string {
	return "shell"
}

// Description returns the tool description
func (t *ShellTool) Description() string {
	return "Execute shell commands. Use for running CLI tools, checking status, or system operations."
}

// Parameters returns the parameter schema
func (t *ShellTool) Parameters() map[string]backend.ToolParameter {
	return map[string]backend.ToolParameter{
		"command": {
			Type:        "string",
			Description: "The shell command to execute",
			Required:    true,
		},
		"working_dir": {
			Type:        "string",
			Description: "Working directory for command execution (optional)",
			Required:    false,
		},
	}
}

// RequiresConfirmation returns true for potentially destructive commands
func (t *ShellTool) RequiresConfirmation() bool {
	return true // Always confirm shell execution
}

// Execute runs a shell command
func (t *ShellTool) Execute(ctx context.Context, params map[string]interface{}) (*Result, error) {
	cmdStr, ok := params["command"].(string)
	if !ok || cmdStr == "" {
		return &Result{
			Success: false,
			Error:   "command parameter is required",
		}, nil
	}

	// Parse command
	parts := strings.Fields(cmdStr)
	if len(parts) == 0 {
		return &Result{
			Success: false,
			Error:   "empty command",
		}, nil
	}

	// Check if command is allowed (basic security)
	baseCmd := parts[0]
	if !t.isAllowed(baseCmd) {
		return &Result{
			Success: false,
			Error:   fmt.Sprintf("command '%s' is not allowed for security reasons", baseCmd),
		}, nil
	}

	// Detect if command is destructive and show preview
	detectResult := preview.Detect(cmdStr)
	if detectResult.IsDestructive && detectResult.HighestSeverity >= preview.SeverityMedium {
		// Use preview buffer for interactive review
		buffer := preview.NewBuffer(cmdStr)
		action, finalCmd, err := buffer.Show()
		if err != nil {
			return &Result{
				Success: false,
				Error:   fmt.Sprintf("preview error: %v", err),
			}, nil
		}

		switch action {
		case preview.ActionQuit:
			return &Result{
				Success: false,
				Output:  "Command cancelled by user",
			}, nil

		case preview.ActionDryRun:
			// Dry run - show what would happen without executing
			return &Result{
				Success: true,
				Output:  fmt.Sprintf("[DRY RUN] Would execute: %s\n\nNo actual changes made.", finalCmd),
			}, nil

		case preview.ActionEdit, preview.ActionExecute:
			// Update command string to potentially edited version
			cmdStr = finalCmd
			parts = strings.Fields(cmdStr)
			if len(parts) == 0 {
				return &Result{
					Success: false,
					Error:   "edited command is empty",
				}, nil
			}

			// Re-check if edited command is allowed
			baseCmd = parts[0]
			if !t.isAllowed(baseCmd) {
				return &Result{
					Success: false,
					Error:   fmt.Sprintf("edited command '%s' is not allowed for security reasons", baseCmd),
				}, nil
			}
			// Continue to execution below
		}
	}

	// Create command with timeout
	cmdCtx, cancel := context.WithTimeout(ctx, t.timeout)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, parts[0], parts[1:]...)

	// Set working directory if provided
	if workDir, ok := params["working_dir"].(string); ok && workDir != "" {
		cmd.Dir = workDir
	}

	// Capture output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Execute
	err := cmd.Run()

	// Build result
	output := stdout.String()
	if stderr.Len() > 0 {
		if output != "" {
			output += "\n\nStderr:\n" + stderr.String()
		} else {
			output = stderr.String()
		}
	}

	if err != nil {
		return &Result{
			Success: false,
			Output:  output,
			Error:   fmt.Sprintf("command failed: %v", err),
		}, nil
	}

	return &Result{
		Success: true,
		Output:  output,
	}, nil
}

// isAllowed checks if a command is in the allowed list
func (t *ShellTool) isAllowed(cmd string) bool {
	// Remove path if present
	parts := strings.Split(cmd, "/")
	baseName := parts[len(parts)-1]

	return t.allowedCommands[baseName] || t.allowedCommands[cmd]
}

// AllowCommand adds a command to the allowed list
func (t *ShellTool) AllowCommand(cmd string) {
	t.allowedCommands[cmd] = true
}

// DenyCommand removes a command from the allowed list
func (t *ShellTool) DenyCommand(cmd string) {
	delete(t.allowedCommands, cmd)
}

// getDefaultAllowedCommands returns the default safe commands
func getDefaultAllowedCommands() map[string]bool {
	return map[string]bool{
		// File operations (read-only)
		"ls":   true,
		"cat":  true,
		"head": true,
		"tail": true,
		"find": true,
		"grep": true,
		"wc":   true,
		"diff": true,
		"file": true,
		"stat": true,
		"du":   true,
		"df":   true,

		// Git commands
		"git": true,

		// System info
		"pwd":    true,
		"whoami": true,
		"date":   true,
		"uname":  true,
		"which":  true,
		"env":    true,
		"echo":   true,
		"printf": true,

		// Network (read-only)
		"curl":     true,
		"wget":     true,
		"ping":     true,
		"dig":      true,
		"nslookup": true,

		// Development tools
		"go":      true,
		"python":  true,
		"python3": true,
		"node":    true,
		"npm":     true,
		"cargo":   true,
		"make":    true,
		"docker":  true,
		"kubectl": true,

		// Text processing
		"awk":  true,
		"sed":  true,
		"sort": true,
		"uniq": true,
		"cut":  true,
		"tr":   true,
		"jq":   true,

		// Common utilities
		"test": true,
		"expr": true,
		"bc":   true,
		"cal":  true,
		"man":  true,
		"help": true,
	}
}
