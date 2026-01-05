package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/scmd/scmd/internal/backend"
	"github.com/scmd/scmd/internal/backend/mock"
	"github.com/scmd/scmd/internal/command"
	"github.com/scmd/scmd/internal/command/builtin"
	"github.com/scmd/scmd/internal/config"
	"github.com/scmd/scmd/pkg/version"
)

var (
	cfg          *config.Config
	verbose      bool
	promptFlag   string
	outputFlag   string
	formatFlag   string
	quietFlag    bool
	contextFlags []string

	// Global registries
	cmdRegistry     *command.Registry
	backendRegistry *backend.Registry
)

var rootCmd = &cobra.Command{
	Use:   "scmd",
	Short: "AI-powered slash commands in your terminal",
	Long: `scmd brings AI-powered slash commands to any terminal.

Examples:
  scmd                           Start interactive mode
  scmd explain file.go           Explain code
  cat foo.md | scmd -p "summarize this" > summary.md
  git diff | scmd review -o review.md`,
	Version:           version.Short(),
	PersistentPreRunE: preRun,
	RunE:              runRoot,
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	// Pipe/prompt flags
	rootCmd.PersistentFlags().StringVarP(&promptFlag, "prompt", "p", "", "inline prompt")
	rootCmd.PersistentFlags().StringVarP(&outputFlag, "output", "o", "", "output file")
	rootCmd.PersistentFlags().StringVarP(&formatFlag, "format", "f", "text", "output format: text, json, markdown")
	rootCmd.PersistentFlags().BoolVarP(&quietFlag, "quiet", "q", false, "suppress progress")
	rootCmd.PersistentFlags().StringArrayVarP(&contextFlags, "context", "c", nil, "context files")

	// Add built-in commands
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(explainCmd)
	rootCmd.AddCommand(reviewCmd)
	rootCmd.AddCommand(configCmd)

	rootCmd.CompletionOptions.DisableDefaultCmd = true
}

// explainCmd wraps the builtin explain command
var explainCmd = &cobra.Command{
	Use:     "explain [file|concept]",
	Short:   "Explain code or concepts",
	Aliases: []string{"e", "what"},
	Example: `  scmd explain main.go
  scmd explain "what is a goroutine"
  cat file.py | scmd explain`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runBuiltinCommand("explain", args)
	},
}

// reviewCmd wraps the builtin review command
var reviewCmd = &cobra.Command{
	Use:     "review [file]",
	Short:   "Review code for issues and improvements",
	Aliases: []string{"r"},
	Example: `  scmd review main.go
  git diff | scmd review`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runBuiltinCommand("review", args)
	},
}

// configCmd wraps the builtin config command
var configCmd = &cobra.Command{
	Use:     "config [key] [value]",
	Short:   "View or modify configuration",
	Aliases: []string{"cfg"},
	Example: `  scmd config
  scmd config backends.default
  scmd config ui.colors true`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runBuiltinCommand("config", args)
	},
}

func runBuiltinCommand(name string, args []string) error {
	ctx := context.Background()
	mode := DetectIOMode()

	// Read stdin if piped
	var stdinContent string
	if mode.PipeIn {
		reader := NewStdinReader()
		content, err := reader.Read(ctx)
		if err != nil {
			return fmt.Errorf("read stdin: %w", err)
		}
		stdinContent = content
	}

	// Setup output
	output, err := NewOutputWriter(&OutputConfig{FilePath: outputFlag, Mode: mode})
	if err != nil {
		return err
	}
	defer output.Close()

	// Create execution context
	defaultBackend, _ := backendRegistry.Default()
	execCtx := &command.ExecContext{
		Config:  cfg,
		Backend: defaultBackend,
		UI:      NewConsoleUI(mode),
	}

	// Get the command
	c, ok := cmdRegistry.Get(name)
	if !ok {
		return fmt.Errorf("unknown command: %s", name)
	}

	// Build args
	cmdArgs := command.NewArgs()
	cmdArgs.Positional = args
	if stdinContent != "" {
		cmdArgs.Options["stdin"] = stdinContent
	}

	// Execute
	result, err := c.Execute(ctx, cmdArgs, execCtx)
	if err != nil {
		return err
	}

	if result.Output != "" {
		output.WriteLine(result.Output)
	}

	if !result.Success {
		if len(result.Suggestions) > 0 {
			fmt.Fprintln(os.Stderr, "Suggestions:")
			for _, s := range result.Suggestions {
				fmt.Fprintf(os.Stderr, "  - %s\n", s)
			}
		}
		return fmt.Errorf("%s", result.Error)
	}

	return nil
}

func preRun(_ *cobra.Command, _ []string) error {
	var err error

	// Load configuration
	cfg, err = config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Initialize registries
	cmdRegistry = command.NewRegistry()
	backendRegistry = backend.NewRegistry()

	// Register mock backend for now
	mockBackend := mock.New()
	if err := backendRegistry.Register(mockBackend); err != nil {
		return fmt.Errorf("register backend: %w", err)
	}

	// Register built-in commands
	if err := builtin.RegisterAll(cmdRegistry); err != nil {
		return fmt.Errorf("register commands: %w", err)
	}

	return nil
}

func runRoot(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	mode := DetectIOMode()

	// Read stdin if piped
	var stdinContent string
	if mode.PipeIn {
		reader := NewStdinReader()
		content, err := reader.Read(ctx)
		if err != nil {
			return fmt.Errorf("read stdin: %w", err)
		}
		stdinContent = content
	}

	// Setup output
	output, err := NewOutputWriter(&OutputConfig{FilePath: outputFlag, Mode: mode})
	if err != nil {
		return err
	}
	defer output.Close()

	// Create execution context
	defaultBackend, _ := backendRegistry.Default()
	execCtx := &command.ExecContext{
		Config:  cfg,
		Backend: defaultBackend,
		UI:      NewConsoleUI(mode),
	}

	// Handle -p flag
	if promptFlag != "" {
		return runPrompt(ctx, promptFlag, stdinContent, mode, output, execCtx)
	}

	// Handle command by name from internal registry (for slash commands in REPL)
	if len(args) > 0 {
		cmdName := args[0]
		if c, ok := cmdRegistry.Get(cmdName); ok {
			cmdArgs := command.NewArgs()
			cmdArgs.Positional = args[1:]
			if stdinContent != "" {
				cmdArgs.Options["stdin"] = stdinContent
			}
			result, err := c.Execute(ctx, cmdArgs, execCtx)
			if err != nil {
				return err
			}
			if result.Output != "" {
				output.WriteLine(result.Output)
			}
			if !result.Success {
				return fmt.Errorf("%s", result.Error)
			}
			return nil
		}
	}

	// Pipe to command
	if mode.PipeIn && len(args) > 0 {
		return runCommandWithStdin(ctx, args[0], args[1:], stdinContent, mode, output, execCtx)
	}

	// Interactive mode or help
	if mode.Interactive {
		return runREPL(execCtx)
	}

	return cmd.Help()
}

func runPrompt(ctx context.Context, prompt, stdin string, mode *IOMode, output *OutputWriter, execCtx *command.ExecContext) error {
	if !quietFlag && mode.StderrIsTTY {
		fmt.Fprintln(mode.ProgressWriter(), "⏳ Processing...")
	}

	// Use backend for completion
	if execCtx.Backend != nil {
		fullPrompt := prompt
		if stdin != "" {
			fullPrompt = fmt.Sprintf("%s\n\nInput:\n%s", prompt, stdin)
		}

		req := &backend.CompletionRequest{
			Prompt:      fullPrompt,
			MaxTokens:   2048,
			Temperature: 0.7,
		}

		resp, err := execCtx.Backend.Complete(ctx, req)
		if err != nil {
			return fmt.Errorf("completion failed: %w", err)
		}

		output.WriteLine(resp.Content)
		return nil
	}

	// Placeholder if no backend
	result := fmt.Sprintf("Prompt: %s\nInput length: %d bytes", prompt, len(stdin))
	output.WriteLine(result)
	return nil
}

func runCommandWithStdin(ctx context.Context, cmdName string, args []string, stdin string, mode *IOMode, output *OutputWriter, execCtx *command.ExecContext) error {
	c, ok := cmdRegistry.Get(cmdName)
	if !ok {
		return fmt.Errorf("unknown command: %s", cmdName)
	}

	cmdArgs := command.NewArgs()
	cmdArgs.Positional = args
	cmdArgs.Options["stdin"] = stdin

	result, err := c.Execute(ctx, cmdArgs, execCtx)
	if err != nil {
		return err
	}

	if result.Output != "" {
		output.WriteLine(result.Output)
	}

	if !result.Success {
		return fmt.Errorf("%s", result.Error)
	}

	return nil
}

func runREPL(execCtx *command.ExecContext) error {
	fmt.Println("scmd - AI-powered slash commands")
	fmt.Println("Type /help for available commands")
	fmt.Println()

	// Simple REPL - for now just show help
	helpCmd, _ := cmdRegistry.Get("help")
	if helpCmd != nil {
		_, _ = helpCmd.Execute(context.Background(), command.NewArgs(), execCtx)
	}

	return nil
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

// ConsoleUI implements command.UI for terminal output
type ConsoleUI struct {
	mode *IOMode
}

// NewConsoleUI creates a new console UI
func NewConsoleUI(mode *IOMode) *ConsoleUI {
	return &ConsoleUI{mode: mode}
}

// Write writes to stdout
func (u *ConsoleUI) Write(s string) {
	fmt.Print(s)
}

// WriteLine writes a line to stdout
func (u *ConsoleUI) WriteLine(s string) {
	fmt.Println(s)
}

// WriteError writes to stderr
func (u *ConsoleUI) WriteError(s string) {
	fmt.Fprintln(os.Stderr, s)
}

// Confirm prompts for confirmation
func (u *ConsoleUI) Confirm(prompt string) bool {
	fmt.Printf("%s [y/N]: ", prompt)
	var response string
	fmt.Scanln(&response)
	return response == "y" || response == "Y"
}

// Spinner shows a loading spinner (simplified)
func (u *ConsoleUI) Spinner(message string) func() {
	if u.mode.StdoutIsTTY {
		fmt.Printf("⏳ %s...", message)
	}
	return func() {
		if u.mode.StdoutIsTTY {
			fmt.Println(" ✓")
		}
	}
}
