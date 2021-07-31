package main

import (
	"github.com/spf13/cobra"
	"goyave.dev/gyv/internal/command/create"
	"goyave.dev/gyv/internal/command/openapi"
)

func buildRootCommand() *cobra.Command {
	gyv := &cobra.Command{
		Use:     "gyv",
		Version: "0.1.0", // TODO use ldflags to set version at compile-time
		Short:   "Productivity CLI for the Goyave framework",
		Long: `gyv productivity command-line interface for the Goyave framework.
All commands can be run either in interactive mode or using POSIX flags.`,
	}

	commands := []*cobra.Command{
		create.BuildCommand(),
		(&openapi.OpenAPI{}).BuildCobraCommand(),
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
