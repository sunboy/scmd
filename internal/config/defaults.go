package config

import (
	"path/filepath"
)

// Default returns default configuration
func Default() *Config {
	return &Config{
		Version: "1.0",
		Backends: BackendsConfig{
			Default: "local",
			Local: LocalBackendConfig{
				Model:         "qwen2.5-coder-1.5b",
				ContextLength: 8192,
				GPULayers:     0,
				Threads:       0,
			},
		},
		UI: UIConfig{
			Streaming: true,
			Colors:    true,
			Verbose:   false,
		},
		Models: ModelsConfig{
			Directory:    filepath.Join(DataDir(), "models"),
			AutoDownload: true,
		},
	}
}
