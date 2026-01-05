package testutil

import (
	"context"

	"github.com/scmd/scmd/internal/backend"
)

// MockBackend implements backend.Backend for testing
type MockBackend struct {
	response string
	err      error
}

// NewMockBackend creates a new mock backend
func NewMockBackend() *MockBackend {
	return &MockBackend{
		response: "Mock response",
	}
}

// SetResponse sets the response to return
func (b *MockBackend) SetResponse(response string) {
	b.response = response
}

// SetError sets the error to return
func (b *MockBackend) SetError(err error) {
	b.err = err
}

// Name returns the backend name
func (b *MockBackend) Name() string { return "test-mock" }

// Type returns the backend type
func (b *MockBackend) Type() backend.Type { return backend.TypeMock }

// SupportsToolCalling returns false
func (b *MockBackend) SupportsToolCalling() bool { return false }

// Initialize initializes the backend
func (b *MockBackend) Initialize(_ context.Context) error {
	return nil
}

// IsAvailable returns true if the backend is available
func (b *MockBackend) IsAvailable(_ context.Context) (bool, error) {
	return true, nil
}

// Shutdown shuts down the backend
func (b *MockBackend) Shutdown(_ context.Context) error {
	return nil
}

// Stream returns a streaming response
func (b *MockBackend) Stream(_ context.Context, _ *backend.CompletionRequest) (<-chan backend.StreamChunk, error) {
	if b.err != nil {
		return nil, b.err
	}

	ch := make(chan backend.StreamChunk)
	go func() {
		defer close(ch)
		ch <- backend.StreamChunk{Content: b.response}
		ch <- backend.StreamChunk{Done: true}
	}()
	return ch, nil
}

// Complete performs a completion request
func (b *MockBackend) Complete(_ context.Context, _ *backend.CompletionRequest) (*backend.CompletionResponse, error) {
	if b.err != nil {
		return nil, b.err
	}
	return &backend.CompletionResponse{
		Content:      b.response,
		FinishReason: backend.FinishComplete,
	}, nil
}

// CompleteWithTools is not supported
func (b *MockBackend) CompleteWithTools(_ context.Context, _ *backend.ToolRequest) (*backend.ToolResponse, error) {
	return nil, nil
}

// ModelInfo returns mock model info
func (b *MockBackend) ModelInfo() *backend.ModelInfo {
	return &backend.ModelInfo{
		Name:          "test-mock-model",
		ContextLength: 8192,
	}
}

// EstimateTokens estimates tokens
func (b *MockBackend) EstimateTokens(text string) int {
	return len(text) / 4
}
