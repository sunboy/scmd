package cli

import (
	"os"

	"golang.org/x/term"
)

// IOMode represents the input/output mode
type IOMode struct {
	HasStdin    bool // Data is being piped in
	StdinIsTTY  bool // Stdin is a terminal
	StdoutIsTTY bool // Stdout is a terminal
	StderrIsTTY bool // Stderr is a terminal
	Interactive bool // Full interactive mode
	PipeIn      bool // Receiving piped input
	PipeOut     bool // Output is being piped
}

// DetectIOMode determines how scmd is being invoked
func DetectIOMode() *IOMode {
	stdinIsTTY := term.IsTerminal(int(os.Stdin.Fd()))
	stdoutIsTTY := term.IsTerminal(int(os.Stdout.Fd()))
	stderrIsTTY := term.IsTerminal(int(os.Stderr.Fd()))

	return &IOMode{
		HasStdin:    !stdinIsTTY,
		StdinIsTTY:  stdinIsTTY,
		StdoutIsTTY: stdoutIsTTY,
		StderrIsTTY: stderrIsTTY,
		Interactive: stdinIsTTY && stdoutIsTTY,
		PipeIn:      !stdinIsTTY,
		PipeOut:     !stdoutIsTTY,
	}
}

// ShouldStream returns true if output should stream
func (m *IOMode) ShouldStream() bool {
	return m.StdoutIsTTY
}

// ShouldShowProgress returns true if progress indicators should show
func (m *IOMode) ShouldShowProgress() bool {
	return m.StdoutIsTTY && m.StderrIsTTY
}

// ShouldUseColors returns true if colors should be used
func (m *IOMode) ShouldUseColors() bool {
	return m.StdoutIsTTY
}

// ProgressWriter returns the appropriate writer for progress output
func (m *IOMode) ProgressWriter() *os.File {
	if m.PipeOut {
		return os.Stderr
	}
	return os.Stdout
}
