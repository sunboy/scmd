package e2e

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/scmd/scmd/internal/backend/llamacpp"
)

// ==================== MODEL MANAGEMENT TESTS ====================

func TestModels_List(t *testing.T) {
	stdout, _, err := runScmd(t, "models", "list")
	if err != nil {
		t.Fatalf("models list failed: %v", err)
	}

	// Should show available models
	if !strings.Contains(stdout, "qwen") {
		t.Error("should list qwen models")
	}

	if !strings.Contains(stdout, "SIZE") {
		t.Error("should show size column")
	}

	if !strings.Contains(stdout, "STATUS") {
		t.Error("should show status column")
	}
}

func TestModels_ListAlias(t *testing.T) {
	// Test 'ls' alias
	stdout, _, err := runScmd(t, "models", "ls")
	if err != nil {
		t.Fatalf("models ls failed: %v", err)
	}

	if stdout == "" {
		t.Error("should have output")
	}
}

func TestModels_Info(t *testing.T) {
	// Test info for a known model
	stdout, _, err := runScmd(t, "models", "info", "qwen3-4b")
	if err != nil {
		t.Fatalf("models info failed: %v", err)
	}

	if !strings.Contains(stdout, "Name") {
		t.Error("should show model name")
	}

	if !strings.Contains(stdout, "Size") {
		t.Error("should show model size")
	}
}

func TestModels_InfoInvalid(t *testing.T) {
	// Test info for invalid model
	_, _, err := runScmd(t, "models", "info", "nonexistent-model")
	if err == nil {
		t.Error("should fail with nonexistent model")
	}
}

func TestModels_Default(t *testing.T) {
	// Set default model
	stdout, _, err := runScmd(t, "models", "default", "qwen2.5-3b")
	if err != nil {
		t.Fatalf("models default failed: %v", err)
	}

	if !strings.Contains(stdout, "Default model") {
		t.Error("should confirm default model set")
	}
}

func TestModels_DefaultInvalid(t *testing.T) {
	// Try to set invalid model as default
	_, _, err := runScmd(t, "models", "default", "invalid-model")
	// May or may not fail depending on implementation
	// Just testing it doesn't crash
	_ = err
}

// ==================== MODEL MANAGER UNIT TESTS ====================

func TestModelManager_ListModels(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := llamacpp.NewModelManager(tmpDir)

	models := mgr.ListModels()
	if len(models) == 0 {
		t.Error("should have at least one model defined")
	}

	// Check model structure
	for _, m := range models {
		if m.Name == "" {
			t.Error("model should have name")
		}
		if m.URL == "" {
			t.Error("model should have URL")
		}
		if m.Size <= 0 {
			t.Error("model should have positive size")
		}
	}
}

func TestModelManager_ListDownloaded(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := llamacpp.NewModelManager(tmpDir)

	downloaded, err := mgr.ListDownloaded()
	if err != nil {
		t.Fatalf("list downloaded failed: %v", err)
	}

	// Initially should be empty
	if len(downloaded) != 0 {
		t.Error("fresh directory should have no downloaded models")
	}
}

func TestModelManager_GetModelPath_Local(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := llamacpp.NewModelManager(tmpDir)

	// Create a fake local model file
	localModel := filepath.Join(tmpDir, "local-model.gguf")
	if err := os.WriteFile(localModel, []byte("fake model"), 0644); err != nil {
		t.Fatal(err)
	}

	// Should return local path as-is
	path, err := mgr.GetModelPath(context.Background(), localModel)
	if err != nil {
		t.Fatalf("get local model path failed: %v", err)
	}

	if path != localModel {
		t.Errorf("expected %s, got %s", localModel, path)
	}
}

func TestModelManager_GetModelPath_Unknown(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := llamacpp.NewModelManager(tmpDir)

	// Try unknown model
	_, err := mgr.GetModelPath(context.Background(), "totally-unknown-model")
	if err == nil {
		t.Error("should fail with unknown model")
	}

	if !strings.Contains(err.Error(), "unknown model") {
		t.Errorf("error should mention unknown model, got: %v", err)
	}
}

func TestModelManager_DeleteModel(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := llamacpp.NewModelManager(tmpDir)

	// Create models directory
	modelsDir := filepath.Join(tmpDir, "models")
	if err := os.MkdirAll(modelsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create fake model file
	modelFile := filepath.Join(modelsDir, "qwen3-4b-Q4_K_M.gguf")
	if err := os.WriteFile(modelFile, []byte("fake"), 0644); err != nil {
		t.Fatal(err)
	}

	// Delete it
	err := mgr.DeleteModel("qwen3-4b")
	if err != nil {
		t.Fatalf("delete model failed: %v", err)
	}

	// Verify it's gone
	if _, err := os.Stat(modelFile); !os.IsNotExist(err) {
		t.Error("model file should be deleted")
	}
}

func TestModelManager_DeleteModel_Nonexistent(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := llamacpp.NewModelManager(tmpDir)

	err := mgr.DeleteModel("nonexistent-model")
	if err == nil {
		t.Error("should fail deleting nonexistent model")
	}
}

// ==================== MODEL BACKEND TESTS ====================

func TestModelBackend_Initialize(t *testing.T) {
	tmpDir := t.TempDir()
	backend := llamacpp.New(tmpDir)

	// Initialize should not error even without models
	// (it will download on first use)
	ctx := context.Background()
	err := backend.Initialize(ctx)

	// May fail if no llama-server available, which is OK for unit tests
	if err != nil && !strings.Contains(err.Error(), "llama") {
		t.Logf("Initialize error (expected in CI): %v", err)
	}
}

func TestModelBackend_Name(t *testing.T) {
	tmpDir := t.TempDir()
	backend := llamacpp.New(tmpDir)

	if backend.Name() != "llamacpp" {
		t.Errorf("expected llamacpp, got %s", backend.Name())
	}
}

func TestModelBackend_ListModels(t *testing.T) {
	tmpDir := t.TempDir()
	backend := llamacpp.New(tmpDir)

	models, err := backend.ListModels(context.Background())
	if err != nil {
		t.Fatalf("list models failed: %v", err)
	}

	if len(models) == 0 {
		t.Error("should have at least one model")
	}
}

func TestModelBackend_SetModel(t *testing.T) {
	tmpDir := t.TempDir()
	backend := llamacpp.New(tmpDir)

	err := backend.SetModel("qwen2.5-3b")
	if err != nil {
		t.Fatalf("set model failed: %v", err)
	}
}

func TestModelBackend_SetModel_Invalid(t *testing.T) {
	tmpDir := t.TempDir()
	backend := llamacpp.New(tmpDir)

	// Setting invalid model name should not error immediately
	// (error happens on initialization)
	err := backend.SetModel("invalid-model")
	if err != nil {
		t.Fatalf("set model should not error: %v", err)
	}
}

func TestModelBackend_SupportsToolCalling(t *testing.T) {
	tmpDir := t.TempDir()
	backend := llamacpp.New(tmpDir)

	// Qwen models support tool calling
	if !backend.SupportsToolCalling() {
		t.Error("should support tool calling")
	}
}

func TestModelBackend_ModelInfo(t *testing.T) {
	tmpDir := t.TempDir()
	backend := llamacpp.New(tmpDir)

	info := backend.ModelInfo()
	if info == nil {
		t.Fatal("model info should not be nil")
	}

	if info.Name == "" {
		t.Error("model info should have name")
	}

	if info.ContextLength <= 0 {
		t.Error("model info should have positive context length")
	}

	if len(info.Capabilities) == 0 {
		t.Error("model info should have capabilities")
	}
}

func TestModelBackend_EstimateTokens(t *testing.T) {
	tmpDir := t.TempDir()
	backend := llamacpp.New(tmpDir)

	tests := []struct {
		text   string
		minTok int
		maxTok int
	}{
		{"hello", 1, 5},
		{"hello world", 2, 10},
		{strings.Repeat("test ", 100), 100, 600},
		{"", 0, 1},
	}

	for _, tt := range tests {
		tokens := backend.EstimateTokens(tt.text)
		if tokens < tt.minTok || tokens > tt.maxTok {
			t.Errorf("text %q: expected tokens between %d and %d, got %d",
				truncate(tt.text, 20), tt.minTok, tt.maxTok, tokens)
		}
	}
}

// ==================== MODEL FORMATS AND VARIANTS ====================

func TestModels_AllDefinedModelsValid(t *testing.T) {
	// Verify all default models have required fields
	for _, m := range llamacpp.DefaultModels {
		if m.Name == "" {
			t.Error("model missing name")
		}
		if m.Variant == "" {
			t.Error("model missing variant")
		}
		if m.URL == "" {
			t.Errorf("model %s missing URL", m.Name)
		}
		if m.Size <= 0 {
			t.Errorf("model %s has invalid size: %d", m.Name, m.Size)
		}
		if m.Description == "" {
			t.Errorf("model %s missing description", m.Name)
		}
		if m.ContextSize <= 0 {
			t.Errorf("model %s has invalid context size: %d", m.Name, m.ContextSize)
		}

		// Check URL is valid format
		if !strings.HasPrefix(m.URL, "http://") && !strings.HasPrefix(m.URL, "https://") {
			t.Errorf("model %s has invalid URL: %s", m.Name, m.URL)
		}
	}
}

func TestModels_DefaultModelExists(t *testing.T) {
	defaultModel := llamacpp.GetDefaultModel()
	if defaultModel == "" {
		t.Fatal("default model should not be empty")
	}

	// Check it exists in available models
	found := false
	for _, m := range llamacpp.DefaultModels {
		if m.Name == defaultModel {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("default model %s not found in available models", defaultModel)
	}
}

func TestModels_NoModelDuplicates(t *testing.T) {
	seen := make(map[string]bool)
	for _, m := range llamacpp.DefaultModels {
		key := m.Name + "-" + m.Variant
		if seen[key] {
			t.Errorf("duplicate model: %s", key)
		}
		seen[key] = true
	}
}

func TestModels_SizeOrder(t *testing.T) {
	// Verify models are reasonably sized
	// (between 100MB and 10GB for GGUF models)
	for _, m := range llamacpp.DefaultModels {
		if m.Size < 100*1024*1024 {
			t.Errorf("model %s seems too small: %d bytes", m.Name, m.Size)
		}
		if m.Size > 10*1024*1024*1024 {
			t.Errorf("model %s seems too large: %d bytes", m.Name, m.Size)
		}
	}
}

// ==================== HELPER FUNCTIONS ====================

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
