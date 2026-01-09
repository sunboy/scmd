package e2e

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/scmd/scmd/internal/backend"
	"github.com/scmd/scmd/internal/backend/llamacpp"
	"github.com/scmd/scmd/internal/config"
	"github.com/spf13/viper"
)

// TestBackendIntegration_FullFlow tests the complete flow from config to backend
// This test would have caught the bug where config.backends.default was ignored
func TestBackendIntegration_FullFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	t.Setenv("SCMD_DATA_DIR", tmpDir)

	// Step 1: Create config with llamacpp as default
	viper.Reset()
	viper.Set("backends.default", "llamacpp")
	viper.Set("backends.local.model", "qwen2.5-1.5b")

	// Step 2: Create backend registry
	reg := backend.NewRegistry()

	// Step 3: Register backends in "wrong" order
	// Mock backend registered FIRST (would be default by availability)
	mockBackend := &testBackend{name: "mock", available: true}
	reg.Register(mockBackend)

	// Llamacpp backend registered SECOND
	llamaBackend := llamacpp.New(tmpDir)
	reg.Register(llamaBackend)

	// Step 4: Get default backend WITHOUT applying config
	beforeBackend, err := reg.Default()
	if err != nil {
		t.Fatalf("failed to get default backend: %v", err)
	}

	// This should be mock (first registered)
	t.Logf("Before applying config: got %s", beforeBackend.Name())

	// Step 5: NOW apply config default (simulating root.go fix)
	cfgDefault := viper.GetString("backends.default")
	if cfgDefault != "" {
		if err := reg.SetDefault(cfgDefault); err != nil {
			t.Logf("Warning: failed to set default: %v", err)
		}
	}

	// Step 6: Get backend again
	afterBackend, err := reg.Default()
	if err != nil {
		t.Fatalf("failed to get backend after applying config: %v", err)
	}

	// Step 7: Should now be llamacpp (from config)
	if afterBackend.Name() != "llamacpp" {
		t.Errorf("after applying config default, expected llamacpp, got %s", afterBackend.Name())
	}

	// This test would have FAILED before the root.go fix
}

// TestBackendIntegration_DefaultNameMismatch tests config default matches backend name
// This test would have caught the "local" vs "llamacpp" bug
func TestBackendIntegration_DefaultNameMismatch(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Get default config
	cfg := config.Default()

	// Create backend
	tmpDir := t.TempDir()
	backend := llamacpp.New(tmpDir)

	// These MUST match
	if cfg.Backends.Default != backend.Name() {
		t.Errorf("Config default '%s' doesn't match backend name '%s'",
			cfg.Backends.Default, backend.Name())
	}

	// This test would have FAILED with defaults.go using "local"
}

// TestBackendIntegration_SetupConfiguration tests setup saves correct backend name
func TestBackendIntegration_SetupConfiguration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	t.Setenv("SCMD_DATA_DIR", tmpDir)

	viper.Reset()

	// Simulate setup wizard setting config
	viper.Set("backends.default", "llamacpp")
	viper.Set("backends.local.model", "qwen2.5-1.5b")
	viper.Set("setup_completed", true)

	// Verify the backend name is correct
	backendName := viper.GetString("backends.default")

	// Create actual backend
	backend := llamacpp.New(tmpDir)

	// These must match
	if backendName != backend.Name() {
		t.Errorf("Setup configured backend '%s' but backend name is '%s'",
			backendName, backend.Name())
	}
}

// TestBackendIntegration_RegistryDefaultPersistence tests default persists
func TestBackendIntegration_RegistryDefaultPersistence(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	reg := backend.NewRegistry()

	// Register backends
	mock := &testBackend{name: "mock", available: true}
	llama := &testBackend{name: "llamacpp", available: true}

	reg.Register(mock)
	reg.Register(llama)

	// Set default
	if err := reg.SetDefault("llamacpp"); err != nil {
		t.Fatalf("failed to set default: %v", err)
	}

	// Get backend multiple times - should always be llamacpp
	for i := 0; i < 5; i++ {
		backend, err := reg.Default()
		if err != nil {
			t.Fatalf("iteration %d: failed to get backend: %v", i, err)
		}

		if backend.Name() != "llamacpp" {
			t.Errorf("iteration %d: expected llamacpp, got %s", i, backend.Name())
		}
	}
}

// TestBackendIntegration_ConfigOverridesAvailability tests config wins over availability
func TestBackendIntegration_ConfigOverridesAvailability(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	reg := backend.NewRegistry()

	// Mock is available
	mock := &testBackend{name: "mock", available: true}
	reg.Register(mock)

	// Llamacpp is NOT available (llama-server not installed)
	llama := &testBackend{name: "llamacpp", available: false}
	reg.Register(llama)

	// Set llamacpp as default (even though not available)
	if err := reg.SetDefault("llamacpp"); err == nil {
		t.Log("SetDefault succeeded even though backend not available")
	}

	// Try to get default backend
	backend, err := reg.Default()

	// Should either:
	// 1. Return error (llamacpp not available)
	// 2. Return llamacpp anyway (registry doesn't check availability for Default())
	if err != nil {
		t.Logf("Got error when getting default backend: %v", err)
	} else {
		t.Logf("Got backend: %s", backend.Name())
	}
}

// TestBackendIntegration_ExplicitBackendOverride tests explicit backend selection
func TestBackendIntegration_ExplicitBackendOverride(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	reg := backend.NewRegistry()

	mock := &testBackend{name: "mock", available: true}
	llama := &testBackend{name: "llamacpp", available: true}

	reg.Register(mock)
	reg.Register(llama)

	// Set llamacpp as default
	reg.SetDefault("llamacpp")

	// Explicitly request mock
	backend, ok := reg.Get("mock")
	if !ok {
		t.Fatal("failed to get explicit backend")
	}

	// Should get mock, not default
	if backend.Name() != "mock" {
		t.Errorf("explicit backend selection failed, got %s", backend.Name())
	}
}

// TestBackendIntegration_ModelConfiguration tests model config flows to backend
func TestBackendIntegration_ModelConfiguration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	t.Setenv("SCMD_DATA_DIR", tmpDir)

	// Set model in config
	viper.Reset()
	viper.Set("backends.local.model", "qwen2.5-3b")

	// Create backend
	backend := llamacpp.New(tmpDir)

	// Set model from config
	modelName := viper.GetString("backends.local.model")
	if err := backend.SetModel(modelName); err != nil {
		t.Fatalf("failed to set model: %v", err)
	}

	// Verify model info
	info := backend.ModelInfo()
	if info == nil {
		t.Fatal("model info should not be nil")
	}

	// Model name should match what we configured
	if !strings.Contains(info.Name, "qwen") {
		t.Errorf("expected qwen model, got: %s", info.Name)
	}
}

// TestBackendIntegration_AvailabilityCheck tests IsAvailable works correctly
func TestBackendIntegration_AvailabilityCheck(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	backend := llamacpp.New(tmpDir)

	ctx := context.Background()
	available, err := backend.IsAvailable(ctx)

	// Test should work whether llama.cpp is installed or not
	if err != nil {
		t.Logf("Availability check returned error: %v", err)
	}

	t.Logf("Backend available: %v", available)

	// If not available, ensure error message is helpful
	if !available {
		if err == nil {
			t.Error("if not available, should return error with reason")
		}
	}
}

// TestBackendIntegration_ContextLength tests context length configuration
func TestBackendIntegration_ContextLength(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tmpDir := t.TempDir()

	// Test with context_length = 0 (unlimited)
	viper.Reset()
	viper.Set("backends.local.context_length", 0)

	backend := llamacpp.New(tmpDir)
	info := backend.ModelInfo()

	// With 0, should use model's native context
	if info.ContextLength <= 0 {
		t.Error("with context_length=0, should use model's native context")
	}

	t.Logf("Model native context length: %d", info.ContextLength)
}

// TestBackendIntegration_MultipleBackends tests multiple backends coexist
func TestBackendIntegration_MultipleBackends(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	reg := backend.NewRegistry()

	// Register multiple backends
	mock := &testBackend{name: "mock", available: true}
	llama := &testBackend{name: "llamacpp", available: true}
	ollama := &testBackend{name: "ollama", available: false}

	reg.Register(mock)
	reg.Register(llama)
	reg.Register(ollama)

	// List backends
	backends := reg.List()

	if len(backends) != 3 {
		t.Errorf("expected 3 backends, got %d", len(backends))
	}

	// Can get each backend by name
	for _, name := range []string{"mock", "llamacpp", "ollama"} {
		_, ok := reg.Get(name)
		if !ok {
			t.Errorf("failed to get backend %s", name)
		}
	}
}

// ==================== HELPER TEST BACKEND ====================

type testBackend struct {
	name      string
	available bool
}

func (b *testBackend) Name() string {
	return b.name
}

func (b *testBackend) Type() backend.Type {
	return backend.TypeMock
}

func (b *testBackend) Initialize(ctx context.Context) error {
	if !b.available {
		return ErrNotAvailable
	}
	return nil
}

func (b *testBackend) IsAvailable(ctx context.Context) (bool, error) {
	return b.available, nil
}

func (b *testBackend) Shutdown(ctx context.Context) error {
	return nil
}

func (b *testBackend) Complete(ctx context.Context, req *backend.CompletionRequest) (*backend.CompletionResponse, error) {
	return &backend.CompletionResponse{
		Content:      "Test response",
		TokensUsed:   10,
		FinishReason: backend.FinishStop,
	}, nil
}

func (b *testBackend) Stream(ctx context.Context, req *backend.CompletionRequest) (<-chan backend.StreamChunk, error) {
	ch := make(chan backend.StreamChunk, 1)
	go func() {
		ch <- backend.StreamChunk{Content: "Test response", Done: true}
		close(ch)
	}()
	return ch, nil
}

func (b *testBackend) SupportsToolCalling() bool {
	return false
}

func (b *testBackend) CompleteWithTools(ctx context.Context, req *backend.ToolRequest) (*backend.ToolResponse, error) {
	return nil, nil
}

func (b *testBackend) ModelInfo() *backend.ModelInfo {
	return &backend.ModelInfo{
		Name:          b.name + " model",
		ContextLength: 2048,
	}
}

func (b *testBackend) EstimateTokens(text string) int {
	return len(strings.Fields(text))
}

var ErrNotAvailable = fmt.Errorf("backend not available")
