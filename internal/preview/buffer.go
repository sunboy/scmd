package preview

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// Buffer provides an interactive command preview and editing interface
type Buffer struct {
	Command      string
	DetectResult *DetectResult
	Impact       *Impact
	Input        io.Reader
	Output       io.Writer
}

// NewBuffer creates a new command preview buffer
func NewBuffer(command string) *Buffer {
	return &Buffer{
		Command:      command,
		DetectResult: Detect(command),
		Impact:       EstimateImpact(command),
		Input:        os.Stdin,
		Output:       os.Stdout,
	}
}

// Show displays the command preview and prompts for action
func (b *Buffer) Show() (Action, string, error) {
	if !b.DetectResult.IsDestructive {
		// Not destructive, allow immediate execution
		return ActionExecute, b.Command, nil
	}

	// Display warning banner
	b.displayWarning()

	// Display command breakdown
	b.displayBreakdown()

	// Prompt for action
	return b.promptAction()
}

// displayWarning shows the severity warning
func (b *Buffer) displayWarning() {
	severity := b.DetectResult.HighestSeverity
	icon := severity.Icon()

	fmt.Fprintf(b.Output, "\n%s %s DESTRUCTIVE COMMAND DETECTED\n", icon, strings.ToUpper(severity.String()))
	fmt.Fprintf(b.Output, "%s\n\n", strings.Repeat("=", 60))
}

// displayBreakdown shows command details and matched patterns
func (b *Buffer) displayBreakdown() {
	fmt.Fprintf(b.Output, "Command:\n")
	fmt.Fprintf(b.Output, "  %s\n\n", b.Command)

	if len(b.DetectResult.Matches) > 0 {
		fmt.Fprintf(b.Output, "Detected Risks:\n")
		for i, match := range b.DetectResult.Matches {
			fmt.Fprintf(b.Output, "  %d. %s %s\n", i+1, match.Pattern.Severity.Icon(), match.Pattern.Description)
			fmt.Fprintf(b.Output, "     Matched: '%s'\n", match.MatchedText)
		}
		fmt.Fprintf(b.Output, "\n")
	}

	// Show impact estimate if available
	if b.Impact != nil && b.Impact.AffectedType != "" {
		fmt.Fprintf(b.Output, "Estimated Impact:\n")
		fmt.Fprintf(b.Output, "  Affects: %s\n", b.Impact.AffectedType)
		if b.Impact.EstimatedCount > 0 {
			fmt.Fprintf(b.Output, "  Count: ~%d items\n", b.Impact.EstimatedCount)
		} else if b.Impact.EstimatedCount == -1 {
			fmt.Fprintf(b.Output, "  Count: Unknown (potentially many)\n")
		}
		if b.Impact.EstimatedSize > 0 {
			fmt.Fprintf(b.Output, "  Size: ~%s\n", formatSize(b.Impact.EstimatedSize))
		}
		fmt.Fprintf(b.Output, "\n")
	}
}

// promptAction prompts the user for an action
func (b *Buffer) promptAction() (Action, string, error) {
	fmt.Fprintf(b.Output, "What would you like to do?\n")
	fmt.Fprintf(b.Output, "  [E]dit command\n")
	fmt.Fprintf(b.Output, "  [D]ry-run (show what would happen)\n")
	fmt.Fprintf(b.Output, "  [Enter] Execute anyway\n")
	fmt.Fprintf(b.Output, "  [Q]uit / Cancel\n")
	fmt.Fprintf(b.Output, "\nChoice: ")

	reader := bufio.NewReader(b.Input)
	input, err := reader.ReadString('\n')
	if err != nil {
		return ActionQuit, "", err
	}

	input = strings.TrimSpace(strings.ToLower(input))

	switch input {
	case "e", "edit":
		// Edit mode
		editedCmd, err := b.editCommand()
		if err != nil {
			return ActionQuit, "", err
		}
		return ActionEdit, editedCmd, nil

	case "d", "dry", "dry-run", "dryrun":
		// Dry run mode
		return ActionDryRun, b.Command, nil

	case "", "y", "yes", "exec", "execute":
		// Execute
		return ActionExecute, b.Command, nil

	case "q", "quit", "n", "no", "cancel":
		// Quit
		return ActionQuit, "", nil

	default:
		// Invalid input, prompt again
		fmt.Fprintf(b.Output, "Invalid choice. Please try again.\n\n")
		return b.promptAction()
	}
}

// editCommand opens an editor for the user to modify the command
func (b *Buffer) editCommand() (string, error) {
	// Try to use the user's preferred editor
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi" // Fallback to vi
	}

	// Check if editor is available
	if _, err := exec.LookPath(editor); err != nil {
		// Fallback to simple inline edit if editor not found
		return b.inlineEdit()
	}

	// Create temporary file with command
	tmpfile, err := os.CreateTemp("", "scmd-edit-*.sh")
	if err != nil {
		return "", fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.WriteString(b.Command + "\n"); err != nil {
		return "", fmt.Errorf("write temp file: %w", err)
	}
	tmpfile.Close()

	// Open editor
	cmd := exec.Command(editor, tmpfile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("run editor: %w", err)
	}

	// Read edited command
	edited, err := os.ReadFile(tmpfile.Name())
	if err != nil {
		return "", fmt.Errorf("read edited file: %w", err)
	}

	return strings.TrimSpace(string(edited)), nil
}

// inlineEdit provides simple inline editing when no editor is available
func (b *Buffer) inlineEdit() (string, error) {
	fmt.Fprintf(b.Output, "\nEdit command (press Enter when done):\n")
	fmt.Fprintf(b.Output, "> ")

	reader := bufio.NewReader(b.Input)
	edited, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(edited), nil
}

// Action represents the user's chosen action
type Action int

const (
	ActionEdit Action = iota
	ActionDryRun
	ActionExecute
	ActionQuit
)

func (a Action) String() string {
	switch a {
	case ActionEdit:
		return "edit"
	case ActionDryRun:
		return "dry-run"
	case ActionExecute:
		return "execute"
	case ActionQuit:
		return "quit"
	default:
		return "unknown"
	}
}

// formatSize formats bytes into human-readable format
func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// Preview shows a preview and waits for user action
func Preview(command string) (Action, string, error) {
	buffer := NewBuffer(command)
	return buffer.Show()
}

// PreviewAndExecute shows preview, gets user action, and executes if confirmed
func PreviewAndExecute(command string) error {
	action, finalCmd, err := Preview(command)
	if err != nil {
		return err
	}

	switch action {
	case ActionQuit:
		return fmt.Errorf("cancelled by user")

	case ActionDryRun:
		// Show what would be executed
		fmt.Println("\n[DRY RUN] Would execute:")
		fmt.Printf("  %s\n", finalCmd)
		return nil

	case ActionExecute, ActionEdit:
		// Execute the command (either original or edited)
		fmt.Printf("\nExecuting: %s\n\n", finalCmd)
		cmd := exec.Command("sh", "-c", finalCmd)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		return cmd.Run()

	default:
		return fmt.Errorf("unknown action: %s", action)
	}
}
