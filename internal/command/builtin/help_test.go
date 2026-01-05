package builtin

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/scmd/scmd/internal/command"
	"github.com/scmd/scmd/tests/testutil"
)

func TestHelpCommand_Name(t *testing.T) {
	registry := command.NewRegistry()
	cmd := NewHelpCommand(registry)

	assert.Equal(t, "help", cmd.Name())
}

func TestHelpCommand_Aliases(t *testing.T) {
	registry := command.NewRegistry()
	cmd := NewHelpCommand(registry)

	aliases := cmd.Aliases()
	assert.Contains(t, aliases, "h")
	assert.Contains(t, aliases, "?")
}

func TestHelpCommand_Category(t *testing.T) {
	registry := command.NewRegistry()
	cmd := NewHelpCommand(registry)

	assert.Equal(t, command.CategoryCore, cmd.Category())
}

func TestHelpCommand_RequiresBackend(t *testing.T) {
	registry := command.NewRegistry()
	cmd := NewHelpCommand(registry)

	assert.False(t, cmd.RequiresBackend())
}

func TestHelpCommand_Execute_ShowAll(t *testing.T) {
	registry := command.NewRegistry()
	RegisterAll(registry)

	helpCmd := NewHelpCommand(registry)
	ui := testutil.NewMockUI()

	execCtx := &command.ExecContext{
		UI: ui,
	}

	args := command.NewArgs()
	result, err := helpCmd.Execute(context.Background(), args, execCtx)

	require.NoError(t, err)
	assert.True(t, result.Success)

	output := ui.GetOutput()
	assert.Contains(t, output, "scmd")
	assert.Contains(t, output, "/help")
}

func TestHelpCommand_Execute_ShowSpecific(t *testing.T) {
	registry := command.NewRegistry()
	RegisterAll(registry)

	helpCmd := NewHelpCommand(registry)
	ui := testutil.NewMockUI()

	execCtx := &command.ExecContext{
		UI: ui,
	}

	args := command.NewArgs()
	args.Positional = []string{"explain"}

	result, err := helpCmd.Execute(context.Background(), args, execCtx)

	require.NoError(t, err)
	assert.True(t, result.Success)

	output := ui.GetOutput()
	assert.Contains(t, output, "explain")
}

func TestHelpCommand_Execute_NotFound(t *testing.T) {
	registry := command.NewRegistry()
	helpCmd := NewHelpCommand(registry)
	ui := testutil.NewMockUI()

	execCtx := &command.ExecContext{
		UI: ui,
	}

	args := command.NewArgs()
	args.Positional = []string{"nonexistent"}

	result, err := helpCmd.Execute(context.Background(), args, execCtx)

	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "unknown command")
}
