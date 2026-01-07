package llamacpp

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

// SystemResources holds information about system resources
type SystemResources struct {
	TotalRAMBytes int64
	AvailableRAMBytes int64
	HasGPU bool
	GPUType string
}

// DetectSystemResources detects available system resources
func DetectSystemResources() (*SystemResources, error) {
	res := &SystemResources{}

	// Detect total RAM
	totalRAM, err := getTotalRAM()
	if err != nil {
		return nil, fmt.Errorf("detect RAM: %w", err)
	}
	res.TotalRAMBytes = totalRAM

	// Estimate available RAM (conservative: 80% of total)
	res.AvailableRAMBytes = int64(float64(totalRAM) * 0.8)

	// Detect GPU
	res.HasGPU, res.GPUType = detectGPU()

	return res, nil
}

// getTotalRAM returns total system RAM in bytes
func getTotalRAM() (int64, error) {
	switch runtime.GOOS {
	case "darwin":
		// macOS
		out, err := exec.Command("sysctl", "-n", "hw.memsize").Output()
		if err != nil {
			return 0, err
		}
		bytes, err := strconv.ParseInt(strings.TrimSpace(string(out)), 10, 64)
		if err != nil {
			return 0, err
		}
		return bytes, nil

	case "linux":
		// Linux - read from /proc/meminfo
		out, err := exec.Command("grep", "MemTotal", "/proc/meminfo").Output()
		if err != nil {
			return 0, err
		}
		var kb int64
		_, err = fmt.Sscanf(string(out), "MemTotal: %d kB", &kb)
		if err != nil {
			return 0, err
		}
		return kb * 1024, nil

	default:
		return 0, fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
}

// detectGPU detects if GPU is available and what type
func detectGPU() (bool, string) {
	switch runtime.GOOS {
	case "darwin":
		// macOS - check for Metal support
		// All modern Macs (M1/M2/M3 or Intel with discrete GPU) support Metal
		out, err := exec.Command("system_profiler", "SPDisplaysDataType").Output()
		if err == nil && strings.Contains(string(out), "Metal") {
			// Detect Apple Silicon
			arch := runtime.GOARCH
			if arch == "arm64" {
				// M1/M2/M3
				return true, "Apple Silicon (Metal)"
			}
			return true, "Metal"
		}
		return false, ""

	case "linux":
		// Check for NVIDIA GPU
		if _, err := exec.LookPath("nvidia-smi"); err == nil {
			return true, "NVIDIA"
		}
		// Check for AMD GPU
		if _, err := exec.Command("lspci").Output(); err == nil {
			return true, "AMD"
		}
		return false, ""

	default:
		return false, ""
	}
}

// CalculateOptimalConfig calculates optimal server configuration based on resources
func CalculateOptimalConfig(res *SystemResources, modelSizeBytes int64) *ServerConfig {
	config := &ServerConfig{
		Port: 8089,
	}

	debug := os.Getenv("SCMD_DEBUG") != ""

	// Calculate available memory for inference (total - 2GB for system)
	systemReserveBytes := int64(2 * 1024 * 1024 * 1024) // 2GB
	availableForModel := res.AvailableRAMBytes - systemReserveBytes

	if debug {
		fmt.Fprintf(os.Stderr, "[DEBUG] Total RAM: %.1f GB\n", float64(res.TotalRAMBytes)/1024/1024/1024)
		fmt.Fprintf(os.Stderr, "[DEBUG] Available for model: %.1f GB\n", float64(availableForModel)/1024/1024/1024)
		fmt.Fprintf(os.Stderr, "[DEBUG] Model size: %.1f GB\n", float64(modelSizeBytes)/1024/1024/1024)
	}

	// Calculate context size
	// Each token uses approximately 4 bytes
	// Formula: available_memory = model_size + (context_size * 4 * num_layers)
	// Simplified: available_memory = model_size + (context_size * 16) for typical models

	memoryAfterModel := availableForModel - modelSizeBytes
	if memoryAfterModel < 0 {
		// Model too large for available RAM - use minimum context
		config.ContextSize = 512
		config.GPULayers = 0 // Force CPU mode
		if debug {
			fmt.Fprintf(os.Stderr, "[DEBUG] Model too large, using minimal config\n")
		}
		return config
	}

	// Calculate safe context size
	// Reserve some memory for context processing
	contextMemoryBytes := memoryAfterModel / 2 // Use half of remaining memory
	estimatedContextSize := int(contextMemoryBytes / 16) // ~16 bytes per token with overhead

	// Clamp to reasonable values
	if estimatedContextSize < 512 {
		config.ContextSize = 512
	} else if estimatedContextSize > 8192 {
		config.ContextSize = 8192
	} else {
		// Round down to nearest power of 2 for efficiency
		config.ContextSize = roundDownToPowerOf2(estimatedContextSize)
	}

	// Determine GPU usage
	if res.HasGPU {
		// Check if we have enough memory for GPU mode
		ramGB := float64(res.TotalRAMBytes) / 1024 / 1024 / 1024
		modelGB := float64(modelSizeBytes) / 1024 / 1024 / 1024

		if ramGB >= 16 {
			// Plenty of RAM - use full GPU
			config.GPULayers = 99
		} else if ramGB >= 8 {
			// Moderate RAM - use GPU but be conservative
			if modelGB <= 2.5 {
				config.GPULayers = 99 // Small model, full GPU
			} else {
				config.GPULayers = 32 // Larger model, partial offload
			}
		} else {
			// Low RAM - use CPU mode to be safe
			config.GPULayers = 0
			if debug {
				fmt.Fprintf(os.Stderr, "[DEBUG] Low RAM detected, using CPU mode\n")
			}
		}
	} else {
		// No GPU available
		config.GPULayers = 0
	}

	if debug {
		fmt.Fprintf(os.Stderr, "[DEBUG] Calculated config: context=%d, gpu_layers=%d\n",
			config.ContextSize, config.GPULayers)
	}

	return config
}

// roundDownToPowerOf2 rounds down to the nearest power of 2
func roundDownToPowerOf2(n int) int {
	if n <= 0 {
		return 1
	}

	power := 1
	for power*2 <= n {
		power *= 2
	}
	return power
}

// FormatBytes formats bytes as human-readable string
func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
