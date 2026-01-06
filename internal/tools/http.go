package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/scmd/scmd/internal/backend"
)

// HTTPGetTool fetches data from URLs
type HTTPGetTool struct {
	client *http.Client
}

// NewHTTPGetTool creates a new HTTP GET tool
func NewHTTPGetTool() *HTTPGetTool {
	return &HTTPGetTool{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Name returns the tool name
func (t *HTTPGetTool) Name() string {
	return "http_get"
}

// Description returns the tool description
func (t *HTTPGetTool) Description() string {
	return "Fetch data from a URL using HTTP GET request"
}

// Parameters returns the parameter schema
func (t *HTTPGetTool) Parameters() map[string]backend.ToolParameter {
	return map[string]backend.ToolParameter{
		"url": {
			Type:        "string",
			Description: "The URL to fetch",
			Required:    true,
		},
		"max_size": {
			Type:        "number",
			Description: "Maximum response size in bytes (default: 1MB, max: 10MB)",
			Required:    false,
		},
	}
}

// RequiresConfirmation returns false for GET requests
func (t *HTTPGetTool) RequiresConfirmation() bool {
	return false
}

// Execute fetches a URL
func (t *HTTPGetTool) Execute(ctx context.Context, params map[string]interface{}) (*Result, error) {
	urlStr, ok := params["url"].(string)
	if !ok || urlStr == "" {
		return &Result{
			Success: false,
			Error:   "url parameter is required",
		}, nil
	}

	// Validate URL scheme
	if !strings.HasPrefix(urlStr, "http://") && !strings.HasPrefix(urlStr, "https://") {
		return &Result{
			Success: false,
			Error:   "only http:// and https:// URLs are supported",
		}, nil
	}

	// Get max size
	maxSize := int64(1024 * 1024) // 1MB default
	if size, ok := params["max_size"].(float64); ok {
		maxSize = int64(size)
		if maxSize > 10*1024*1024 { // Max 10MB
			maxSize = 10 * 1024 * 1024
		}
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
	if err != nil {
		return &Result{
			Success: false,
			Error:   fmt.Sprintf("failed to create request: %v", err),
		}, nil
	}

	// Set user agent
	req.Header.Set("User-Agent", "scmd/1.0")

	// Execute request
	resp, err := t.client.Do(req)
	if err != nil {
		return &Result{
			Success: false,
			Error:   fmt.Sprintf("request failed: %v", err),
		}, nil
	}
	defer resp.Body.Close()

	// Check status
	if resp.StatusCode != http.StatusOK {
		return &Result{
			Success: false,
			Error:   fmt.Sprintf("HTTP %d: %s", resp.StatusCode, resp.Status),
		}, nil
	}

	// Read body with size limit
	limitedReader := io.LimitReader(resp.Body, maxSize)
	body, err := io.ReadAll(limitedReader)
	if err != nil {
		return &Result{
			Success: false,
			Error:   fmt.Sprintf("failed to read response: %v", err),
		}, nil
	}

	// Check if response was truncated
	truncated := ""
	if int64(len(body)) == maxSize {
		truncated = fmt.Sprintf("\n\n[Response truncated at %d bytes]", maxSize)
	}

	return &Result{
		Success: true,
		Output:  string(body) + truncated,
	}, nil
}
