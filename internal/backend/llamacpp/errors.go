package llamacpp

import (
	"fmt"
	"strings"
)

// BackendError represents a llamacpp-specific error with helpful context
type BackendError struct {
	Type        ErrorType
	Message     string
	Cause       error
	Suggestions []string
}

// ErrorType categorizes the error for better handling
type ErrorType string

const (
	ErrorServerNotRunning ErrorType = "server_not_running"
	ErrorServerNotFound   ErrorType = "server_not_found"
	ErrorOutOfMemory      ErrorType = "out_of_memory"
	ErrorConnectionFailed ErrorType = "connection_failed"
	ErrorModelNotFound    ErrorType = "model_not_found"
	ErrorTimeout          ErrorType = "timeout"
	ErrorInference        ErrorType = "inference_failed"
)

// Error implements the error interface
func (e *BackendError) Error() string {
	var sb strings.Builder

	// Main error message
	sb.WriteString(fmt.Sprintf("❌ %s\n", e.Message))

	// Add cause if available
	if e.Cause != nil {
		sb.WriteString(fmt.Sprintf("\nCause: %v\n", e.Cause))
	}

	// Add suggestions
	if len(e.Suggestions) > 0 {
		sb.WriteString("\nSolutions:\n")
		for i, suggestion := range e.Suggestions {
			sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, suggestion))
		}
	}

	return sb.String()
}

// Unwrap returns the underlying error
func (e *BackendError) Unwrap() error {
	return e.Cause
}

// NewServerNotRunningError creates an error when server is not running
func NewServerNotRunningError(cause error) *BackendError {
	return &BackendError{
		Type:    ErrorServerNotRunning,
		Message: "Cannot connect to llama-server",
		Cause:   cause,
		Suggestions: []string{
			"scmd will auto-start llama-server on the next command",
			"Or start manually: llama-server -m ~/.scmd/models/MODEL.gguf --port 8089",
			"Use a cloud provider: export OPENAI_API_KEY=your-key && scmd -b openai",
			"Run 'scmd doctor' to diagnose issues",
		},
	}
}

// NewServerNotFoundError creates an error when llama-server binary is not found
func NewServerNotFoundError(cause error) *BackendError {
	return &BackendError{
		Type:    ErrorServerNotFound,
		Message: "llama-server not found",
		Cause:   cause,
		Suggestions: []string{
			"Install llama.cpp: brew install llama.cpp",
			"Or build from source: https://github.com/ggerganov/llama.cpp",
			"Use a cloud provider instead: export OPENAI_API_KEY=your-key",
			"Use Ollama: brew install ollama && ollama serve",
		},
	}
}

// NewOutOfMemoryError creates an error for GPU/RAM OOM situations
func NewOutOfMemoryError(cause error) *BackendError {
	return &BackendError{
		Type:    ErrorOutOfMemory,
		Message: "Out of memory - GPU or RAM exhausted",
		Cause:   cause,
		Suggestions: []string{
			"Restart with CPU mode: scmd server restart --cpu",
			"Use smaller context: scmd server restart -c 2048",
			"Switch to smaller model: scmd models switch qwen2.5-1.5b",
			"Use cloud provider: export OPENAI_API_KEY=your-key",
			"Check memory: scmd doctor",
		},
	}
}

// NewConnectionFailedError creates an error for connection failures
func NewConnectionFailedError(cause error) *BackendError {
	return &BackendError{
		Type:    ErrorConnectionFailed,
		Message: "Failed to connect to llama-server",
		Cause:   cause,
		Suggestions: []string{
			"Check if server is running: scmd server status",
			"Restart server: scmd server restart",
			"Check port 8089 is not blocked by firewall",
			"Run diagnostics: scmd doctor",
		},
	}
}

// NewModelNotFoundError creates an error when model file is missing
func NewModelNotFoundError(modelName string, cause error) *BackendError {
	return &BackendError{
		Type:    ErrorModelNotFound,
		Message: fmt.Sprintf("Model '%s' not found", modelName),
		Cause:   cause,
		Suggestions: []string{
			fmt.Sprintf("Download model: scmd models download %s", modelName),
			"List available models: scmd models list",
			"Use different model: scmd -m qwen2.5-3b",
			"Use cloud provider: export OPENAI_API_KEY=your-key",
		},
	}
}

// NewTimeoutError creates an error for operation timeouts
func NewTimeoutError(operation string, cause error) *BackendError {
	return &BackendError{
		Type:    ErrorTimeout,
		Message: fmt.Sprintf("Operation timed out: %s", operation),
		Cause:   cause,
		Suggestions: []string{
			"Server may be slow to start - wait a moment and retry",
			"Check system resources: scmd doctor",
			"Try smaller model: scmd -m qwen2.5-1.5b",
			"Use faster backend: scmd -b openai",
		},
	}
}

// NewInferenceError creates an error for inference failures
func NewInferenceError(cause error) *BackendError {
	return &BackendError{
		Type:    ErrorInference,
		Message: "Inference failed",
		Cause:   cause,
		Suggestions: []string{
			"Check server logs: tail ~/.scmd/logs/llama-server.log",
			"Restart server: scmd server restart",
			"Run diagnostics: scmd doctor",
			"Try different backend: scmd -b openai",
		},
	}
}

// ParseError detects error types from error messages and creates appropriate errors
func ParseError(err error) error {
	if err == nil {
		return nil
	}

	errStr := strings.ToLower(err.Error())

	// Check for specific error patterns
	if strings.Contains(errStr, "connection refused") || strings.Contains(errStr, "eof") {
		return NewServerNotRunningError(err)
	}

	if strings.Contains(errStr, "out of memory") || strings.Contains(errStr, "kIOGPUCommandBufferCallbackErrorOutOfMemory") {
		return NewOutOfMemoryError(err)
	}

	if strings.Contains(errStr, "llama-server not found") || strings.Contains(errStr, "executable file not found") {
		return NewServerNotFoundError(err)
	}

	if strings.Contains(errStr, "timeout") || strings.Contains(errStr, "deadline exceeded") {
		return NewTimeoutError("server start", err)
	}

	if strings.Contains(errStr, "model not found") || strings.Contains(errStr, "no such file") {
		return NewModelNotFoundError("unknown", err)
	}

	// Default to generic inference error
	return NewInferenceError(err)
}

// WrapError wraps an error with helpful context
func WrapError(err error, context string) error {
	if err == nil {
		return nil
	}

	// Try to parse and enhance the error
	parsed := ParseError(err)
	if backendErr, ok := parsed.(*BackendError); ok {
		backendErr.Message = fmt.Sprintf("%s: %s", context, backendErr.Message)
		return backendErr
	}

	return fmt.Errorf("%s: %w", context, err)
}
