package cli

import (
	"context"
	"io"
	"os"
	"time"
)

// StdinReader handles piped input
type StdinReader struct {
	timeout time.Duration
	maxSize int64
}

// NewStdinReader creates a new stdin reader
func NewStdinReader() *StdinReader {
	return &StdinReader{
		timeout: 30 * time.Second,
		maxSize: 10 * 1024 * 1024, // 10MB
	}
}

// WithTimeout sets the read timeout
func (r *StdinReader) WithTimeout(d time.Duration) *StdinReader {
	r.timeout = d
	return r
}

// WithMaxSize sets the maximum read size
func (r *StdinReader) WithMaxSize(size int64) *StdinReader {
	r.maxSize = size
	return r
}

// Read reads all stdin with timeout
func (r *StdinReader) Read(ctx context.Context) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	type result struct {
		data []byte
		err  error
	}
	ch := make(chan result, 1)

	go func() {
		data, err := io.ReadAll(io.LimitReader(os.Stdin, r.maxSize))
		ch <- result{data, err}
	}()

	select {
	case res := <-ch:
		return string(res.data), res.err
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

// HasInput checks if there is data available on stdin
func HasInput() bool {
	stat, _ := os.Stdin.Stat()
	return (stat.Mode() & os.ModeCharDevice) == 0
}
