package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDataDir(t *testing.T) {
	result := DataDir()
	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".scmd")
	assert.Equal(t, expected, result)
}

func TestConfigPath(t *testing.T) {
	result := ConfigPath()
	assert.Contains(t, result, ".scmd")
	assert.Contains(t, result, "config.yaml")
}

func TestDefault(t *testing.T) {
	cfg := Default()

	assert.Equal(t, "1.0", cfg.Version)
	assert.Equal(t, "local", cfg.Backends.Default)
	assert.Equal(t, "qwen2.5-coder-1.5b", cfg.Backends.Local.Model)
	assert.Equal(t, 8192, cfg.Backends.Local.ContextLength)
	assert.True(t, cfg.UI.Streaming)
	assert.True(t, cfg.UI.Colors)
	assert.False(t, cfg.UI.Verbose)
	assert.True(t, cfg.Models.AutoDownload)
}

func TestConfig_GetString(t *testing.T) {
	cfg := Default()

	assert.Equal(t, "local", cfg.GetString("backends.default"))
	assert.Equal(t, "qwen2.5-coder-1.5b", cfg.GetString("backends.local.model"))
	assert.Equal(t, "", cfg.GetString("nonexistent"))
}

func TestConfig_GetBool(t *testing.T) {
	cfg := Default()

	assert.True(t, cfg.GetBool("ui.streaming"))
	assert.True(t, cfg.GetBool("ui.colors"))
	assert.False(t, cfg.GetBool("ui.verbose"))
}

func TestConfig_GetInt(t *testing.T) {
	cfg := Default()

	assert.Equal(t, 8192, cfg.GetInt("backends.local.context_length"))
	assert.Equal(t, 0, cfg.GetInt("backends.local.gpu_layers"))
	assert.Equal(t, 0, cfg.GetInt("nonexistent"))
}

func TestLoad(t *testing.T) {
	// Load should succeed even without a config file
	cfg, err := Load()
	assert.NoError(t, err)
	assert.NotNil(t, cfg)

	// Should have default values
	assert.Equal(t, "local", cfg.Backends.Default)
}

func TestEnsureDataDir(t *testing.T) {
	err := EnsureDataDir()
	assert.NoError(t, err)

	// Directory should exist
	_, err = os.Stat(DataDir())
	assert.NoError(t, err)
}
