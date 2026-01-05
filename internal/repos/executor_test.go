package repos

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/scmd/scmd/internal/backend"
	"github.com/scmd/scmd/internal/backend/mock"
	"github.com/scmd/scmd/internal/command"
)

func TestPluginCommand_Name(t *testing.T) {
	spec := &CommandSpec{Name: "test-cmd"}
	cmd := NewPluginCommand(spec)
	assert.Equal(t, "test-cmd", cmd.Name())
}

func TestPluginCommand_Aliases(t *testing.T) {
	spec := &CommandSpec{
		Name:    "test-cmd",
		Aliases: []string{"tc", "test"},
	}
	cmd := NewPluginCommand(spec)
	assert.Equal(t, []string{"tc", "test"}, cmd.Aliases())
}

func TestPluginCommand_Description(t *testing.T) {
	spec := &CommandSpec{Description: "Test description"}
	cmd := NewPluginCommand(spec)
	assert.Equal(t, "Test description", cmd.Description())
}

func TestPluginCommand_Usage(t *testing.T) {
	spec := &CommandSpec{Usage: "test-cmd [args]"}
	cmd := NewPluginCommand(spec)
	assert.Equal(t, "test-cmd [args]", cmd.Usage())
}

func TestPluginCommand_Category(t *testing.T) {
	// With category
	spec := &CommandSpec{Category: "code"}
	cmd := NewPluginCommand(spec)
	assert.Equal(t, command.Category("code"), cmd.Category())

	// Without category - defaults to plugin
	spec2 := &CommandSpec{}
	cmd2 := NewPluginCommand(spec2)
	assert.Equal(t, command.CategoryPlugin, cmd2.Category())
}

func TestPluginCommand_Examples(t *testing.T) {
	spec := &CommandSpec{
		Examples: []string{"example 1", "example 2"},
	}
	cmd := NewPluginCommand(spec)
	assert.Equal(t, []string{"example 1", "example 2"}, cmd.Examples())
}

func TestPluginCommand_RequiresBackend(t *testing.T) {
	cmd := NewPluginCommand(&CommandSpec{})
	assert.True(t, cmd.RequiresBackend())
}

func TestPluginCommand_Validate(t *testing.T) {
	spec := &CommandSpec{
		Args: []ArgSpec{
			{Name: "file", Required: true},
			{Name: "output", Required: false, Default: "out.txt"},
		},
	}
	cmd := NewPluginCommand(spec)

	// Missing required arg
	args := command.NewArgs()
	err := cmd.Validate(args)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing required argument")

	// With required arg
	args.Positional = []string{"input.txt"}
	err = cmd.Validate(args)
	assert.NoError(t, err)
}

func TestPluginCommand_Execute(t *testing.T) {
	spec := &CommandSpec{
		Name:        "greet",
		Description: "Greet someone",
		Args: []ArgSpec{
			{Name: "name", Required: true},
		},
		Prompt: PromptSpec{
			System:   "You are a friendly greeter.",
			Template: "Greet {{.name}} warmly.",
		},
	}
	cmd := NewPluginCommand(spec)

	ctx := context.Background()
	mockBackend := mock.New()

	args := command.NewArgs()
	args.Positional = []string{"Alice"}

	execCtx := &command.ExecContext{
		Backend: mockBackend,
	}

	result, err := cmd.Execute(ctx, args, execCtx)
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.NotEmpty(t, result.Output)
}

func TestPluginCommand_Execute_NoBackend(t *testing.T) {
	spec := &CommandSpec{
		Name: "test",
		Prompt: PromptSpec{
			Template: "Test prompt",
		},
	}
	cmd := NewPluginCommand(spec)

	ctx := context.Background()
	args := command.NewArgs()
	execCtx := &command.ExecContext{
		Backend: nil,
	}

	result, err := cmd.Execute(ctx, args, execCtx)
	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "no backend")
}

func TestPluginCommand_Execute_ValidationError(t *testing.T) {
	spec := &CommandSpec{
		Name: "test",
		Args: []ArgSpec{
			{Name: "required", Required: true},
		},
		Prompt: PromptSpec{
			Template: "Test",
		},
	}
	cmd := NewPluginCommand(spec)

	ctx := context.Background()
	args := command.NewArgs()
	execCtx := &command.ExecContext{
		Backend: mock.New(),
	}

	result, err := cmd.Execute(ctx, args, execCtx)
	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "missing required")
}

func TestPluginCommand_BuildTemplateContext(t *testing.T) {
	spec := &CommandSpec{
		Args: []ArgSpec{
			{Name: "file", Required: true},
			{Name: "output", Required: false, Default: "out.txt"},
		},
		Flags: []FlagSpec{
			{Name: "verbose", Default: "false"},
		},
	}
	cmd := NewPluginCommand(spec)

	args := command.NewArgs()
	args.Positional = []string{"input.txt"}
	args.Options["stdin"] = "stdin content"
	args.Options["verbose"] = "true"

	ctx := cmd.buildTemplateContext(args)

	assert.Equal(t, "input.txt", ctx["file"])
	assert.Equal(t, "out.txt", ctx["output"]) // default
	assert.Equal(t, "true", ctx["verbose"])
	assert.Equal(t, "stdin content", ctx["stdin"])
	assert.Equal(t, "stdin content", ctx["input"])
	assert.Equal(t, []string{"input.txt"}, ctx["args"])
	assert.Equal(t, "input.txt", ctx["all_args"])
}

func TestPluginCommand_ExecuteTemplate(t *testing.T) {
	cmd := NewPluginCommand(&CommandSpec{})

	ctx := map[string]interface{}{
		"name": "World",
		"adj":  "beautiful",
	}

	result, err := cmd.executeTemplate("Hello, {{.name}}! What a {{.adj}} day!", ctx)
	require.NoError(t, err)
	assert.Equal(t, "Hello, World! What a beautiful day!", result)
}

func TestPluginCommand_ExecuteTemplate_Error(t *testing.T) {
	cmd := NewPluginCommand(&CommandSpec{})

	ctx := map[string]interface{}{}

	_, err := cmd.executeTemplate("{{.invalid}", ctx)
	assert.Error(t, err)
}

func TestLoader_LoadAll(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewManager(tmpDir)

	// Install some commands
	specs := []*CommandSpec{
		{Name: "cmd1", Version: "1.0", Prompt: PromptSpec{Template: "Test 1"}},
		{Name: "cmd2", Version: "1.0", Prompt: PromptSpec{Template: "Test 2"}},
	}
	for _, spec := range specs {
		_ = m.InstallCommand(spec, tmpDir)
	}

	loader := NewLoader(m, tmpDir)
	commands, err := loader.LoadAll()
	require.NoError(t, err)
	assert.Len(t, commands, 2)
}

func TestLoader_RegisterAll(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewManager(tmpDir)

	// Install a command
	spec := &CommandSpec{
		Name:    "plugin-cmd",
		Version: "1.0",
		Prompt:  PromptSpec{Template: "Test"},
	}
	_ = m.InstallCommand(spec, tmpDir)

	loader := NewLoader(m, tmpDir)
	registry := command.NewRegistry()

	err := loader.RegisterAll(registry)
	require.NoError(t, err)

	// Check command is registered
	cmd, ok := registry.Get("plugin-cmd")
	assert.True(t, ok)
	assert.Equal(t, "plugin-cmd", cmd.Name())
}

func TestPluginCommand_WithModelPreferences(t *testing.T) {
	spec := &CommandSpec{
		Name: "test",
		Prompt: PromptSpec{
			Template: "Test prompt",
		},
		Model: ModelSpec{
			MaxTokens:   4096,
			Temperature: 0.5,
		},
	}
	cmd := NewPluginCommand(spec)

	ctx := context.Background()
	args := command.NewArgs()

	// Use a mock backend that we can inspect
	mockBackend := mock.New()
	execCtx := &command.ExecContext{
		Backend: mockBackend,
	}

	result, err := cmd.Execute(ctx, args, execCtx)
	require.NoError(t, err)
	assert.True(t, result.Success)
}

func TestPluginCommand_WithStdin(t *testing.T) {
	spec := &CommandSpec{
		Name: "process-input",
		Prompt: PromptSpec{
			Template: "Process this: {{.stdin}}",
		},
	}
	cmd := NewPluginCommand(spec)

	ctx := context.Background()
	args := command.NewArgs()
	args.Options["stdin"] = "Hello from stdin"

	execCtx := &command.ExecContext{
		Backend: mock.New(),
	}

	result, err := cmd.Execute(ctx, args, execCtx)
	require.NoError(t, err)
	assert.True(t, result.Success)
}

// TestCompleteFlow tests the complete flow of installing and running a plugin
func TestCompleteFlow(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewManager(tmpDir)

	// 1. Install a command spec
	spec := &CommandSpec{
		Name:        "summarize",
		Version:     "1.0.0",
		Description: "Summarize text",
		Usage:       "summarize",
		Args: []ArgSpec{
			{Name: "length", Required: false, Default: "short"},
		},
		Prompt: PromptSpec{
			System:   "You are a summarization expert.",
			Template: "Summarize the following text in a {{.length}} format:\n\n{{.stdin}}",
		},
		Model: ModelSpec{
			Temperature: 0.3,
		},
	}

	err := m.InstallCommand(spec, tmpDir)
	require.NoError(t, err)

	// 2. Load commands
	loader := NewLoader(m, tmpDir)
	commands, err := loader.LoadAll()
	require.NoError(t, err)
	assert.Len(t, commands, 1)

	// 3. Execute command
	cmd := commands[0]
	ctx := context.Background()
	args := command.NewArgs()
	args.Options["stdin"] = "This is a long text that needs summarization."

	execCtx := &command.ExecContext{
		Backend: mock.New(),
	}

	result, err := cmd.Execute(ctx, args, execCtx)
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.NotEmpty(t, result.Output)

	// 4. Uninstall
	err = m.UninstallCommand("summarize", tmpDir)
	require.NoError(t, err)

	// 5. Verify uninstalled
	commands, _ = loader.LoadAll()
	assert.Empty(t, commands)
}

// MockBackendWithCapture is a mock that captures requests
type MockBackendWithCapture struct {
	backend.Backend
	LastRequest *backend.CompletionRequest
}

func (m *MockBackendWithCapture) Complete(ctx context.Context, req *backend.CompletionRequest) (*backend.CompletionResponse, error) {
	m.LastRequest = req
	return &backend.CompletionResponse{Content: "Mock response"}, nil
}
