package main

import (
	"github.com/spf13/cobra"
	"goyave.dev/gyv/command/create"
)

func buildRootCommand() *cobra.Command {
	gyv := &cobra.Command{
		Use:     "cli-sample",
		Version: "0.1.0",
		Short:   "Productivity CLI for the Goyave framework",
		Long: `gyv productivity command-line interface for the Goyave framework.
All commands can be run either in interactive mode or using POSIX flags.`,
	}

	commands := []*cobra.Command{
		create.BuildCommand(),
	}

	for _, c := range commands {
		gyv.AddCommand(c)
	}

	return gyv
}

func execute() {
	rootCommand := buildRootCommand()
	_ = rootCommand.Execute()
}

func main() {
	execute()
}
