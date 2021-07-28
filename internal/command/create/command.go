package create

import (
	"golang.org/x/mod/modfile"
	"goyave.dev/gyv/internal/command"
	"goyave.dev/gyv/internal/mod"

	"github.com/Masterminds/semver"
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

// projectPathCommand shared composition struct for commands
// using a Goyave project path.
// All commands compositing with this one should call "setup()"
// first in their "Execute()" function.
type projectPathCommand struct {
	GoyaveMod     *modfile.Require
	GoyaveVersion *semver.Version
	ProjectPath   string
}

// setup ensure the `ProjectPath` field is correctly set.
// If `ProjectPath` is empty at the time `setup()` is called, its value
// will be set to `fs.FindParentModule()`.
// The project's `go.mod` file is parsed and put into the `GoyaveMod` field.
// The Goyave framework version is parsed and put into the `GoyaveVersion` field.
func (c *projectPathCommand) setup() error {
	if c.ProjectPath == "" {
		c.ProjectPath = mod.FindParentModule()
		if c.ProjectPath == "" {
			return mod.ErrNoGoMod
		}
	}
	modFile, err := mod.Parse(c.ProjectPath)
	if err != nil {
		return err
	}

	c.GoyaveMod = mod.FindGoyaveRequire(modFile)
	if c.GoyaveMod == nil {
		return mod.ErrNotAGoyaveProject
	}

	c.GoyaveVersion, err = semver.NewVersion(c.GoyaveMod.Mod.Version)
	if err != nil {
		return err
	}
	return nil
}

// TODO projectPathCommand could also contain cobra flags and survey
