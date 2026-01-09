package ui

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"time"
)

// ProgressBar provides a clean single-line progress display
type ProgressBar struct {
	Total       int64
	current     int64
	startTime   time.Time
	lastUpdate  time.Time
	writer      io.Writer
	description string
	mu          sync.Mutex
	finished    bool
}

// NewProgressBar creates a new progress bar
func NewProgressBar(total int64, description string, writer io.Writer) *ProgressBar {
	if writer == nil {
		writer = io.Discard
	}
	return &ProgressBar{
		Total:       total,
		description: description,
		writer:      writer,
		startTime:   time.Now(),
		lastUpdate:  time.Now(),
	}
}

// Update updates the progress
func (p *ProgressBar) Update(current int64) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.finished {
		return
	}

	p.current = current

	// Only update display every 100ms to reduce flicker
	if time.Since(p.lastUpdate) < 100*time.Millisecond && current < p.Total {
		return
	}

	p.lastUpdate = time.Now()
	p.render()
}

// Finish marks the progress as complete
func (p *ProgressBar) Finish() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.finished {
		return
	}

	p.current = p.Total
	p.finished = true
	p.render()
	fmt.Fprintln(p.writer) // New line after completion
}

// render draws the progress bar
func (p *ProgressBar) render() {
	if p.Total <= 0 {
		return
	}

	percent := float64(p.current) * 100 / float64(p.Total)
	elapsed := time.Since(p.startTime)

	// Calculate ETA
	var eta string
	if p.current > 0 && p.current < p.Total {
		totalTime := elapsed.Seconds() * float64(p.Total) / float64(p.current)
		remaining := time.Duration(totalTime-elapsed.Seconds()) * time.Second
		eta = fmt.Sprintf(" ETA: %s", formatDuration(remaining))
	} else if p.current >= p.Total {
		eta = " Complete!"
	}

	// Build progress bar
	barWidth := 30
	filled := int(float64(barWidth) * percent / 100)
	if filled > barWidth {
		filled = barWidth
	}

	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)

	// Format bytes
	currentMB := float64(p.current) / (1024 * 1024)
	totalMB := float64(p.Total) / (1024 * 1024)

	// Clear line and draw progress
	fmt.Fprintf(p.writer, "\r%-20s [%s] %6.1f%% %.1f/%.1f MB%s    ",
		p.description,
		bar,
		percent,
		currentMB,
		totalMB,
		eta,
	)
}

// formatDuration formats a duration in a human-readable way
func formatDuration(d time.Duration) string {
	if d < time.Second {
		return "< 1s"
	}
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	minutes := int(d.Minutes())
	seconds := int(d.Seconds()) % 60
	if minutes < 60 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	hours := minutes / 60
	minutes = minutes % 60
	return fmt.Sprintf("%dh %dm", hours, minutes)
}

// SimpleProgress provides a minimal progress indicator
func SimpleProgress(description string, writer io.Writer) func() {
	if writer == nil {
		writer = io.Discard
	}

	done := make(chan bool)
	go func() {
		frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		i := 0
		for {
			select {
			case <-done:
				fmt.Fprintf(writer, "\r%-50s\n", description+" ✓")
				return
			case <-time.After(100 * time.Millisecond):
				fmt.Fprintf(writer, "\r%s %s", frames[i%len(frames)], description)
				i++
			}
		}
	}()

	return func() {
		close(done)
		time.Sleep(100 * time.Millisecond) // Let the goroutine finish
	}
}
