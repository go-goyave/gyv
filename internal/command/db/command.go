package db

import (
	"goyave.dev/gyv/internal/command"

	"github.com/spf13/cobra"
)

// BuildCommand builds a parent command for all database-related subcommands
func BuildCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "db",
		Short: "Database operations",
		Long:  "Command for database operations, such as seeding, migrations, etc.",
	}

	commands := []command.Command{
		&Migrate{},
		&Seed{},
	}

	for _, c := range commands {
		cmd.AddCommand(c.BuildCobraCommand())
	}

	return cmd
}
