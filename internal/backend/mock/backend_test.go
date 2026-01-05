package mock

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/scmd/scmd/internal/backend"
)

func TestNew(t *testing.T) {
	b := New()
	assert.NotNil(t, b)
	assert.Equal(t, "mock", b.Name())
}

func TestBackend_Type(t *testing.T) {
	b := New()
	assert.Equal(t, backend.TypeMock, b.Type())
}

func TestBackend_Initialize(t *testing.T) {
	b := New()
	err := b.Initialize(context.Background())
	assert.NoError(t, err)
}

func TestBackend_IsAvailable(t *testing.T) {
	b := New()
	available, err := b.IsAvailable(context.Background())
	assert.NoError(t, err)
	assert.True(t, available)
}

func TestBackend_Shutdown(t *testing.T) {
	b := New()
	err := b.Shutdown(context.Background())
	assert.NoError(t, err)
}

func TestBackend_Complete(t *testing.T) {
	b := New()
	b.SetResponse("test response")

	req := &backend.CompletionRequest{
		Prompt: "test prompt",
	}

	resp, err := b.Complete(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, "test response", resp.Content)
	assert.Equal(t, backend.FinishComplete, resp.FinishReason)
}

func TestBackend_Complete_Error(t *testing.T) {
	b := New()
	b.SetError(errors.New("test error"))

	req := &backend.CompletionRequest{}

	_, err := b.Complete(context.Background(), req)
	assert.Error(t, err)
}

func TestBackend_Stream(t *testing.T) {
	b := New()
	b.SetResponse("hello world")

	req := &backend.CompletionRequest{}

	ch, err := b.Stream(context.Background(), req)
	require.NoError(t, err)

	var content string
	for chunk := range ch {
		if chunk.Done {
			break
		}
		content += chunk.Content
	}

	assert.Equal(t, "hello world", content)
}

func TestBackend_Stream_Error(t *testing.T) {
	b := New()
	b.SetError(errors.New("test error"))

	req := &backend.CompletionRequest{}

	_, err := b.Stream(context.Background(), req)
	assert.Error(t, err)
}

func TestBackend_SupportsToolCalling(t *testing.T) {
	b := New()
	assert.False(t, b.SupportsToolCalling())
}

func TestBackend_CompleteWithTools(t *testing.T) {
	b := New()
	resp, err := b.CompleteWithTools(context.Background(), nil)
	assert.NoError(t, err)
	assert.Nil(t, resp)
}

func TestBackend_ModelInfo(t *testing.T) {
	b := New()
	info := b.ModelInfo()

	assert.NotNil(t, info)
	assert.Equal(t, "mock-model", info.Name)
	assert.Equal(t, 8192, info.ContextLength)
}

func TestBackend_EstimateTokens(t *testing.T) {
	b := New()
	tokens := b.EstimateTokens("hello world")
	assert.Greater(t, tokens, 0)
}
