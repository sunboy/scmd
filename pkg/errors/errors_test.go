package errors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommandError_Error(t *testing.T) {
	err := &CommandError{
		Command: "test",
		Message: "something went wrong",
	}

	assert.Equal(t, "test: something went wrong", err.Error())
}

func TestCommandError_Error_WithCause(t *testing.T) {
	cause := errors.New("root cause")
	err := &CommandError{
		Command: "test",
		Message: "something went wrong",
		Cause:   cause,
	}

	assert.Equal(t, "test: something went wrong: root cause", err.Error())
}

func TestCommandError_Unwrap(t *testing.T) {
	cause := errors.New("root cause")
	err := &CommandError{
		Command: "test",
		Message: "wrapper",
		Cause:   cause,
	}

	assert.Equal(t, cause, err.Unwrap())
}

func TestNewCommandError(t *testing.T) {
	err := NewCommandError("cmd", "message", "suggestion1", "suggestion2")

	assert.Equal(t, "cmd", err.Command)
	assert.Equal(t, "message", err.Message)
	assert.Equal(t, []string{"suggestion1", "suggestion2"}, err.Suggestions)
}

func TestWrap(t *testing.T) {
	cause := errors.New("original error")
	err := Wrap("cmd", cause)

	assert.Equal(t, "cmd", err.Command)
	assert.Equal(t, "original error", err.Message)
	assert.Equal(t, cause, err.Cause)
}

func TestCommandError_WithSuggestions(t *testing.T) {
	err := NewCommandError("cmd", "message")
	err = err.WithSuggestions("try this", "or this")

	assert.Equal(t, []string{"try this", "or this"}, err.Suggestions)
}

func TestValidationError_Error(t *testing.T) {
	err := &ValidationError{
		Field:   "name",
		Message: "cannot be empty",
	}

	assert.Equal(t, "validation error: name: cannot be empty", err.Error())
}

func TestNewValidationError(t *testing.T) {
	err := NewValidationError("email", "invalid format")

	assert.Equal(t, "email", err.Field)
	assert.Equal(t, "invalid format", err.Message)
}

func TestStandardErrors(t *testing.T) {
	assert.NotNil(t, ErrNotFound)
	assert.NotNil(t, ErrInvalidInput)
	assert.NotNil(t, ErrPermission)
	assert.NotNil(t, ErrTimeout)
	assert.NotNil(t, ErrBackendFailed)
	assert.NotNil(t, ErrConfigInvalid)
	assert.NotNil(t, ErrNoBackend)
	assert.NotNil(t, ErrCanceled)
}
