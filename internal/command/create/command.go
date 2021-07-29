package create

import (
	"goyave.dev/gyv/internal/command"

	"github.com/spf13/cobra"
)

// BuildCommand builds a parent command for all creation-related subcommands
func BuildCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create Goyave projects and resources",
		Long:  "Command to create Goyave projects and resources, such as controllers or models.",
	}

	commands := []command.Command{
		&ProjectData{},
		&ControllerData{},
		&MiddlewareData{},
		&ModelData{},
	}

	for _, c := range commands {
		cmd.AddCommand(c.BuildCobraCommand())
	}

	return cmd
}
