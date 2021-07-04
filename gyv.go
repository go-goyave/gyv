package gyv

import (
	"goyave.dev/gyv/command/create"

	"github.com/spf13/cobra"
)

func buildGyv() *cobra.Command {
	gyv := &cobra.Command{
		Use:     "cli-sample",
		Version: "0.1.0",
		Short:   "Resource generator tool for the goyave framework",
		Long: `gyv is a resource generator tool for the Goyave framework.
This tool work with all goyave versions.`,
	}

	commands := []*cobra.Command{
		create.BuildCommand(),
	}

	for _, c := range commands {
		gyv.AddCommand(c)
	}

	return gyv
}

// Execute is a function which start the root command gyv
func Execute() {
	rootCommand := buildGyv()
	_ = rootCommand.Execute()
}
