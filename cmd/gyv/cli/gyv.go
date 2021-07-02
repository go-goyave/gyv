package cli

import (
	"goyave.dev/gyv/create"

	"github.com/spf13/cobra"
)

func BuildGyv() *cobra.Command {
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

func Execute() {
	rootCommand := BuildGyv()
	_ = rootCommand.Execute()
}
