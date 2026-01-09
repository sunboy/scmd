package command

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockCommand struct {
	name        string
	aliases     []string
	description string
	category    Category
}

func (c *mockCommand) Name() string           { return c.name }
func (c *mockCommand) Aliases() []string      { return c.aliases }
func (c *mockCommand) Description() string    { return c.description }
func (c *mockCommand) Usage() string          { return "/" + c.name }
func (c *mockCommand) Examples() []string     { return nil }
func (c *mockCommand) Category() Category     { return c.category }
func (c *mockCommand) RequiresBackend() bool  { return false }
func (c *mockCommand) Validate(_ *Args) error { return nil }
func (c *mockCommand) Execute(_ context.Context, _ *Args, _ *ExecContext) (*Result, error) {
	return &Result{Success: true}, nil
}

func TestRegistry_Register(t *testing.T) {
	r := NewRegistry()

	cmd := &mockCommand{
		name:    "test",
		aliases: []string{"t"},
	}

	err := r.Register(cmd)
	require.NoError(t, err)

	// Should find by name
	found, ok := r.Get("test")
	assert.True(t, ok)
	assert.Equal(t, "test", found.Name())

	// Should find by alias
	found, ok = r.Get("t")
	assert.True(t, ok)
	assert.Equal(t, "test", found.Name())
}

func TestRegistry_Register_Duplicate(t *testing.T) {
	r := NewRegistry()

	cmd1 := &mockCommand{name: "test"}
	cmd2 := &mockCommand{name: "test"}

	err := r.Register(cmd1)
	require.NoError(t, err)

	err = r.Register(cmd2)
	assert.Error(t, err)
}

func TestRegistry_Register_DuplicateAlias(t *testing.T) {
	r := NewRegistry()

	cmd1 := &mockCommand{name: "test1", aliases: []string{"t"}}
	cmd2 := &mockCommand{name: "test2", aliases: []string{"t"}}

	err := r.Register(cmd1)
	require.NoError(t, err)

	err = r.Register(cmd2)
	assert.Error(t, err)
}

func TestRegistry_Get_NotFound(t *testing.T) {
	r := NewRegistry()

	_, ok := r.Get("nonexistent")
	assert.False(t, ok)
}

func TestRegistry_ListByCategory(t *testing.T) {
	r := NewRegistry()

	r.Register(&mockCommand{name: "cmd1", category: CategoryCode})
	r.Register(&mockCommand{name: "cmd2", category: CategoryCode})
	r.Register(&mockCommand{name: "cmd3", category: CategoryGit})

	codeCmds := r.ListByCategory(CategoryCode)
	assert.Len(t, codeCmds, 2)

	gitCmds := r.ListByCategory(CategoryGit)
	assert.Len(t, gitCmds, 1)

	configCmds := r.ListByCategory(CategoryConfig)
	assert.Len(t, configCmds, 0)
}

func TestRegistry_List(t *testing.T) {
	r := NewRegistry()

	r.Register(&mockCommand{name: "zebra"})
	r.Register(&mockCommand{name: "alpha"})
	r.Register(&mockCommand{name: "beta"})

	cmds := r.List()
	assert.Len(t, cmds, 3)

	// Should be sorted by name
	assert.Equal(t, "alpha", cmds[0].Name())
	assert.Equal(t, "beta", cmds[1].Name())
	assert.Equal(t, "zebra", cmds[2].Name())
}

func TestRegistry_Names(t *testing.T) {
	r := NewRegistry()

	r.Register(&mockCommand{name: "cmd1"})
	r.Register(&mockCommand{name: "cmd2"})

	names := r.Names()
	assert.Contains(t, names, "cmd1")
	assert.Contains(t, names, "cmd2")
}

func TestRegistry_Count(t *testing.T) {
	r := NewRegistry()

	assert.Equal(t, 0, r.Count())

	r.Register(&mockCommand{name: "cmd1"})
	assert.Equal(t, 1, r.Count())

	r.Register(&mockCommand{name: "cmd2"})
	assert.Equal(t, 2, r.Count())
}
