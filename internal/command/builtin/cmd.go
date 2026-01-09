package builtin

import (
	"context"
	"fmt"
	"strings"

	"github.com/scmd/scmd/internal/backend"
	"github.com/scmd/scmd/internal/command"
	"github.com/scmd/scmd/internal/utils/manpage"
)

// CmdCommand implements /cmd - generates exact commands from natural language queries
type CmdCommand struct{}

// NewCmdCommand creates a new cmd command
func NewCmdCommand() *CmdCommand {
	return &CmdCommand{}
}

// Name returns the command name
func (c *CmdCommand) Name() string { return "cmd" }

// Aliases returns command aliases
func (c *CmdCommand) Aliases() []string { return []string{"command", "howto"} }

// Description returns the command description
func (c *CmdCommand) Description() string {
	return "Generate exact CLI commands from natural language questions"
}

// Usage returns usage information
func (c *CmdCommand) Usage() string {
	return "/cmd <question>"
}

// Category returns the command category
func (c *CmdCommand) Category() command.Category { return command.CategoryCore }

// RequiresBackend returns true
func (c *CmdCommand) RequiresBackend() bool { return true }

// Examples returns example usages
func (c *CmdCommand) Examples() []string {
	return []string{
		`/cmd "how do I find all files modified in the last 24 hours?"`,
		`/cmd "search for text in all .go files"`,
		`/cmd "compress a directory into a tar.gz file"`,
		`/cmd "list all running processes sorted by memory usage"`,
		`scmd /cmd "download a file from a URL"`,
	}
}

// Validate validates arguments
func (c *CmdCommand) Validate(args *command.Args) error {
	// Need either a question or piped input
	if len(args.Positional) == 0 && args.Options["stdin"] == "" {
		return fmt.Errorf("provide a question about what command you need")
	}
	return nil
}

// Execute runs the cmd command
func (c *CmdCommand) Execute(
	ctx context.Context,
	args *command.Args,
	execCtx *command.ExecContext,
) (*command.Result, error) {
	var query string

	// Get query from stdin or positional args
	if stdin, ok := args.Options["stdin"]; ok && stdin != "" {
		query = stdin
	} else if len(args.Positional) > 0 {
		query = strings.Join(args.Positional, " ")
	}

	query = strings.TrimSpace(query)
	if query == "" {
		return command.NewErrorResult(
			"empty query",
			"Provide a question about what command you need",
			"Example: /cmd \"how do I find files modified today?\"",
		), nil
	}

	// Check if backend is available
	if execCtx.Backend == nil {
		return command.NewErrorResult(
			"no backend available",
			"Configure a backend with 'scmd setup'",
		), nil
	}

	// Show progress
	execCtx.UI.WriteLine(fmt.Sprintf("🔍 Analyzing query: %s", query))

	// Detect relevant commands
	detectedCommands := manpage.DetectCommands(query)

	if len(detectedCommands) == 0 {
		execCtx.UI.WriteLine("⚠️  No specific commands detected, using general CLI knowledge")
	} else {
		execCtx.UI.WriteLine(fmt.Sprintf("📖 Reading man pages for: %s", strings.Join(detectedCommands, ", ")))
	}

	// Read man pages for detected commands
	manPages := manpage.ReadMultiple(detectedCommands)

	// Build prompt
	prompt := buildCmdPrompt(query, manPages)

	// Show progress
	stop := execCtx.UI.Spinner("Generating command")
	defer stop()

	// Call backend
	req := &backend.CompletionRequest{
		Prompt:      prompt,
		MaxTokens:   1024,
		Temperature: 0.1, // Low temperature for precise command generation
		SystemPrompt: `You are a CLI command expert. Generate exact, precise commands based on user questions and man page documentation.

IMPORTANT RULES:
1. Provide ONLY the exact command to run - no extra explanation unless asked
2. Use the most common, safe, and widely compatible options
3. If the query is unclear, ask for clarification
4. Always explain what the command does after showing it
5. Include helpful flags and options based on the man pages
6. Format output as:
   Command: <exact command>

   Explanation: <what it does and why>

Be precise and accurate. Double-check your commands.`,
	}

	resp, err := execCtx.Backend.Complete(ctx, req)
	if err != nil {
		return command.NewErrorResult(
			fmt.Sprintf("backend error: %v", err),
		), nil
	}

	// Format the output nicely
	output := formatCmdOutput(resp.Content, query)

	return command.NewResult(output), nil
}

// buildCmdPrompt builds the prompt for the LLM
func buildCmdPrompt(query string, manPages map[string]*manpage.ManPage) string {
	var parts []string

	parts = append(parts, "USER QUESTION:")
	parts = append(parts, query)
	parts = append(parts, "")

	if len(manPages) > 0 {
		parts = append(parts, "RELEVANT MAN PAGES:")
		parts = append(parts, "")
		parts = append(parts, manpage.FormatForLLM(manPages))
		parts = append(parts, "")
		parts = append(parts, "Based on the man pages above and the user's question, provide the exact command needed.")
	} else {
		parts = append(parts, "No man pages found. Use your general CLI knowledge to provide the command.")
	}

	return strings.Join(parts, "\n")
}

// formatCmdOutput formats the LLM response nicely
func formatCmdOutput(response, query string) string {
	// Add a nice header
	var output strings.Builder

	output.WriteString("═══════════════════════════════════════════════════\n")
	output.WriteString("📝 Command for: " + query + "\n")
	output.WriteString("═══════════════════════════════════════════════════\n\n")

	output.WriteString(response)

	output.WriteString("\n\n═══════════════════════════════════════════════════\n")
	output.WriteString("💡 Tip: Always test commands in a safe environment first!\n")
	output.WriteString("═══════════════════════════════════════════════════\n")

	return output.String()
}
