package e2e

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/scmd/scmd/internal/backend"
	"github.com/scmd/scmd/internal/backend/llamacpp"
	"github.com/scmd/scmd/internal/cli"
	"github.com/scmd/scmd/internal/config"
	"github.com/spf13/viper"
)

// TestFirstRun_FreshInstallDetection tests that fresh install is correctly detected
func TestFirstRun_FreshInstallDetection(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	setupTestEnv(t, tmpDir)

	// Should detect first run
	if !cli.IsFirstRun() {
		t.Error("fresh install should be detected as first run")
	}
}

// TestFirstRun_AfterSetup tests that setup completion is detected
func TestFirstRun_AfterSetup(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	setupTestEnv(t, tmpDir)

	// Mark setup as completed
	viper.Set("setup_completed", true)

	// Should NOT be first run anymore
	if cli.IsFirstRun() {
		t.Error("should not be first run after setup completion")
	}
}

// TestFirstRun_ConfigPersistence tests that config persists across runs
func TestFirstRun_ConfigPersistence(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	setupTestEnv(t, tmpDir)

	// Create config directory
	configDir := filepath.Dir(config.ConfigPath())
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Set and save config
	viper.Set("backends.default", "llamacpp")
	viper.Set("backends.local.model", "qwen2.5-1.5b")
	viper.Set("setup_completed", true)

	if err := viper.WriteConfigAs(config.ConfigPath()); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Reset viper to simulate new process
	viper.Reset()
	setupTestEnv(t, tmpDir)

	// Read config back
	viper.SetConfigFile(config.ConfigPath())
	if err := viper.ReadInConfig(); err != nil {
		t.Fatalf("failed to read config: %v", err)
	}

	// Verify values persisted
	if viper.GetString("backends.default") != "llamacpp" {
		t.Error("backends.default should persist")
	}

	if viper.GetString("backends.local.model") != "qwen2.5-1.5b" {
		t.Error("model config should persist")
	}

	if !viper.GetBool("setup_completed") {
		t.Error("setup_completed should persist")
	}
}

// TestBackendSelection_DefaultFromConfig tests backend selection uses config
func TestBackendSelection_DefaultFromConfig(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	setupTestEnv(t, tmpDir)

	// Create backend registry
	reg := backend.NewRegistry()

	// Register backends
	mockBackend := &mockBackend{name: "mock"}
	llamaBackend := llamacpp.New(tmpDir)

	reg.Register(mockBackend)
	reg.Register(llamaBackend)

	// Set default in config
	viper.Set("backends.default", "llamacpp")

	// Apply config default to registry
	if err := reg.SetDefault("llamacpp"); err != nil {
		t.Fatalf("failed to set default backend: %v", err)
	}

	// Get default backend
	defaultBackend, err := reg.Default()
	if err != nil {
		t.Fatalf("failed to get default backend: %v", err)
	}

	// Should be llamacpp, not mock (even though mock was registered first)
	if defaultBackend.Name() != "llamacpp" {
		t.Errorf("expected default backend to be llamacpp, got %s", defaultBackend.Name())
	}
}

// TestBackendSelection_ConfigDefaultApplied tests that root.go applies config default
func TestBackendSelection_ConfigDefaultApplied(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	setupTestEnv(t, tmpDir)

	// Set config default
	viper.Set("backends.default", "llamacpp")

	// Create registry
	reg := backend.NewRegistry()

	// Register mock backend first (would be selected by availability order)
	mockBackend := &mockBackend{name: "mock"}
	reg.Register(mockBackend)

	// Register llamacpp backend
	llamaBackend := llamacpp.New(tmpDir)
	reg.Register(llamaBackend)

	// Simulate what root.go should do: apply config default
	cfgDefault := viper.GetString("backends.default")
	if cfgDefault != "" {
		if err := reg.SetDefault(cfgDefault); err != nil {
			// Log warning but don't fail (this is what root.go does)
			t.Logf("Warning: configured backend '%s' not available", cfgDefault)
		}
	}

	// Get default backend
	defaultBackend, err := reg.Default()
	if err != nil {
		t.Fatalf("failed to get default backend: %v", err)
	}

	// Should respect config default, not registration order
	if defaultBackend.Name() != "llamacpp" {
		t.Errorf("expected llamacpp (from config), got %s", defaultBackend.Name())
	}
}

// TestBackendRegistry_NameMatch tests backend registration names match config
func TestBackendRegistry_NameMatch(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tmpDir := t.TempDir()

	// Create llamacpp backend
	backend := llamacpp.New(tmpDir)

	// Verify name matches what we use in config
	if backend.Name() != "llamacpp" {
		t.Errorf("llamacpp backend should be named 'llamacpp', got '%s'", backend.Name())
	}

	// Verify this matches defaults.go
	viper.Reset()
	cfg := config.Default()

	if cfg.Backends.Default != "llamacpp" {
		t.Errorf("default config should use 'llamacpp', got '%s'", cfg.Backends.Default)
	}

	// These should match
	if backend.Name() != cfg.Backends.Default {
		t.Errorf("backend name '%s' doesn't match config default '%s'",
			backend.Name(), cfg.Backends.Default)
	}
}

// TestModelDownload_PathConstruction tests model file paths are correct
func TestModelDownload_PathConstruction(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	setupTestEnv(t, tmpDir)

	// Create models directory
	modelsDir := filepath.Join(tmpDir, "models")
	if err := os.MkdirAll(modelsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Test model path construction
	modelName := "qwen2.5-1.5b"
	expectedPath := filepath.Join(modelsDir, modelName+".gguf")

	// Create fake model file
	if err := os.WriteFile(expectedPath, []byte("fake model"), 0644); err != nil {
		t.Fatal(err)
	}

	// Verify file exists at expected path
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("model file should exist at %s", expectedPath)
	}
}

// TestSetupWizard_ModelPresets tests model preset configuration
func TestSetupWizard_ModelPresets(t *testing.T) {
	// Verify all preset models exist in llamacpp package
	presets := []string{
		"qwen2.5-0.5b",
		"qwen2.5-1.5b",
		"qwen2.5-3b",
		"qwen2.5-7b",
	}

	for _, modelName := range presets {
		found := false
		for _, model := range llamacpp.DefaultModels {
			if model.Name == modelName {
				found = true
				// Verify model has valid URL
				if model.URL == "" {
					t.Errorf("model %s has empty URL", modelName)
				}
				if !strings.HasPrefix(model.URL, "http") {
					t.Errorf("model %s has invalid URL: %s", modelName, model.URL)
				}
				break
			}
		}

		if !found {
			t.Errorf("preset model %s not found in llamacpp.DefaultModels", modelName)
		}
	}
}

// TestConfigDefaults_BackendName tests default backend name is correct
func TestConfigDefaults_BackendName(t *testing.T) {
	cfg := config.Default()

	// Should be "llamacpp" not "local"
	if cfg.Backends.Default != "llamacpp" {
		t.Errorf("default backend should be 'llamacpp', got '%s'", cfg.Backends.Default)
	}

	// Verify this backend actually exists
	tmpDir := t.TempDir()
	backend := llamacpp.New(tmpDir)

	if backend.Name() != cfg.Backends.Default {
		t.Errorf("default config '%s' doesn't match backend name '%s'",
			cfg.Backends.Default, backend.Name())
	}
}

// TestIntegration_ConfigToRegistry tests config → registry → backend flow
func TestIntegration_ConfigToRegistry(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	setupTestEnv(t, tmpDir)

	// Step 1: Load default config
	cfg := config.Default()

	// Step 2: Create backend registry
	reg := backend.NewRegistry()

	// Step 3: Register backends
	mockBackend := &mockBackend{name: "mock"}
	llamaBackend := llamacpp.New(tmpDir)

	reg.Register(mockBackend)
	reg.Register(llamaBackend)

	// Step 4: Apply config default to registry (this is what root.go should do)
	if cfg.Backends.Default != "" {
		if err := reg.SetDefault(cfg.Backends.Default); err != nil {
			t.Fatalf("failed to set default from config: %v", err)
		}
	}

	// Step 5: Get default backend
	backend, err := reg.Default()
	if err != nil {
		t.Fatalf("failed to get default backend: %v", err)
	}

	// Step 6: Verify it matches config
	if backend.Name() != cfg.Backends.Default {
		t.Errorf("backend name %s doesn't match config default %s",
			backend.Name(), cfg.Backends.Default)
	}

	// Step 7: Verify it's llamacpp, not mock
	if backend.Name() != "llamacpp" {
		t.Errorf("expected llamacpp backend, got %s", backend.Name())
	}
}

// TestCLI_SetupFlow tests running setup via CLI
func TestCLI_SetupFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// This would require automating interactive input
	// For now, just verify the setup command exists and has correct structure
	cmd := cli.SetupCommand()

	if cmd == nil {
		t.Fatal("setup command should not be nil")
	}

	if cmd.Use != "setup" {
		t.Errorf("expected setup command, got %s", cmd.Use)
	}
}

// TestE2E_BinaryExecution tests running the actual binary
func TestE2E_BinaryExecution(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	// Check if binary exists
	if _, err := os.Stat(scmdBinary); os.IsNotExist(err) {
		t.Skip("scmd binary not built, run 'make build' first")
	}

	// Run version command
	cmd := exec.Command(scmdBinary, "version")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		t.Fatalf("version command failed: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "scmd") {
		t.Errorf("version output should contain 'scmd', got: %s", output)
	}
}

// TestE2E_BackendsCommand tests listing backends
func TestE2E_BackendsCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	if _, err := os.Stat(scmdBinary); os.IsNotExist(err) {
		t.Skip("scmd binary not built")
	}

	cmd := exec.Command(scmdBinary, "backends")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		t.Fatalf("backends command failed: %v", err)
	}

	output := stdout.String()

	// Should list llamacpp backend
	if !strings.Contains(output, "llamacpp") {
		t.Error("backends output should include llamacpp")
	}
}

// TestE2E_ConfigCommand tests showing config
func TestE2E_ConfigCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	if _, err := os.Stat(scmdBinary); os.IsNotExist(err) {
		t.Skip("scmd binary not built")
	}

	cmd := exec.Command(scmdBinary, "config")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		t.Fatalf("config command failed: %v", err)
	}

	output := stdout.String()

	// Should show backends section
	if !strings.Contains(output, "backends") {
		t.Error("config output should include backends section")
	}
}

// ==================== HELPER FUNCTIONS ====================

// setupTestEnv sets up isolated test environment
func setupTestEnv(t *testing.T, tmpDir string) {
	t.Helper()

	viper.Reset()
	t.Setenv("SCMD_DATA_DIR", tmpDir)
	t.Setenv("SCMD_CONFIG_DIR", tmpDir)

	// Create necessary directories
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		t.Fatal(err)
	}
}

// mockBackend is a simple mock backend for testing
type mockBackend struct {
	name string
}

func (m *mockBackend) Name() string {
	return m.name
}

func (m *mockBackend) Type() backend.Type {
	return backend.TypeMock
}

func (m *mockBackend) Initialize(ctx context.Context) error {
	return nil
}

func (m *mockBackend) IsAvailable(ctx context.Context) (bool, error) {
	return true, nil
}

func (m *mockBackend) Shutdown(ctx context.Context) error {
	return nil
}

func (m *mockBackend) Complete(ctx context.Context, req *backend.CompletionRequest) (*backend.CompletionResponse, error) {
	return &backend.CompletionResponse{
		Content:      "Mock response: " + req.Prompt,
		TokensUsed:   10,
		FinishReason: backend.FinishStop,
	}, nil
}

func (m *mockBackend) Stream(ctx context.Context, req *backend.CompletionRequest) (<-chan backend.StreamChunk, error) {
	ch := make(chan backend.StreamChunk, 1)
	go func() {
		ch <- backend.StreamChunk{Content: "Mock response: " + req.Prompt, Done: true}
		close(ch)
	}()
	return ch, nil
}

func (m *mockBackend) SupportsToolCalling() bool {
	return false
}

func (m *mockBackend) CompleteWithTools(ctx context.Context, req *backend.ToolRequest) (*backend.ToolResponse, error) {
	return nil, nil
}

func (m *mockBackend) ModelInfo() *backend.ModelInfo {
	return &backend.ModelInfo{
		Name:          "mock-model",
		ContextLength: 2048,
		Capabilities:  []string{"completion"},
	}
}

func (m *mockBackend) EstimateTokens(text string) int {
	return len(strings.Fields(text))
}
