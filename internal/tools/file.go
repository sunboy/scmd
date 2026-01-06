package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/scmd/scmd/internal/backend"
)

// ReadFileTool reads file contents
type ReadFileTool struct{}

// NewReadFileTool creates a new read file tool
func NewReadFileTool() *ReadFileTool {
	return &ReadFileTool{}
}

// Name returns the tool name
func (t *ReadFileTool) Name() string {
	return "read_file"
}

// Description returns the tool description
func (t *ReadFileTool) Description() string {
	return "Read the contents of a file"
}

// Parameters returns the parameter schema
func (t *ReadFileTool) Parameters() map[string]backend.ToolParameter {
	return map[string]backend.ToolParameter{
		"path": {
			Type:        "string",
			Description: "Path to the file to read",
			Required:    true,
		},
		"max_lines": {
			Type:        "number",
			Description: "Maximum number of lines to read (optional, default: all)",
			Required:    false,
		},
	}
}

// RequiresConfirmation returns false for read operations
func (t *ReadFileTool) RequiresConfirmation() bool {
	return false
}

// Execute reads a file
func (t *ReadFileTool) Execute(ctx context.Context, params map[string]interface{}) (*Result, error) {
	pathStr, ok := params["path"].(string)
	if !ok || pathStr == "" {
		return &Result{
			Success: false,
			Error:   "path parameter is required",
		}, nil
	}

	// Clean and resolve path
	pathStr = filepath.Clean(pathStr)

	// Read file
	content, err := os.ReadFile(pathStr)
	if err != nil {
		return &Result{
			Success: false,
			Error:   fmt.Sprintf("failed to read file: %v", err),
		}, nil
	}

	output := string(content)

	// Apply max_lines if specified
	if maxLines, ok := params["max_lines"].(float64); ok && maxLines > 0 {
		lines := strings.Split(output, "\n")
		if len(lines) > int(maxLines) {
			output = strings.Join(lines[:int(maxLines)], "\n")
			output += fmt.Sprintf("\n\n... (truncated, %d more lines)", len(lines)-int(maxLines))
		}
	}

	return &Result{
		Success: true,
		Output:  output,
	}, nil
}

// WriteFileTool writes to files
type WriteFileTool struct {
	confirmUI ConfirmUI
}

// NewWriteFileTool creates a new write file tool
func NewWriteFileTool(confirmUI ConfirmUI) *WriteFileTool {
	return &WriteFileTool{
		confirmUI: confirmUI,
	}
}

// Name returns the tool name
func (t *WriteFileTool) Name() string {
	return "write_file"
}

// Description returns the tool description
func (t *WriteFileTool) Description() string {
	return "Write content to a file. Use carefully as this modifies files."
}

// Parameters returns the parameter schema
func (t *WriteFileTool) Parameters() map[string]backend.ToolParameter {
	return map[string]backend.ToolParameter{
		"path": {
			Type:        "string",
			Description: "Path to the file to write",
			Required:    true,
		},
		"content": {
			Type:        "string",
			Description: "Content to write to the file",
			Required:    true,
		},
		"append": {
			Type:        "boolean",
			Description: "If true, append to file instead of overwriting (default: false)",
			Required:    false,
		},
	}
}

// RequiresConfirmation returns true for write operations
func (t *WriteFileTool) RequiresConfirmation() bool {
	return true
}

// Execute writes to a file
func (t *WriteFileTool) Execute(ctx context.Context, params map[string]interface{}) (*Result, error) {
	pathStr, ok := params["path"].(string)
	if !ok || pathStr == "" {
		return &Result{
			Success: false,
			Error:   "path parameter is required",
		}, nil
	}

	content, ok := params["content"].(string)
	if !ok {
		return &Result{
			Success: false,
			Error:   "content parameter is required",
		}, nil
	}

	// Clean path
	pathStr = filepath.Clean(pathStr)

	// Check if appending
	append := false
	if appendVal, ok := params["append"].(bool); ok {
		append = appendVal
	}

	// Additional confirmation with file details
	if t.confirmUI != nil {
		action := "overwrite"
		if append {
			action = "append to"
		}
		if _, err := os.Stat(pathStr); err == nil {
			if !t.confirmUI.Confirm(fmt.Sprintf("File exists. %s %s?", action, pathStr)) {
				return &Result{
					Success: false,
					Error:   "user cancelled operation",
				}, nil
			}
		}
	}

	// Ensure directory exists
	dir := filepath.Dir(pathStr)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return &Result{
			Success: false,
			Error:   fmt.Sprintf("failed to create directory: %v", err),
		}, nil
	}

	// Write file
	var err error
	if append {
		f, err := os.OpenFile(pathStr, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return &Result{
				Success: false,
				Error:   fmt.Sprintf("failed to open file for append: %v", err),
			}, nil
		}
		defer f.Close()
		_, err = f.WriteString(content)
	} else {
		err = os.WriteFile(pathStr, []byte(content), 0644)
	}

	if err != nil {
		return &Result{
			Success: false,
			Error:   fmt.Sprintf("failed to write file: %v", err),
		}, nil
	}

	action := "Written"
	if append {
		action = "Appended"
	}

	return &Result{
		Success: true,
		Output:  fmt.Sprintf("%s to %s (%d bytes)", action, pathStr, len(content)),
	}, nil
}
