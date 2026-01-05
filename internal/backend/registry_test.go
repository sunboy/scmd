package backend

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testBackend struct {
	name      string
	available bool
}

func (b *testBackend) Name() string                                      { return b.name }
func (b *testBackend) Type() Type                                        { return TypeMock }
func (b *testBackend) Initialize(_ context.Context) error                { return nil }
func (b *testBackend) IsAvailable(_ context.Context) (bool, error)       { return b.available, nil }
func (b *testBackend) Shutdown(_ context.Context) error                  { return nil }
func (b *testBackend) Complete(_ context.Context, _ *CompletionRequest) (*CompletionResponse, error) {
	return &CompletionResponse{Content: "test"}, nil
}
func (b *testBackend) Stream(_ context.Context, _ *CompletionRequest) (<-chan StreamChunk, error) {
	return nil, nil
}
func (b *testBackend) SupportsToolCalling() bool                           { return false }
func (b *testBackend) CompleteWithTools(_ context.Context, _ *ToolRequest) (*ToolResponse, error) {
	return nil, nil
}
func (b *testBackend) ModelInfo() *ModelInfo        { return &ModelInfo{Name: "test"} }
func (b *testBackend) EstimateTokens(_ string) int  { return 0 }

func TestRegistry_Register(t *testing.T) {
	r := NewRegistry()

	backend := &testBackend{name: "test"}
	err := r.Register(backend)
	require.NoError(t, err)

	found, ok := r.Get("test")
	assert.True(t, ok)
	assert.Equal(t, "test", found.Name())
}

func TestRegistry_Register_Duplicate(t *testing.T) {
	r := NewRegistry()

	backend1 := &testBackend{name: "test"}
	backend2 := &testBackend{name: "test"}

	err := r.Register(backend1)
	require.NoError(t, err)

	err = r.Register(backend2)
	assert.Error(t, err)
}

func TestRegistry_Get_NotFound(t *testing.T) {
	r := NewRegistry()

	_, ok := r.Get("nonexistent")
	assert.False(t, ok)
}

func TestRegistry_SetDefault(t *testing.T) {
	r := NewRegistry()

	backend := &testBackend{name: "test"}
	r.Register(backend)

	err := r.SetDefault("test")
	assert.NoError(t, err)

	def, err := r.Default()
	require.NoError(t, err)
	assert.Equal(t, "test", def.Name())
}

func TestRegistry_SetDefault_NotFound(t *testing.T) {
	r := NewRegistry()

	err := r.SetDefault("nonexistent")
	assert.Error(t, err)
}

func TestRegistry_Default_NoDefault(t *testing.T) {
	r := NewRegistry()

	backend := &testBackend{name: "test"}
	r.Register(backend)

	// Should return first backend when no default set
	def, err := r.Default()
	require.NoError(t, err)
	assert.NotNil(t, def)
}

func TestRegistry_Default_Empty(t *testing.T) {
	r := NewRegistry()

	_, err := r.Default()
	assert.Error(t, err)
}

func TestRegistry_List(t *testing.T) {
	r := NewRegistry()

	r.Register(&testBackend{name: "b1"})
	r.Register(&testBackend{name: "b2"})

	backends := r.List()
	assert.Len(t, backends, 2)
}

func TestRegistry_GetAvailable(t *testing.T) {
	r := NewRegistry()

	r.Register(&testBackend{name: "unavailable", available: false})
	r.Register(&testBackend{name: "available", available: true})

	backend, err := r.GetAvailable(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "available", backend.Name())
}

func TestRegistry_GetAvailable_None(t *testing.T) {
	r := NewRegistry()

	r.Register(&testBackend{name: "b1", available: false})

	_, err := r.GetAvailable(context.Background())
	assert.Error(t, err)
}
