package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/scmd/scmd/internal/backend/llamacpp"
	"github.com/spf13/cobra"
)

var (
	serverModelFlag   string
	serverContextFlag int
	serverGPUFlag     bool
	serverCPUFlag     bool
	serverTailFlag    int
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Manage llama-server lifecycle",
	Long: `Manage the llama-server process for local inference.

Commands:
  start   - Start llama-server
  stop    - Stop llama-server
  status  - Show server status
  restart - Restart llama-server
  logs    - View server logs

The server is normally auto-started when needed, but these commands
give you manual control for troubleshooting or configuration.`,
}

var serverStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start llama-server",
	Long: `Start the llama-server process for local inference.

Examples:
  scmd server start                    # Start with defaults
  scmd server start -m qwen2.5-3b      # Start with specific model
  scmd server start --cpu              # Start in CPU-only mode
  scmd server start -c 2048            # Start with 2048 context size`,
	RunE: runServerStart,
}

var serverStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop llama-server",
	RunE:  runServerStop,
}

var serverStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show llama-server status",
	RunE:  runServerStatus,
}

var serverRestartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart llama-server",
	RunE:  runServerRestart,
}

var serverLogsCmd = &cobra.Command{
	Use:   "logs",
	Short: "View llama-server logs",
	Long: `View the llama-server log file.

Examples:
  scmd server logs              # View all logs
  scmd server logs --tail 50    # View last 50 lines`,
	RunE: runServerLogs,
}

func init() {
	// Add subcommands
	serverCmd.AddCommand(serverStartCmd)
	serverCmd.AddCommand(serverStopCmd)
	serverCmd.AddCommand(serverStatusCmd)
	serverCmd.AddCommand(serverRestartCmd)
	serverCmd.AddCommand(serverLogsCmd)

	// Flags for start command
	serverStartCmd.Flags().StringVarP(&serverModelFlag, "model", "m", "", "model to use")
	serverStartCmd.Flags().IntVarP(&serverContextFlag, "context", "c", 0, "context size (auto-detected if not specified)")
	serverStartCmd.Flags().BoolVar(&serverGPUFlag, "gpu", false, "force GPU mode")
	serverStartCmd.Flags().BoolVar(&serverCPUFlag, "cpu", false, "force CPU mode")

	// Flags for logs command
	serverLogsCmd.Flags().IntVar(&serverTailFlag, "tail", 100, "number of lines to show")
}

func runServerStart(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	dataDir := getDataDir()

	// Check if server is already running
	if isServerRunning() {
		fmt.Println("✅ llama-server is already running on port 8089")
		return nil
	}

	// Determine model
	modelName := serverModelFlag
	if modelName == "" {
		modelName = llamacpp.GetDefaultModel()
	}

	fmt.Printf("Starting llama-server with model: %s\n", modelName)

	// Get model path
	modelManager := llamacpp.NewModelManager(dataDir)
	modelPath, err := modelManager.GetModelPath(ctx, modelName)
	if err != nil {
		return fmt.Errorf("get model: %w", err)
	}

	// Build configuration
	config := llamacpp.DefaultServerConfig(modelPath)

	// Apply user preferences
	if serverContextFlag > 0 {
		config.ContextSize = serverContextFlag
	}

	if serverCPUFlag {
		config.GPULayers = 0
		fmt.Println("  Mode: CPU only")
	} else if serverGPUFlag {
		config.GPULayers = 99
		fmt.Println("  Mode: Full GPU")
	} else {
		// Auto-detect
		resources, err := llamacpp.DetectSystemResources()
		if err == nil {
			modelInfo, _ := os.Stat(modelPath)
			var modelSize int64
			if modelInfo != nil {
				modelSize = modelInfo.Size()
			}
			optimalConfig := llamacpp.CalculateOptimalConfig(resources, modelSize)
			config.ContextSize = optimalConfig.ContextSize
			config.GPULayers = optimalConfig.GPULayers

			if config.GPULayers > 0 {
				fmt.Printf("  Mode: GPU (%d layers)\n", config.GPULayers)
			} else {
				fmt.Println("  Mode: CPU only")
			}
		}
	}

	fmt.Printf("  Context size: %d\n", config.ContextSize)
	fmt.Println()
	fmt.Println("Starting server... (this may take a few seconds)")

	// Start server
	server, err := llamacpp.StartServerWithConfig(config)
	if err != nil {
		return fmt.Errorf("start server: %w", err)
	}

	fmt.Println("✅ llama-server started successfully")
	fmt.Printf("   Port: %d\n", server.Port())
	fmt.Printf("   Model: %s\n", modelName)
	fmt.Printf("   Logs: %s\n", filepath.Join(dataDir, "logs", "llama-server.log"))
	fmt.Println()
	fmt.Println("Try: echo 'Hello' | scmd /explain")

	return nil
}

func runServerStop(cmd *cobra.Command, args []string) error {
	if !isServerRunning() {
		fmt.Println("ℹ️  llama-server is not running")
		return nil
	}

	fmt.Println("Stopping llama-server...")
	llamacpp.StopServer()

	// Wait a moment and verify it stopped
	time.Sleep(1 * time.Second)

	if !isServerRunning() {
		fmt.Println("✅ llama-server stopped successfully")
	} else {
		fmt.Println("⚠️  Server may still be running. Try: pkill llama-server")
	}

	return nil
}

func runServerStatus(cmd *cobra.Command, args []string) error {
	dataDir := getDataDir()

	fmt.Println("🔍 llama-server Status")
	fmt.Println(strings.Repeat("─", 50))

	if isServerRunning() {
		fmt.Println("Status: ✅ Running")
		fmt.Println("Port:   8089")

		// Try to read PID
		pidPath := filepath.Join(dataDir, "llama-server.pid")
		if pidData, err := os.ReadFile(pidPath); err == nil {
			fmt.Printf("PID:    %s", string(pidData))
		}

		// Check log file size
		logPath := filepath.Join(dataDir, "logs", "llama-server.log")
		if info, err := os.Stat(logPath); err == nil {
			sizeKB := float64(info.Size()) / 1024
			fmt.Printf("Logs:   %.1f KB\n", sizeKB)
		}
	} else {
		fmt.Println("Status: ❌ Not running")
		fmt.Println()
		fmt.Println("Start with: scmd server start")
	}

	return nil
}

func runServerRestart(cmd *cobra.Command, args []string) error {
	fmt.Println("Restarting llama-server...")

	// Stop if running
	if isServerRunning() {
		fmt.Println("  Stopping current server...")
		llamacpp.StopServer()
		time.Sleep(1 * time.Second)
	}

	// Start
	fmt.Println("  Starting server...")
	return runServerStart(cmd, args)
}

func runServerLogs(cmd *cobra.Command, args []string) error {
	dataDir := getDataDir()
	logPath := filepath.Join(dataDir, "logs", "llama-server.log")

	// Check if log file exists
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		fmt.Println("No log file found.")
		fmt.Println("Logs are created when llama-server starts.")
		return nil
	}

	// Read log file
	data, err := os.ReadFile(logPath)
	if err != nil {
		return fmt.Errorf("read log file: %w", err)
	}

	lines := strings.Split(string(data), "\n")

	// Apply tail limit
	startLine := 0
	if serverTailFlag > 0 && len(lines) > serverTailFlag {
		startLine = len(lines) - serverTailFlag
	}

	fmt.Printf("📄 llama-server logs (last %d lines)\n", len(lines)-startLine)
	fmt.Println(strings.Repeat("─", 50))

	for i := startLine; i < len(lines); i++ {
		fmt.Println(lines[i])
	}

	return nil
}

func isServerRunning() bool {
	return llamacpp.IsServerRunning(8089)
}
