//go:build !windows
// +build !windows

package llamacpp

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

// DownloadConfig contains configuration for downloads
type DownloadConfig struct {
	MaxRetries      int
	RetryDelay      time.Duration
	BufferSize      int
	ResumeSupported bool
}

// DefaultDownloadConfig returns the default download configuration
func DefaultDownloadConfig() DownloadConfig {
	return DownloadConfig{
		MaxRetries:      3,
		RetryDelay:      time.Second,
		BufferSize:      128 * 1024, // 128KB buffer for optimal performance
		ResumeSupported: true,
	}
}

// EnhancedDownloader provides production-grade downloading with retry and resume
type EnhancedDownloader struct {
	client *http.Client
	config DownloadConfig
}

// NewEnhancedDownloader creates a new enhanced downloader
func NewEnhancedDownloader(config DownloadConfig) *EnhancedDownloader {
	return &EnhancedDownloader{
		client: &http.Client{
			Timeout: 30 * time.Minute,
		},
		config: config,
	}
}

// DownloadError represents a download error with context
type DownloadError struct {
	Stage   string
	Err     error
	Message string
	Help    []string
}

func (e *DownloadError) Error() string {
	msg := fmt.Sprintf("%s: %v", e.Stage, e.Err)
	if e.Message != "" {
		msg = fmt.Sprintf("%s\n\n%s", msg, e.Message)
	}
	if len(e.Help) > 0 {
		msg = fmt.Sprintf("%s\n\nWhat to try:\n", msg)
		for i, h := range e.Help {
			msg = fmt.Sprintf("%s  %d. %s\n", msg, i+1, h)
		}
	}
	return msg
}

// CheckDiskSpace verifies there is enough disk space for the download
func (d *EnhancedDownloader) CheckDiskSpace(destPath string, requiredBytes int64) error {
	// Get directory to check
	dir := filepath.Dir(destPath)

	// Ensure directory exists
	if err := os.MkdirAll(dir, 0755); err != nil {
		return &DownloadError{
			Stage:   "disk space check",
			Err:     err,
			Message: "Failed to create directory for model download.",
			Help: []string{
				"Check if you have write permissions to the directory",
				fmt.Sprintf("Try: chmod u+w %s", dir),
			},
		}
	}

	// Get filesystem stats
	var stat syscall.Statfs_t
	if err := syscall.Statfs(dir, &stat); err != nil {
		// If we can't check, proceed anyway
		return nil
	}

	// Calculate available space
	availableBytes := int64(stat.Bavail) * int64(stat.Bsize)

	// Require 1.2x the file size (20% buffer for temp files)
	requiredWithBuffer := int64(float64(requiredBytes) * 1.2)

	if availableBytes < requiredWithBuffer {
		return &DownloadError{
			Stage: "disk space check",
			Err:   fmt.Errorf("insufficient disk space"),
			Message: fmt.Sprintf("Need %s free, but only %s available.",
				formatBytes(requiredWithBuffer), formatBytes(availableBytes)),
			Help: []string{
				"Free up disk space by removing unused files",
				"Choose a smaller model (e.g., qwen2.5-1.5b instead of qwen2.5-7b)",
				fmt.Sprintf("Delete old models: scmd models remove <model-name>"),
			},
		}
	}

	return nil
}

// DownloadWithProgress downloads a file with retry logic and resume support
func (d *EnhancedDownloader) DownloadWithProgress(ctx context.Context, url, destPath string, expectedSize int64, onProgress func(current, total int64)) error {
	// Check disk space first
	if err := d.CheckDiskSpace(destPath, expectedSize); err != nil {
		return err
	}

	tempPath := destPath + ".tmp"
	var startOffset int64

	// Check if we can resume
	if d.config.ResumeSupported {
		if info, err := os.Stat(tempPath); err == nil {
			startOffset = info.Size()
			fmt.Printf("\n⚡ Resuming download from %.1f MB\n", float64(startOffset)/(1024*1024))
		}
	}

	// Retry loop
	var lastErr error
	for attempt := 1; attempt <= d.config.MaxRetries; attempt++ {
		err := d.downloadAttempt(ctx, url, tempPath, destPath, expectedSize, startOffset, onProgress)
		if err == nil {
			return nil
		}

		lastErr = err

		// Don't retry on certain errors
		if isNonRetryableError(err) {
			return err
		}

		// Last attempt failed
		if attempt == d.config.MaxRetries {
			break
		}

		// Exponential backoff
		delay := d.config.RetryDelay * time.Duration(1<<uint(attempt-1))
		fmt.Printf("\n⚠️  Download failed (attempt %d/%d): %v\n", attempt, d.config.MaxRetries, err)
		fmt.Printf("   Retrying in %s...\n", delay)

		select {
		case <-time.After(delay):
			// Update start offset if we can resume
			if info, err := os.Stat(tempPath); err == nil {
				startOffset = info.Size()
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	// All attempts failed
	return &DownloadError{
		Stage:   "download",
		Err:     lastErr,
		Message: fmt.Sprintf("Failed after %d attempts.", d.config.MaxRetries),
		Help: []string{
			"Check your internet connection",
			"Try again later (network may be temporarily unavailable)",
			"Download manually and place in: " + filepath.Dir(destPath),
			"Use a different network or VPN if the server is blocked",
		},
	}
}

// downloadAttempt performs a single download attempt
func (d *EnhancedDownloader) downloadAttempt(ctx context.Context, url, tempPath, destPath string, expectedSize, startOffset int64, onProgress func(current, total int64)) error {
	// Create or open temp file
	var out *os.File
	var err error

	if startOffset > 0 {
		// Resume: open existing file for append
		out, err = os.OpenFile(tempPath, os.O_WRONLY|os.O_APPEND, 0644)
	} else {
		// New download: create file
		out, err = os.Create(tempPath)
	}

	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer out.Close()

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	// Add headers
	req.Header.Set("User-Agent", "scmd/1.0 (https://github.com/scmd/scmd)")

	// Request resume if we have partial data
	if startOffset > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", startOffset))
	}

	// Execute request
	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if startOffset > 0 {
		// For resume, we accept 206 (Partial Content) or 200 (full content)
		if resp.StatusCode != http.StatusPartialContent && resp.StatusCode != http.StatusOK {
			// Reset and try full download
			startOffset = 0
			out.Close()
			os.Remove(tempPath)
			return fmt.Errorf("resume not supported, restarting download")
		}
		if resp.StatusCode == http.StatusOK {
			// Server doesn't support resume, restart
			startOffset = 0
			out.Close()
			out, err = os.Create(tempPath)
			if err != nil {
				return fmt.Errorf("restart download: %w", err)
			}
			defer out.Close()
		}
	} else {
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("HTTP %d", resp.StatusCode)
		}
	}

	// Get total size
	total := resp.ContentLength
	if startOffset > 0 && resp.StatusCode == http.StatusPartialContent {
		total += startOffset
	}
	if total <= 0 {
		total = expectedSize
	}

	// Download with progress updates
	current := startOffset
	buffer := make([]byte, d.config.BufferSize)

	if onProgress != nil {
		onProgress(current, total)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		n, err := resp.Body.Read(buffer)
		if n > 0 {
			written, writeErr := out.Write(buffer[:n])
			if writeErr != nil {
				return fmt.Errorf("write to file: %w", writeErr)
			}
			current += int64(written)

			if onProgress != nil {
				onProgress(current, total)
			}
		}

		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("read response: %w", err)
		}
	}

	// Close temp file before rename
	out.Close()

	// Verify size if expected
	if expectedSize > 0 {
		info, err := os.Stat(tempPath)
		if err != nil {
			return fmt.Errorf("verify download: %w", err)
		}
		if info.Size() != expectedSize {
			return fmt.Errorf("size mismatch: expected %d bytes, got %d bytes", expectedSize, info.Size())
		}
	}

	// Move to final destination
	if err := os.Rename(tempPath, destPath); err != nil {
		return fmt.Errorf("move file: %w", err)
	}

	return nil
}

// isNonRetryableError checks if an error should not be retried
func isNonRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Context cancellation and disk space errors are non-retryable
	if err == context.Canceled || err == context.DeadlineExceeded {
		return true
	}

	// Disk space errors
	if _, ok := err.(*DownloadError); ok {
		return true
	}

	return false
}
