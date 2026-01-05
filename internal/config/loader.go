package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Load loads configuration from file and environment
func Load() (*Config, error) {
	v := viper.New()

	// Set defaults
	defaults := Default()
	v.SetDefault("version", defaults.Version)
	v.SetDefault("backends.default", defaults.Backends.Default)
	v.SetDefault("backends.local.model", defaults.Backends.Local.Model)
	v.SetDefault("backends.local.context_length", defaults.Backends.Local.ContextLength)
	v.SetDefault("backends.local.gpu_layers", defaults.Backends.Local.GPULayers)
	v.SetDefault("backends.local.threads", defaults.Backends.Local.Threads)
	v.SetDefault("ui.streaming", defaults.UI.Streaming)
	v.SetDefault("ui.colors", defaults.UI.Colors)
	v.SetDefault("ui.verbose", defaults.UI.Verbose)
	v.SetDefault("models.directory", defaults.Models.Directory)
	v.SetDefault("models.auto_download", defaults.Models.AutoDownload)

	// Config file
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(DataDir())

	// Environment variables
	v.SetEnvPrefix("SCMD")
	v.AutomaticEnv()

	// Read config
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
		// Config not found is OK, use defaults
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Save saves configuration to file
func Save(cfg *Config) error {
	dir := DataDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	v := viper.New()
	v.Set("version", cfg.Version)
	v.Set("backends", cfg.Backends)
	v.Set("ui", cfg.UI)
	v.Set("models", cfg.Models)

	return v.WriteConfigAs(filepath.Join(dir, "config.yaml"))
}

// EnsureDataDir creates the data directory if it doesn't exist
func EnsureDataDir() error {
	return os.MkdirAll(DataDir(), 0755)
}
