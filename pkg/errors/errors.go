// Package errors provides error types for scmd
package errors

import (
	"errors"
	"fmt"
)

// Standard errors
var (
	ErrNotFound      = errors.New("not found")
	ErrInvalidInput  = errors.New("invalid input")
	ErrPermission    = errors.New("permission denied")
	ErrTimeout       = errors.New("operation timed out")
	ErrBackendFailed = errors.New("backend operation failed")
	ErrConfigInvalid = errors.New("invalid configuration")
	ErrNoBackend     = errors.New("no backend available")
	ErrCanceled      = errors.New("operation canceled")
)

// CommandError represents a command execution error
type CommandError struct {
	Command     string
	Message     string
	Suggestions []string
	Cause       error
}

// Error returns the error message
func (e *CommandError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Command, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Command, e.Message)
}

// Unwrap returns the underlying error
func (e *CommandError) Unwrap() error {
	return e.Cause
}

// NewCommandError creates a new command error
func NewCommandError(cmd, msg string, suggestions ...string) *CommandError {
	return &CommandError{
		Command:     cmd,
		Message:     msg,
		Suggestions: suggestions,
	}
}

// Wrap wraps an error with command context
func Wrap(cmd string, err error) *CommandError {
	return &CommandError{
		Command: cmd,
		Message: err.Error(),
		Cause:   err,
	}
}

// WithSuggestions adds suggestions to an error
func (e *CommandError) WithSuggestions(suggestions ...string) *CommandError {
	e.Suggestions = suggestions
	return e
}

// ValidationError represents input validation errors
type ValidationError struct {
	Field   string
	Message string
}

// Error returns the validation error message
func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error: %s: %s", e.Field, e.Message)
}

// NewValidationError creates a new validation error
func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
	}
}
