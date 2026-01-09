package builtin

import (
	"github.com/scmd/scmd/internal/command"
)

// RegisterAll registers all built-in commands
func RegisterAll(registry *command.Registry) error {
	// Create help command with registry reference
	helpCmd := NewHelpCommand(registry)

	commands := []command.Command{
		helpCmd,
		NewExplainCommand(),
		NewReviewCommand(),
		NewConfigCommand(),
		NewCmdCommand(),
		&KillProcessCmd{},
	}

	for _, cmd := range commands {
		if err := registry.Register(cmd); err != nil {
			return err
		}
	}

	return nil
}
