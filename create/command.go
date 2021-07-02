package create

import (
	"goyave.dev/gyv/command"

	"github.com/spf13/cobra"
)

func BuildCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create goyave framework resources",
		Long: `Command for create goyave resources, it's composed of many subcommands.
You can either use the flags to inject the information or not use them.
If you don't use the flags a survey will be launched and with it you will be able to insert the information.`,
	}

	commands := []command.Command{
		&CreateProject{},
		&CreateController{},
		&CreateMiddleware{},
		&CreateModel{},
	}

	for _, c := range commands {
		cmd.AddCommand(c.BuildCobraCommand())
	}

	return cmd
}
