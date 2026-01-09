//go:build windows
// +build windows

package llamacpp

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/sys/windows"
)

// checkDiskSpace checks if there's enough disk space for the download (Windows version)
func checkDiskSpace(path string, requiredBytes uint64) error {
	// Get the volume path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	volumePath := filepath.VolumeName(absPath)
	if volumePath == "" {
		volumePath = filepath.Dir(absPath)
	}

	// Ensure it ends with a backslash
	if !strings.HasSuffix(volumePath, "\\") {
		volumePath += "\\"
	}

	var freeBytesAvailable, totalBytes, totalFreeBytes uint64
	volumePathPtr, err := windows.UTF16PtrFromString(volumePath)
	if err != nil {
		return fmt.Errorf("failed to convert path: %w", err)
	}

	err = windows.GetDiskFreeSpaceEx(
		volumePathPtr,
		&freeBytesAvailable,
		&totalBytes,
		&totalFreeBytes,
	)
	if err != nil {
		return fmt.Errorf("failed to get disk space: %w", err)
	}

	availableGB := float64(freeBytesAvailable) / (1024 * 1024 * 1024)
	requiredGB := float64(requiredBytes) / (1024 * 1024 * 1024)

	if freeBytesAvailable < requiredBytes {
		return fmt.Errorf("insufficient disk space: need %.2f GB, only %.2f GB available",
			requiredGB, availableGB)
	}

	return nil
}

// downloadWithResume downloads a file with resume support (Windows version)
func downloadWithResume(ctx context.Context, url, destPath string, expectedSize int64) error {
	// Check if partial download exists
	existingSize := int64(0)
	if info, err := os.Stat(destPath); err == nil {
		existingSize = info.Size()
	}

	// Create HTTP request with Range header if resuming
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if existingSize > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", existingSize))
	}

	client := &http.Client{
		Timeout: 30 * time.Minute,
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to start download: %w", err)
	}
	defer resp.Body.Close()

	// Check if resume is supported
	if existingSize > 0 && resp.StatusCode != http.StatusPartialContent {
		// Resume not supported, start over
		existingSize = 0
		os.Remove(destPath)
	}

	// Open file for writing (append if resuming)
	flag := os.O_CREATE | os.O_WRONLY
	if existingSize > 0 {
		flag |= os.O_APPEND
	} else {
		flag |= os.O_TRUNC
	}

	out, err := os.OpenFile(destPath, flag, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer out.Close()

	// Copy with progress
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("download interrupted: %w", err)
	}

	// Verify size
	finalInfo, err := os.Stat(destPath)
	if err != nil {
		return fmt.Errorf("failed to stat downloaded file: %w", err)
	}

	if expectedSize > 0 && finalInfo.Size() != expectedSize {
		return fmt.Errorf("size mismatch: expected %d bytes, got %d bytes",
			expectedSize, finalInfo.Size())
	}

	return nil
}
