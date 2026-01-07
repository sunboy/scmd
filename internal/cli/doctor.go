package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"github.com/scmd/scmd/internal/backend/llamacpp"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check scmd installation and diagnose issues",
	Long: `Run diagnostics to check if scmd is properly configured.

This command checks:
  - scmd binary version
  - Models directory and downloaded models
  - llama-server binary availability
  - llama-server status (running/stopped)
  - System resources (memory, disk space)
  - Backend connectivity
  - Configuration validity

Use this to troubleshoot issues before running inference.`,
	RunE: runDoctor,
}

func init() {
	// This will be added to rootCmd in root.go
}

func runDoctor(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	fmt.Println("🏥 scmd Health Check")
	fmt.Println(strings.Repeat("═", 60))
	fmt.Println()

	allOK := true

	// 1. Check scmd binary
	if ok := checkScmdBinary(); !ok {
		allOK = false
	}

	// 2. Check data directory
	dataDir := getDataDir()
	if ok := checkDataDirectory(dataDir); !ok {
		allOK = false
	}

	// 3. Check models
	if ok := checkModels(ctx, dataDir); !ok {
		allOK = false
	}

	// 4. Check llama-server binary
	llamaServerPath, ok := checkLlamaServerBinary()
	if !ok {
		allOK = false
	}

	// 5. Check llama-server status
	if ok := checkLlamaServerStatus(); !ok {
		// Not critical if server isn't running (we auto-start)
		// allOK = false
	}

	// 6. Check system resources
	if ok := checkSystemResources(); !ok {
		// Warnings only, not failures
	}

	// 7. Check port availability
	if ok := checkPortAvailability(8089); !ok {
		// Not critical if server is running
	}

	// 8. Test backend connectivity
	if llamaServerPath != "" {
		if ok := checkBackendConnectivity(ctx, dataDir); !ok {
			// Not critical - might just not be started yet
		}
	}

	// 9. Check disk space
	if ok := checkDiskSpace(dataDir); !ok {
		// Warning only
	}

	fmt.Println()
	fmt.Println(strings.Repeat("═", 60))
	if allOK {
		fmt.Println("✅ All checks passed! scmd is ready to use.")
		fmt.Println()
		fmt.Println("Try running: echo 'Hello world' | scmd /explain")
	} else {
		fmt.Println("⚠️  Some issues found. See recommendations above.")
		fmt.Println()
		fmt.Println("For help, visit: https://github.com/scmd/scmd/blob/main/docs/troubleshooting.md")
	}

	return nil
}

func checkScmdBinary() bool {
	exe, err := os.Executable()
	if err != nil {
		printCheck("scmd binary", false, "Could not determine executable path")
		return false
	}

	info, err := os.Stat(exe)
	if err != nil {
		printCheck("scmd binary", false, fmt.Sprintf("Error: %v", err))
		return false
	}

	size := float64(info.Size()) / 1024 / 1024
	printCheck("scmd binary", true, fmt.Sprintf("%s (%.1f MB)", exe, size))
	return true
}

func checkDataDirectory(dataDir string) bool {
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		printCheck("Data directory", false, fmt.Sprintf("%s does not exist", dataDir))
		printRecommendation("Run any scmd command to create it automatically")
		return false
	}

	printCheck("Data directory", true, dataDir)
	return true
}

func checkModels(ctx context.Context, dataDir string) bool {
	modelsDir := filepath.Join(dataDir, "models")

	if _, err := os.Stat(modelsDir); os.IsNotExist(err) {
		printCheck("Models directory", false, "No models downloaded yet")
		printRecommendation("Models will download automatically on first use")
		printRecommendation("Or download manually: scmd models list")
		return false
	}

	entries, err := os.ReadDir(modelsDir)
	if err != nil {
		printCheck("Models directory", false, fmt.Sprintf("Error reading: %v", err))
		return false
	}

	modelFiles := []string{}
	var totalSize int64
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".gguf") {
			modelFiles = append(modelFiles, e.Name())
			info, _ := e.Info()
			if info != nil {
				totalSize += info.Size()
			}
		}
	}

	if len(modelFiles) == 0 {
		printCheck("Downloaded models", false, "No models found")
		printRecommendation("Download a model: scmd models download qwen3-4b")
		return false
	}

	sizeGB := float64(totalSize) / 1024 / 1024 / 1024
	printCheck("Downloaded models", true, fmt.Sprintf("%d model(s), %.1f GB total", len(modelFiles), sizeGB))
	for _, model := range modelFiles {
		fmt.Printf("  - %s\n", model)
	}
	return true
}

func checkLlamaServerBinary() (string, bool) {
	// Try to find llama-server
	candidates := []string{
		"llama-server",
		"/usr/local/bin/llama-server",
		"/opt/homebrew/bin/llama-server",
	}

	homeDir, _ := os.UserHomeDir()
	if homeDir != "" {
		candidates = append(candidates, filepath.Join(homeDir, ".local", "bin", "llama-server"))
	}

	for _, path := range candidates {
		if fullPath, err := exec.LookPath(path); err == nil {
			printCheck("llama-server binary", true, fullPath)
			return fullPath, true
		}
	}

	printCheck("llama-server binary", false, "Not found in PATH")
	printRecommendation("Install llama.cpp:")
	printRecommendation("  macOS: brew install llama.cpp")
	printRecommendation("  Linux: Build from https://github.com/ggerganov/llama.cpp")
	printRecommendation("Or use cloud provider: export OPENAI_API_KEY=your-key")
	return "", false
}

func checkLlamaServerStatus() bool {
	if llamacpp.IsServerRunning(8089) {
		printCheck("llama-server status", true, "Running on port 8089")
		return true
	}

	printCheck("llama-server status", false, "Not running")
	printRecommendation("scmd will auto-start the server when needed")
	printRecommendation("Or start manually: scmd server start")
	return false
}

func checkSystemResources() bool {
	// Get system memory (platform-specific)
	var totalRAM int64
	var available bool

	switch runtime.GOOS {
	case "darwin":
		// macOS
		out, err := exec.Command("sysctl", "-n", "hw.memsize").Output()
		if err == nil {
			fmt.Sscanf(string(out), "%d", &totalRAM)
			available = true
		}
	case "linux":
		// Linux
		out, err := exec.Command("grep", "MemTotal", "/proc/meminfo").Output()
		if err == nil {
			var kb int64
			fmt.Sscanf(string(out), "MemTotal: %d kB", &kb)
			totalRAM = kb * 1024
			available = true
		}
	}

	if !available {
		printCheck("System memory", false, "Could not determine")
		return false
	}

	ramGB := float64(totalRAM) / 1024 / 1024 / 1024
	printCheck("System memory", true, fmt.Sprintf("%.1f GB total", ramGB))

	// Provide recommendations based on RAM
	if ramGB < 4 {
		printRecommendation("Low memory detected. Use qwen2.5-0.5b model")
		printRecommendation("Or use cloud provider for better performance")
		return false
	} else if ramGB < 8 {
		printRecommendation("Limited memory. Recommended: qwen2.5-1.5b or qwen2.5-3b")
	} else if ramGB >= 16 {
		fmt.Println("  💡 You have enough RAM for larger models (qwen2.5-7b)")
	}

	return true
}

func checkPortAvailability(port int) bool {
	if llamacpp.IsServerRunning(port) {
		// Port is in use
		printCheck(fmt.Sprintf("Port %d availability", port), false, "Already in use (server running)")
		return false
	}

	printCheck(fmt.Sprintf("Port %d availability", port), true, "Available")
	return true
}

func checkBackendConnectivity(ctx context.Context, dataDir string) bool {
	// Try to initialize the llamacpp backend
	backend := llamacpp.New(dataDir)

	// Check if available
	available, err := backend.IsAvailable(ctx)
	if err != nil {
		printCheck("Backend connectivity", false, fmt.Sprintf("Error: %v", err))
		return false
	}

	if !available {
		printCheck("Backend connectivity", false, "Backend not available")
		printRecommendation("Check llama-server installation")
		return false
	}

	printCheck("Backend connectivity", true, "llamacpp backend available")
	return true
}

func checkDiskSpace(dataDir string) bool {
	// Get available disk space (platform-specific)
	var availableGB float64
	var success bool

	switch runtime.GOOS {
	case "darwin", "linux":
		out, err := exec.Command("df", "-k", dataDir).Output()
		if err == nil {
			lines := strings.Split(string(out), "\n")
			if len(lines) >= 2 {
				fields := strings.Fields(lines[1])
				if len(fields) >= 4 {
					var kb int64
					fmt.Sscanf(fields[3], "%d", &kb)
					availableGB = float64(kb) / 1024 / 1024
					success = true
				}
			}
		}
	}

	if !success {
		printCheck("Disk space", false, "Could not determine")
		return false
	}

	printCheck("Disk space", true, fmt.Sprintf("%.1f GB available", availableGB))

	if availableGB < 5 {
		printRecommendation("Low disk space. Models require 1-5 GB each")
		return false
	}

	return true
}

// Helper functions

func printCheck(item string, ok bool, message string) {
	status := "✅"
	if !ok {
		status = "❌"
	}
	fmt.Printf("%s %-25s %s\n", status, item+":", message)
}

func printRecommendation(text string) {
	fmt.Printf("   💡 %s\n", text)
}
