package command

import (
	"fmt"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/Masterminds/semver"
	"github.com/spf13/cobra"
	"golang.org/x/mod/modfile"
	"goyave.dev/gyv/internal/mod"
)

// Command is a interface which represents the actions possible for the commands
type Command interface {
	Execute() error
	Validate() error
	UsedFlags() bool
	BuildSurvey() ([]*survey.Question, error)
	BuildCobraCommand() *cobra.Command
}

// GenerateRunFunc generic cobra handler
// If all required flags are set, the command's specific behavior is executed.
// Otherwise a survey is launched for allow the user to inject the data
func GenerateRunFunc(c Command) func(*cobra.Command, []string) error {
	return func(*cobra.Command, []string) error {
		if !c.UsedFlags() {
			questions, err := c.BuildSurvey()
			if err != nil {
				return err
			}

			if err := survey.Ask(questions, c); err != nil {
				return err
			}

		} else if err := c.Validate(); err != nil {
			return err
		}

		if err := c.Execute(); err != nil {
			fmt.Fprintf(os.Stderr, "❌ %s\n", err.Error())
		}

		return nil
	}
}

// ProjectPathCommand shared composition struct for commands
// using a Goyave project path.
// All commands compositing with this one should call "setup()"
// first in their "Execute()" function.
type ProjectPathCommand struct {
	GoyaveMod     *modfile.Require
	GoyaveVersion *semver.Version
	ProjectPath   string
}

// Setup ensure the `ProjectPath` field is correctly set.
// If `ProjectPath` is empty at the time `setup()` is called, its value
// will be set to `fs.FindParentModule()`.
// The project's `go.mod` file is parsed and put into the `GoyaveMod` field.
// The Goyave framework version is parsed and put into the `GoyaveVersion` field.
func (c *ProjectPathCommand) Setup() error {
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
