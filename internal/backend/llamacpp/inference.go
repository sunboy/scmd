package llamacpp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/scmd/scmd/internal/backend"
)

// Server wraps llama-server for inference
type Server struct {
	cmd       *exec.Cmd
	port      int
	modelPath string
	contextSize int
	gpuLayers int
	ready     bool
	mu        sync.Mutex
	logFile   *os.File
}

var (
	globalServer *Server
	serverMu     sync.Mutex
)

// ServerConfig holds server configuration
type ServerConfig struct {
	ModelPath   string
	Port        int
	ContextSize int
	GPULayers   int
}

// DefaultServerConfig returns default configuration
func DefaultServerConfig(modelPath string) *ServerConfig {
	return &ServerConfig{
		ModelPath:   modelPath,
		Port:        8089,
		ContextSize: 4096,
		GPULayers:   99, // Auto-detect and use GPU
	}
}

// IsServerRunning checks if a server is already running on the given port
func IsServerRunning(port int) bool {
	url := fmt.Sprintf("http://127.0.0.1:%d/health", port)
	client := &http.Client{Timeout: 1 * time.Second}
	resp, err := client.Get(url)
	if err == nil {
		resp.Body.Close()
		return resp.StatusCode == http.StatusOK
	}
	return false
}

// Port returns the server port
func (s *Server) Port() int {
	return s.port
}

// StartServer starts a llama-server instance
func StartServer(modelPath string, port int) (*Server, error) {
	return StartServerWithConfig(&ServerConfig{
		ModelPath:   modelPath,
		Port:        port,
		ContextSize: 4096,
		GPULayers:   99,
	})
}

// StartServerWithConfig starts llama-server with custom configuration
func StartServerWithConfig(config *ServerConfig) (*Server, error) {
	serverMu.Lock()
	defer serverMu.Unlock()

	debug := os.Getenv("SCMD_DEBUG") != ""

	// Check if already running with same model
	if globalServer != nil && globalServer.modelPath == config.ModelPath && globalServer.ready {
		return globalServer, nil
	}

	// Check if a server is already running on this port (external instance)
	if IsServerRunning(config.Port) {
		// Use the existing external server
		if debug {
			fmt.Fprintf(os.Stderr, "[DEBUG] Using existing llama-server on port %d\n", config.Port)
		}
		server := &Server{
			port:      config.Port,
			modelPath: config.ModelPath,
			ready:     true,
		}
		globalServer = server
		return server, nil
	}

	// Auto-tune configuration based on system resources if not explicitly set
	if config.ContextSize == 0 || config.GPULayers == 0 {
		resources, err := DetectSystemResources()
		if err == nil {
			// Get model size
			modelInfo, err := os.Stat(config.ModelPath)
			var modelSize int64
			if err == nil {
				modelSize = modelInfo.Size()
			} else {
				// Estimate based on common model sizes
				modelSize = 2 * 1024 * 1024 * 1024 // Default 2GB estimate
			}

			// Calculate optimal config
			optimalConfig := CalculateOptimalConfig(resources, modelSize)

			// Use optimal values if not set
			if config.ContextSize == 0 {
				config.ContextSize = optimalConfig.ContextSize
			}
			if config.GPULayers == 0 {
				config.GPULayers = optimalConfig.GPULayers
			}

			if debug {
				fmt.Fprintf(os.Stderr, "[DEBUG] Auto-tuned config: context=%d, gpu_layers=%d\n",
					config.ContextSize, config.GPULayers)
			}
		} else if debug {
			fmt.Fprintf(os.Stderr, "[DEBUG] Could not detect system resources: %v\n", err)
		}
	}

	// Ensure we have sensible defaults
	if config.ContextSize == 0 {
		config.ContextSize = 4096
	}
	if config.GPULayers < 0 {
		config.GPULayers = 99 // Default to full GPU
	}

	// Stop existing server if any
	if globalServer != nil {
		globalServer.Stop()
	}

	// Find llama-server binary
	serverPath, err := findLlamaServer()
	if err != nil {
		return nil, err
	}

	// Create log file
	dataDir := getDataDir()
	logDir := filepath.Join(dataDir, "logs")
	os.MkdirAll(logDir, 0755)
	logPath := filepath.Join(logDir, "llama-server.log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not create log file: %v\n", err)
		logFile = nil
	}

	// Build arguments
	args := []string{
		"-m", config.ModelPath,
		"--port", fmt.Sprintf("%d", config.Port),
		"-c", fmt.Sprintf("%d", config.ContextSize),
		"-ngl", fmt.Sprintf("%d", config.GPULayers),
		"--log-disable", // Disable verbose logging to stdout
	}

	cmd := exec.Command(serverPath, args...)
	if logFile != nil {
		cmd.Stdout = logFile
		cmd.Stderr = logFile
	} else {
		cmd.Stdout = io.Discard
		cmd.Stderr = io.Discard
	}

	if debug {
		fmt.Fprintf(os.Stderr, "[DEBUG] Starting llama-server: %s %v\n", serverPath, args)
	}

	if err := cmd.Start(); err != nil {
		if logFile != nil {
			logFile.Close()
		}
		return nil, fmt.Errorf("start server: %w", err)
	}

	server := &Server{
		cmd:         cmd,
		port:        config.Port,
		modelPath:   config.ModelPath,
		contextSize: config.ContextSize,
		gpuLayers:   config.GPULayers,
		logFile:     logFile,
	}

	// Wait for server to be ready
	if err := server.waitReady(30 * time.Second); err != nil {
		cmd.Process.Kill()
		if logFile != nil {
			logFile.Close()
		}
		return nil, err
	}

	server.ready = true
	globalServer = server

	// Write PID file for management
	pidPath := filepath.Join(dataDir, "llama-server.pid")
	os.WriteFile(pidPath, []byte(fmt.Sprintf("%d", cmd.Process.Pid)), 0644)

	if debug {
		fmt.Fprintf(os.Stderr, "[DEBUG] llama-server started successfully (PID: %d)\n", cmd.Process.Pid)
	}

	return server, nil
}

// getDataDir returns the scmd data directory
func getDataDir() string {
	if dir := os.Getenv("SCMD_DATA_DIR"); dir != "" {
		return dir
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".scmd")
}

// waitReady waits for the server to be ready
func (s *Server) waitReady(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	url := fmt.Sprintf("http://127.0.0.1:%d/health", s.port)

	for time.Now().Before(deadline) {
		resp, err := http.Get(url)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
		time.Sleep(100 * time.Millisecond)
	}

	return fmt.Errorf("server not ready after %v", timeout)
}

// Stop stops the server
func (s *Server) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cmd != nil && s.cmd.Process != nil {
		// Try graceful shutdown first
		s.cmd.Process.Signal(os.Interrupt)

		// Wait up to 5 seconds for graceful shutdown
		done := make(chan error, 1)
		go func() {
			done <- s.cmd.Wait()
		}()

		select {
		case <-done:
			// Graceful shutdown succeeded
		case <-time.After(5 * time.Second):
			// Force kill if not stopped gracefully
			s.cmd.Process.Kill()
			s.cmd.Wait()
		}

		// Clean up PID file
		dataDir := getDataDir()
		pidPath := filepath.Join(dataDir, "llama-server.pid")
		os.Remove(pidPath)
	}

	if s.logFile != nil {
		s.logFile.Close()
		s.logFile = nil
	}

	s.ready = false
	return nil
}

// Complete sends a completion request to the server
func (s *Server) Complete(ctx context.Context, prompt string, req *backend.CompletionRequest) (string, error) {
	debug := os.Getenv("SCMD_DEBUG") != ""
	url := fmt.Sprintf("http://127.0.0.1:%d/completion", s.port)

	// Build request body
	reqBody := map[string]interface{}{
		"prompt":      prompt,
		"n_predict":   req.MaxTokens,
		"temperature": req.Temperature,
		"stop":        []string{"<|im_end|>", "<|endoftext|>"},
		"stream":      false,
	}

	if req.MaxTokens == 0 {
		reqBody["n_predict"] = 2048
	}
	if req.Temperature == 0 {
		reqBody["temperature"] = 0.7
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	if debug {
		fmt.Fprintf(os.Stderr, "[DEBUG] Sending request to %s\n", url)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if debug {
		fmt.Fprintf(os.Stderr, "[DEBUG] Response status: %d\n", resp.StatusCode)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	if debug {
		fmt.Fprintf(os.Stderr, "[DEBUG] Response body length: %d bytes\n", len(respBody))
		if len(respBody) < 1000 {
			fmt.Fprintf(os.Stderr, "[DEBUG] Response body: %s\n", string(respBody))
		}
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("server error (HTTP %d): %s", resp.StatusCode, string(respBody))
	}

	// Parse response - llama-server returns {"content": "...", ...}
	var result struct {
		Content string `json:"content"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("parse response: %w\nRaw: %s", err, string(respBody))
	}

	content := strings.TrimSpace(result.Content)
	if content == "" {
		// Check if there was an error in the response
		var errResult struct {
			Error string `json:"error"`
		}
		if json.Unmarshal(respBody, &errResult) == nil && errResult.Error != "" {
			return "", fmt.Errorf("llama-server: %s", errResult.Error)
		}
		return "", fmt.Errorf("empty response from model.\nPrompt was: %s...\nResponse: %s", truncate(prompt, 100), string(respBody))
	}

	return content, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

// findLlamaServer finds the llama-server binary
func findLlamaServer() (string, error) {
	// Check common locations
	candidates := []string{
		"llama-server",
		"llama.cpp/build/bin/llama-server",
		"/usr/local/bin/llama-server",
		"/opt/llama.cpp/llama-server",
	}

	// Add platform-specific paths
	homeDir, _ := os.UserHomeDir()
	if homeDir != "" {
		candidates = append(candidates,
			filepath.Join(homeDir, ".local", "bin", "llama-server"),
			filepath.Join(homeDir, "llama.cpp", "build", "bin", "llama-server"),
		)
	}

	// Check bundled binary
	execPath, _ := os.Executable()
	if execPath != "" {
		bundledPath := filepath.Join(filepath.Dir(execPath), "llama-server")
		if runtime.GOOS == "windows" {
			bundledPath += ".exe"
		}
		candidates = append([]string{bundledPath}, candidates...)
	}

	for _, path := range candidates {
		if _, err := exec.LookPath(path); err == nil {
			return path, nil
		}
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("llama-server not found. Install with: make install-llamacpp")
}

// runServerInference uses llama-server for inference
func (b *Backend) runServerInference(ctx context.Context, prompt string, req *backend.CompletionRequest) (string, error) {
	// Use existing server URL
	url := b.serverURL + "/completion"

	reqBody := map[string]interface{}{
		"prompt":      prompt,
		"n_predict":   req.MaxTokens,
		"temperature": req.Temperature,
		"stop":        []string{"<|im_end|>", "<|endoftext|>"},
		"stream":      false,
	}

	if req.MaxTokens == 0 {
		reqBody["n_predict"] = 2048
	}
	if req.Temperature == 0 {
		reqBody["temperature"] = 0.7
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", ParseError(err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return "", ParseError(err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return "", ParseError(err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", ParseError(fmt.Errorf("read response: %w", err))
	}

	if resp.StatusCode != http.StatusOK {
		return "", ParseError(fmt.Errorf("server error (HTTP %d): %s", resp.StatusCode, string(respBody)))
	}

	var result struct {
		Content string `json:"content"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", ParseError(fmt.Errorf("parse response: %w", err))
	}

	content := strings.TrimSpace(result.Content)
	if content == "" {
		return "", ParseError(fmt.Errorf("empty response from server"))
	}

	return content, nil
}

// runCGOInference uses CGO bindings for direct inference
// This requires the go-llama.cpp library to be properly linked
func (b *Backend) runCGOInference(ctx context.Context, prompt string, req *backend.CompletionRequest) (string, error) {
	// Start a server if not running
	server, err := StartServer(b.modelPath, 8089)
	if err != nil {
		return "", ParseError(err)
	}

	result, err := server.Complete(ctx, prompt, req)
	if err != nil {
		return "", ParseError(err)
	}

	return result, nil
}

// SetServerURL sets the URL of an external llama-server
func (b *Backend) SetServerURL(url string) {
	b.serverURL = url
}

// StopServer stops the global inference server
func StopServer() {
	serverMu.Lock()
	defer serverMu.Unlock()
	if globalServer != nil {
		globalServer.Stop()
		globalServer = nil
	}
}
