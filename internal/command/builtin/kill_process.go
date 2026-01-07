package builtin

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/scmd/scmd/internal/command"
)

// KillProcessCmd implements process killing with confirmation
type KillProcessCmd struct{}

func (c *KillProcessCmd) Name() string        { return "kill-process" }
func (c *KillProcessCmd) Aliases() []string   { return []string{"kp", "killp"} }
func (c *KillProcessCmd) Description() string { return "Find and kill processes by name" }
func (c *KillProcessCmd) Usage() string {
	return "kill-process <process-name>"
}
func (c *KillProcessCmd) Examples() []string {
	return []string{
		"scmd kill-process cursor",
		"scmd /kp node",
		"scmd kill-process chrome",
	}
}
func (c *KillProcessCmd) Category() command.Category { return command.CategoryCore }
func (c *KillProcessCmd) RequiresBackend() bool      { return false }

func (c *KillProcessCmd) Validate(args *command.Args) error {
	if len(args.Positional) == 0 {
		return fmt.Errorf("process name required")
	}
	return nil
}

func (c *KillProcessCmd) Execute(ctx context.Context, args *command.Args, execCtx *command.ExecContext) (*command.Result, error) {
	if err := c.Validate(args); err != nil {
		return command.NewErrorResult(err.Error()), nil
	}

	processName := args.Positional[0]

	// Find processes matching the name
	processes, err := findProcesses(processName)
	if err != nil {
		return command.NewErrorResult(fmt.Sprintf("Error finding processes: %v", err)), nil
	}

	if len(processes) == 0 {
		return command.NewErrorResult(
			fmt.Sprintf("No processes found matching '%s'", processName),
			"Try running: ps aux | grep "+processName,
		), nil
	}

	// Display processes
	var output strings.Builder
	output.WriteString(fmt.Sprintf("Found %d process(es) matching '%s':\n\n", len(processes), processName))
	output.WriteString("PID     USER       COMMAND\n")
	output.WriteString("-----   --------   -------\n")
	for _, p := range processes {
		output.WriteString(fmt.Sprintf("%-7s %-10s %s\n", p.PID, p.User, p.Command))
	}
	output.WriteString("\n")

	// Write the list to stderr so user sees it
	execCtx.UI.WriteError(output.String())

	// Ask for confirmation
	confirmMsg := fmt.Sprintf("Kill %d process(es)?", len(processes))
	if !execCtx.UI.Confirm(confirmMsg) {
		return command.NewResult("Operation cancelled"), nil
	}

	// Kill the processes
	killed := []string{}
	failed := []string{}

	for _, p := range processes {
		err := killProcess(p.PID)
		if err != nil {
			failed = append(failed, fmt.Sprintf("%s (%v)", p.PID, err))
		} else {
			killed = append(killed, p.PID)
		}
	}

	// Build result message
	var result strings.Builder
	if len(killed) > 0 {
		result.WriteString(fmt.Sprintf("✓ Killed %d process(es): %s\n", len(killed), strings.Join(killed, ", ")))
	}
	if len(failed) > 0 {
		result.WriteString(fmt.Sprintf("✗ Failed to kill %d process(es): %s\n", len(failed), strings.Join(failed, ", ")))
	}

	if len(failed) > 0 {
		return command.NewErrorResult(result.String()), nil
	}

	return command.NewResult(result.String()), nil
}

// Process represents a running process
type Process struct {
	PID     string
	User    string
	Command string
}

// findProcesses finds processes matching the given name
func findProcesses(name string) ([]Process, error) {
	// Use pgrep -l to find processes (works on macOS and Linux)
	cmd := exec.Command("pgrep", "-l", name)
	output, err := cmd.Output()
	if err != nil {
		// pgrep returns exit code 1 if no matches found
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() == 1 {
				return []Process{}, nil
			}
		}
		return nil, err
	}

	// Parse output (format: "PID name")
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	processes := []Process{}

	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, " ", 2)
		if len(parts) < 2 {
			continue
		}

		pid := parts[0]
		cmdName := parts[1]

		// Get user for this PID
		user := getProcessUser(pid)

		// Get full command line
		fullCmd := getProcessCommand(pid)
		if fullCmd == "" {
			fullCmd = cmdName
		}

		processes = append(processes, Process{
			PID:     pid,
			User:    user,
			Command: fullCmd,
		})
	}

	return processes, nil
}

// getProcessUser gets the user owning a process
func getProcessUser(pid string) string {
	cmd := exec.Command("ps", "-p", pid, "-o", "user=")
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(output))
}

// getProcessCommand gets the full command line for a process
func getProcessCommand(pid string) string {
	cmd := exec.Command("ps", "-p", pid, "-o", "command=")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	fullCmd := strings.TrimSpace(string(output))

	// Truncate long commands
	if len(fullCmd) > 60 {
		fullCmd = fullCmd[:57] + "..."
	}

	return fullCmd
}

// killProcess kills a process by PID
func killProcess(pid string) error {
	pidInt, err := strconv.Atoi(pid)
	if err != nil {
		return fmt.Errorf("invalid PID: %s", pid)
	}

	cmd := exec.Command("kill", fmt.Sprintf("%d", pidInt))
	return cmd.Run()
}
