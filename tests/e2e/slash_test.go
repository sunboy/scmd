package e2e

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/scmd/scmd/internal/backend/mock"
	"github.com/scmd/scmd/internal/command"
	"github.com/scmd/scmd/internal/repos"
	"github.com/scmd/scmd/internal/slash"
)

// ==================== SLASH COMMAND CLI TESTS ====================

func TestSlash_List(t *testing.T) {
	stdout, _, err := runScmd(t, "slash", "list")
	if err != nil {
		t.Fatalf("slash list failed: %v", err)
	}

	// Should show default slash commands
	if !strings.Contains(stdout, "explain") {
		t.Error("should list explain command")
	}

	if !strings.Contains(stdout, "review") {
		t.Error("should list review command")
	}
}

func TestSlash_Init_Bash(t *testing.T) {
	stdout, _, err := runScmd(t, "slash", "init", "bash")
	if err != nil {
		t.Fatalf("slash init bash failed: %v", err)
	}

	// Should generate bash integration
	if !strings.Contains(stdout, "function") || !strings.Contains(stdout, "alias") {
		t.Error("should generate bash functions and aliases")
	}

	if !strings.Contains(stdout, "/explain") {
		t.Error("should include /explain alias")
	}
}

func TestSlash_Init_Zsh(t *testing.T) {
	stdout, _, err := runScmd(t, "slash", "init", "zsh")
	if err != nil {
		t.Fatalf("slash init zsh failed: %v", err)
	}

	if !strings.Contains(stdout, "function") {
		t.Error("should generate zsh integration")
	}
}

func TestSlash_Init_Fish(t *testing.T) {
	stdout, _, err := runScmd(t, "slash", "init", "fish")
	if err != nil {
		t.Fatalf("slash init fish failed: %v", err)
	}

	if !strings.Contains(stdout, "function /") {
		t.Error("should generate fish function")
	}

	if !strings.Contains(stdout, "alias") {
		t.Error("should generate fish aliases")
	}
}

func TestSlash_Run_Explain(t *testing.T) {
	code := "func hello() { println(\"world\") }"
	stdout, _, err := runScmdWithStdin(t, code, "-b", "mock", "slash", "run", "explain")
	if err != nil {
		t.Fatalf("slash run explain failed: %v", err)
	}

	if stdout == "" {
		t.Error("should have output")
	}
}

func TestSlash_Run_Review(t *testing.T) {
	code := "var x = 1"
	stdout, _, err := runScmdWithStdin(t, code, "-b", "mock", "slash", "run", "review")
	if err != nil {
		t.Fatalf("slash run review failed: %v", err)
	}

	if stdout == "" {
		t.Error("should have output")
	}
}

func TestSlash_Run_WithAlias(t *testing.T) {
	// Test using alias 'e' for explain
	stdout, _, err := runScmd(t, "-b", "mock", "slash", "run", "e", "goroutines")
	if err != nil {
		t.Fatalf("slash run with alias failed: %v", err)
	}

	if stdout == "" {
		t.Error("should have output")
	}
}

func TestSlash_DirectInvocation_Explain(t *testing.T) {
	// Test direct /explain invocation
	code := "print('hello')"
	stdout, _, err := runScmdWithStdin(t, code, "-b", "mock", "/explain")
	if err != nil {
		t.Fatalf("direct /explain failed: %v", err)
	}

	if stdout == "" {
		t.Error("should have output")
	}
}

func TestSlash_DirectInvocation_Review(t *testing.T) {
	code := "def foo(): pass"
	stdout, _, err := runScmdWithStdin(t, code, "-b", "mock", "/review")
	if err != nil {
		t.Fatalf("direct /review failed: %v", err)
	}

	if stdout == "" {
		t.Error("should have output")
	}
}

func TestSlash_DirectInvocation_WithArgs(t *testing.T) {
	stdout, _, err := runScmd(t, "-b", "mock", "/explain", "what", "is", "a", "closure")
	if err != nil {
		t.Fatalf("direct /explain with args failed: %v", err)
	}

	if stdout == "" {
		t.Error("should have output")
	}
}

// ==================== SLASH RUNNER UNIT TESTS ====================

func TestSlashRunner_LoadConfig(t *testing.T) {
	tmpDir := t.TempDir()
	registry := command.NewRegistry()
	repoMgr := repos.NewManager(tmpDir)

	runner := slash.NewRunner(tmpDir, registry, repoMgr)

	err := runner.LoadConfig()
	if err != nil {
		t.Fatalf("load config failed: %v", err)
	}

	// Should have default commands
	commands := runner.List()
	if len(commands) == 0 {
		t.Error("should have default commands")
	}

	// Check for expected default commands
	found := make(map[string]bool)
	for _, cmd := range commands {
		found[cmd.Name] = true
	}

	expected := []string{"explain", "review", "commit", "summarize", "fix"}
	for _, name := range expected {
		if !found[name] {
			t.Errorf("missing default command: %s", name)
		}
	}
}

func TestSlashRunner_FindCommand(t *testing.T) {
	tmpDir := t.TempDir()
	registry := command.NewRegistry()
	repoMgr := repos.NewManager(tmpDir)

	runner := slash.NewRunner(tmpDir, registry, repoMgr)
	runner.LoadConfig()

	// Find by name
	cmd := runner.FindCommand("explain")
	if cmd == nil {
		t.Error("should find explain command")
	}

	if cmd.Name != "explain" {
		t.Errorf("expected name 'explain', got '%s'", cmd.Name)
	}

	// Find by alias
	cmd = runner.FindCommand("e")
	if cmd == nil {
		t.Error("should find command by alias 'e'")
	}

	// Should be case-insensitive
	cmd = runner.FindCommand("EXPLAIN")
	if cmd == nil {
		t.Error("should find command case-insensitively")
	}

	// Non-existent command
	cmd = runner.FindCommand("nonexistent")
	if cmd != nil {
		t.Error("should not find nonexistent command")
	}
}

func TestSlashRunner_Parse(t *testing.T) {
	tmpDir := t.TempDir()
	registry := command.NewRegistry()
	repoMgr := repos.NewManager(tmpDir)

	runner := slash.NewRunner(tmpDir, registry, repoMgr)
	runner.LoadConfig()

	tests := []struct {
		input       string
		expectCmd   string
		expectArgs  []string
		expectError bool
	}{
		{"/explain hello", "explain", []string{"hello"}, false},
		{"/review", "review", []string{}, false},
		{"/e what is go", "explain", []string{"what", "is", "go"}, false},
		{"/r code.py", "review", []string{"code.py"}, false},
		{"explain", "", nil, true}, // Missing /
		{"/", "", nil, true},        // Empty command
		{"/nonexistent", "", nil, true},
	}

	for _, tt := range tests {
		cmd, args, err := runner.Parse(tt.input)

		if tt.expectError {
			if err == nil {
				t.Errorf("input %q: expected error", tt.input)
			}
			continue
		}

		if err != nil {
			t.Errorf("input %q: unexpected error: %v", tt.input, err)
			continue
		}

		if cmd == nil {
			t.Errorf("input %q: got nil command", tt.input)
			continue
		}

		if cmd.Name != tt.expectCmd && cmd.Command != tt.expectCmd {
			t.Errorf("input %q: expected command %s, got %s/%s",
				tt.input, tt.expectCmd, cmd.Name, cmd.Command)
		}

		if len(args) != len(tt.expectArgs) {
			t.Errorf("input %q: expected %d args, got %d",
				tt.input, len(tt.expectArgs), len(args))
		}

		for i, arg := range args {
			if i >= len(tt.expectArgs) {
				break
			}
			if arg != tt.expectArgs[i] {
				t.Errorf("input %q: arg %d: expected %s, got %s",
					tt.input, i, tt.expectArgs[i], arg)
			}
		}
	}
}

func TestSlashRunner_Add(t *testing.T) {
	tmpDir := t.TempDir()
	registry := command.NewRegistry()
	repoMgr := repos.NewManager(tmpDir)

	runner := slash.NewRunner(tmpDir, registry, repoMgr)
	runner.LoadConfig()

	// Add new command
	newCmd := slash.SlashCommand{
		Name:        "test",
		Command:     "test-cmd",
		Aliases:     []string{"t"},
		Description: "Test command",
		Stdin:       true,
	}

	err := runner.Add(newCmd)
	if err != nil {
		t.Fatalf("add command failed: %v", err)
	}

	// Verify it was added
	cmd := runner.FindCommand("test")
	if cmd == nil {
		t.Error("should find newly added command")
	}

	// Try adding duplicate
	err = runner.Add(newCmd)
	if err == nil {
		t.Error("should fail adding duplicate command")
	}
}

func TestSlashRunner_Remove(t *testing.T) {
	tmpDir := t.TempDir()
	registry := command.NewRegistry()
	repoMgr := repos.NewManager(tmpDir)

	runner := slash.NewRunner(tmpDir, registry, repoMgr)
	runner.LoadConfig()

	// Add a command first
	newCmd := slash.SlashCommand{
		Name:        "temp",
		Command:     "temp-cmd",
		Description: "Temporary command",
	}
	runner.Add(newCmd)

	// Remove it
	err := runner.Remove("temp")
	if err != nil {
		t.Fatalf("remove command failed: %v", err)
	}

	// Verify it's gone
	cmd := runner.FindCommand("temp")
	if cmd != nil {
		t.Error("command should be removed")
	}

	// Try removing nonexistent
	err = runner.Remove("nonexistent")
	if err == nil {
		t.Error("should fail removing nonexistent command")
	}
}

func TestSlashRunner_AddAlias(t *testing.T) {
	tmpDir := t.TempDir()
	registry := command.NewRegistry()
	repoMgr := repos.NewManager(tmpDir)

	runner := slash.NewRunner(tmpDir, registry, repoMgr)
	runner.LoadConfig()

	// Add alias to existing command
	err := runner.AddAlias("explain", "exp")
	if err != nil {
		t.Fatalf("add alias failed: %v", err)
	}

	// Verify alias works
	cmd := runner.FindCommand("exp")
	if cmd == nil {
		t.Error("should find command by new alias")
	}

	if cmd.Name != "explain" {
		t.Error("alias should point to explain command")
	}

	// Try adding duplicate alias
	err = runner.AddAlias("review", "e") // 'e' already used
	if err == nil {
		t.Error("should fail adding duplicate alias")
	}

	// Try adding alias to nonexistent command
	err = runner.AddAlias("nonexistent", "x")
	if err == nil {
		t.Error("should fail adding alias to nonexistent command")
	}
}

func TestSlashRunner_SaveAndReload(t *testing.T) {
	tmpDir := t.TempDir()
	registry := command.NewRegistry()
	repoMgr := repos.NewManager(tmpDir)

	// Create runner and add custom command
	runner1 := slash.NewRunner(tmpDir, registry, repoMgr)
	runner1.LoadConfig()

	customCmd := slash.SlashCommand{
		Name:        "custom",
		Command:     "custom-command",
		Aliases:     []string{"c", "cust"},
		Description: "Custom test command",
		Args:        "arg1 arg2",
		Stdin:       true,
	}

	err := runner1.Add(customCmd)
	if err != nil {
		t.Fatalf("add command failed: %v", err)
	}

	// Create new runner and load config
	runner2 := slash.NewRunner(tmpDir, registry, repoMgr)
	err = runner2.LoadConfig()
	if err != nil {
		t.Fatalf("load config failed: %v", err)
	}

	// Verify custom command persisted
	cmd := runner2.FindCommand("custom")
	if cmd == nil {
		t.Fatal("custom command should be persisted")
	}

	if cmd.Command != "custom-command" {
		t.Error("command field not persisted correctly")
	}

	if len(cmd.Aliases) != 2 {
		t.Errorf("expected 2 aliases, got %d", len(cmd.Aliases))
	}

	if cmd.Args != "arg1 arg2" {
		t.Errorf("args not persisted correctly: %s", cmd.Args)
	}

	if !cmd.Stdin {
		t.Error("stdin flag not persisted correctly")
	}
}

func TestSlashRunner_GenerateShellIntegration_Bash(t *testing.T) {
	tmpDir := t.TempDir()
	registry := command.NewRegistry()
	repoMgr := repos.NewManager(tmpDir)

	runner := slash.NewRunner(tmpDir, registry, repoMgr)
	runner.LoadConfig()

	script := runner.GenerateShellIntegration("bash")

	// Check for bash syntax
	if !strings.Contains(script, "function") && !strings.Contains(script, "() {") {
		t.Error("should contain bash function syntax")
	}

	// Check for command aliases
	if !strings.Contains(script, "/explain") {
		t.Error("should contain /explain alias")
	}

	// Check for completion
	if !strings.Contains(script, "complete") {
		t.Error("should contain bash completion")
	}
}

func TestSlashRunner_GenerateShellIntegration_Fish(t *testing.T) {
	tmpDir := t.TempDir()
	registry := command.NewRegistry()
	repoMgr := repos.NewManager(tmpDir)

	runner := slash.NewRunner(tmpDir, registry, repoMgr)
	runner.LoadConfig()

	script := runner.GenerateShellIntegration("fish")

	// Check for fish syntax
	if !strings.Contains(script, "function /") {
		t.Error("should contain fish function syntax")
	}

	// Check for command aliases
	if !strings.Contains(script, "alias") {
		t.Error("should contain fish aliases")
	}
}

// ==================== SLASH COMMAND EXECUTION ====================

func TestSlashRunner_Run_WithMockBackend(t *testing.T) {
	tmpDir := t.TempDir()
	registry := command.NewRegistry()
	repoMgr := repos.NewManager(tmpDir)

	// Register mock backend commands
	registry.Register(newMockExplainCommand())

	runner := slash.NewRunner(tmpDir, registry, repoMgr)
	runner.LoadConfig()

	slashCmd := runner.FindCommand("explain")
	if slashCmd == nil {
		t.Fatal("explain command not found")
	}

	backend := mock.New()
	result, err := runner.Run(
		context.Background(),
		slashCmd,
		[]string{"test", "args"},
		"test input",
		backend,
	)

	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	if result == nil {
		t.Fatal("result should not be nil")
	}
}

// ==================== EDGE CASES ====================

func TestSlashRunner_EmptyConfig(t *testing.T) {
	tmpDir := t.TempDir()

	// Create empty config file
	configPath := filepath.Join(tmpDir, "slash.yaml")
	if err := os.WriteFile(configPath, []byte("commands: []"), 0644); err != nil {
		t.Fatal(err)
	}

	registry := command.NewRegistry()
	repoMgr := repos.NewManager(tmpDir)

	runner := slash.NewRunner(tmpDir, registry, repoMgr)
	err := runner.LoadConfig()
	if err != nil {
		t.Fatalf("load empty config failed: %v", err)
	}

	// Should handle empty config gracefully
	commands := runner.List()
	if len(commands) != 0 {
		t.Error("empty config should have no commands")
	}
}

func TestSlashRunner_CorruptConfig(t *testing.T) {
	tmpDir := t.TempDir()

	// Create corrupt config file
	configPath := filepath.Join(tmpDir, "slash.yaml")
	if err := os.WriteFile(configPath, []byte("invalid: yaml: ["), 0644); err != nil {
		t.Fatal(err)
	}

	registry := command.NewRegistry()
	repoMgr := repos.NewManager(tmpDir)

	runner := slash.NewRunner(tmpDir, registry, repoMgr)
	err := runner.LoadConfig()
	if err == nil {
		t.Error("should fail with corrupt config")
	}
}

func TestSlashRunner_CommandWithSpecialChars(t *testing.T) {
	tmpDir := t.TempDir()
	registry := command.NewRegistry()
	repoMgr := repos.NewManager(tmpDir)

	runner := slash.NewRunner(tmpDir, registry, repoMgr)
	runner.LoadConfig()

	// Add command with special characters in name
	newCmd := slash.SlashCommand{
		Name:        "test-cmd",
		Command:     "test-command",
		Description: "Test with dash",
	}

	err := runner.Add(newCmd)
	if err != nil {
		t.Fatalf("add command with dash failed: %v", err)
	}

	cmd := runner.FindCommand("test-cmd")
	if cmd == nil {
		t.Error("should find command with dash")
	}
}

// ==================== HELPER COMMANDS ====================

// Mock command for testing
type mockExplainCommand struct{}

func newMockExplainCommand() *mockExplainCommand {
	return &mockExplainCommand{}
}

func (c *mockExplainCommand) Name() string                                    { return "explain" }
func (c *mockExplainCommand) Aliases() []string                               { return []string{"e"} }
func (c *mockExplainCommand) Description() string                             { return "Explain code" }
func (c *mockExplainCommand) Usage() string                                   { return "explain <code>" }
func (c *mockExplainCommand) Examples() []string                              { return []string{} }
func (c *mockExplainCommand) Category() command.Category                      { return command.CategoryCode }
func (c *mockExplainCommand) Validate(args *command.Args) error               { return nil }
func (c *mockExplainCommand) RequiresBackend() bool                           { return true }
func (c *mockExplainCommand) Execute(ctx context.Context, args *command.Args, execCtx *command.ExecContext) (*command.Result, error) {
	return &command.Result{
		Success: true,
		Output:  "Mock explanation",
	}, nil
}
