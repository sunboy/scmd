package llamacpp

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// Downloader handles model downloads
type Downloader struct {
	client *http.Client
}

// NewDownloader creates a new downloader
func NewDownloader() *Downloader {
	return &Downloader{
		client: &http.Client{
			Timeout: 30 * time.Minute, // Large models can take time
		},
	}
}

// DownloadWithProgress downloads a file with progress callback
func (d *Downloader) DownloadWithProgress(url, destPath string, onProgress func(current, total int64)) error {
	// Create temp file
	tempPath := destPath + ".downloading"

	// Clean up temp file on error
	defer func() {
		if _, err := os.Stat(tempPath); err == nil {
			os.Remove(tempPath)
		}
	}()

	// Create the file
	out, err := os.Create(tempPath)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer out.Close()

	// Start download
	resp, err := d.client.Get(url)
	if err != nil {
		return fmt.Errorf("download request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
	}

	// Get total size
	total := resp.ContentLength
	if total <= 0 {
		total = 0 // Unknown size
	}

	// Create progress reader
	var current int64
	buffer := make([]byte, 32*1024) // 32KB buffer

	// Update progress initially
	if onProgress != nil {
		onProgress(0, total)
	}

	// Download with progress updates
	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			written, writeErr := out.Write(buffer[:n])
			if writeErr != nil {
				return fmt.Errorf("write to file: %w", writeErr)
			}
			current += int64(written)

			// Update progress
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

	// Close the file
	if err := out.Close(); err != nil {
		return fmt.Errorf("close file: %w", err)
	}

	// Move temp file to final destination
	if err := os.Rename(tempPath, destPath); err != nil {
		return fmt.Errorf("move file: %w", err)
	}

	return nil
}

// Download downloads a file without progress tracking
func (d *Downloader) Download(url, destPath string) error {
	return d.DownloadWithProgress(url, destPath, nil)
}

// GetFileSize gets the size of a remote file
func (d *Downloader) GetFileSize(url string) (int64, error) {
	resp, err := d.client.Head(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	return resp.ContentLength, nil
}
