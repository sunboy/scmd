// Package testutil provides utilities for testing scmd
package testutil

import (
	"bytes"
	"fmt"
)

// MockUI implements command.UI for testing
type MockUI struct {
	Output          bytes.Buffer
	Errors          bytes.Buffer
	Prompts         []string
	PromptResponses []bool
	promptIndex     int
}

// NewMockUI creates a new mock UI
func NewMockUI() *MockUI {
	return &MockUI{
		PromptResponses: []bool{true},
	}
}

// Write writes to the output buffer
func (u *MockUI) Write(s string) {
	u.Output.WriteString(s)
}

// WriteLine writes a line to the output buffer
func (u *MockUI) WriteLine(s string) {
	u.Output.WriteString(s + "\n")
}

// WriteError writes to the error buffer
func (u *MockUI) WriteError(s string) {
	u.Errors.WriteString(s + "\n")
}

// Confirm records the prompt and returns the configured response
func (u *MockUI) Confirm(prompt string) bool {
	u.Prompts = append(u.Prompts, prompt)
	if u.promptIndex < len(u.PromptResponses) {
		resp := u.PromptResponses[u.promptIndex]
		u.promptIndex++
		return resp
	}
	return true
}

// Spinner records the spinner message and returns a no-op function
func (u *MockUI) Spinner(message string) func() {
	u.Output.WriteString(fmt.Sprintf("[spinner] %s\n", message))
	return func() {
		u.Output.WriteString("[spinner done]\n")
	}
}

// GetOutput returns all output
func (u *MockUI) GetOutput() string {
	return u.Output.String()
}

// GetErrors returns all errors
func (u *MockUI) GetErrors() string {
	return u.Errors.String()
}

// Reset clears all buffers
func (u *MockUI) Reset() {
	u.Output.Reset()
	u.Errors.Reset()
	u.Prompts = nil
	u.promptIndex = 0
}

// SetConfirmResponse sets the responses for Confirm calls
func (u *MockUI) SetConfirmResponse(responses ...bool) {
	u.PromptResponses = responses
	u.promptIndex = 0
}
