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

// Command minimal command definition.
type Command interface {
	Execute() error
	Validate() error
	BuildSurvey() ([]*survey.Question, error)
	BuildCobraCommand() *cobra.Command
}

// SetupCommand for commands needing to execute some logic
// before being executed. If a command implements "Setup()",
// this function will be called first, before surveys and validation.
type SetupCommand interface {
	// Setup returns the number of flags consumed by the setup operation.
	Setup() (int, error)
}

// GenerateRunFunc generic cobra handler
// If all required flags are set, the command's specific behavior is executed.
// Otherwise a survey is launched for allow the user to inject the data
func GenerateRunFunc(c Command) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		consumedFlags := 0
		if c, ok := c.(SetupCommand); ok {
			consumed, err := c.Setup()
			if err != nil {
				fmt.Fprintf(os.Stderr, "❌ %s\n", err.Error())
				return nil
			}
			consumedFlags += consumed
		}

		if cmd.Flags().NFlag()-consumedFlags == 0 {
			questions, err := c.BuildSurvey()
			if err != nil {
				fmt.Fprintf(os.Stderr, "❌ %s\n", err.Error())
				return nil
			}

			if err := survey.Ask(questions, c); err != nil {
				fmt.Fprintf(os.Stderr, "❌ %s\n", err.Error())
				return nil
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
// Commands using ProjectPathCommand should have a project path flag
// and survey entry.
type ProjectPathCommand struct {
	GoyaveMod     *modfile.Require
	GoyaveVersion *semver.Version
	ProjectPath   string
}

// Setup ensure the `ProjectPath` field is correctly set.
// If `ProjectPath` is empty at the time `Setup()` is called, its value
// will be set to `fs.FindParentModule()`.
// The project's `go.mod` file is parsed and put into the `GoyaveMod` field.
// The Goyave framework version is parsed and put into the `GoyaveVersion` field.
func (c *ProjectPathCommand) Setup() (int, error) {
	consumedFlags := 1
	if c.ProjectPath == "" {
		consumedFlags = 0
		c.ProjectPath = mod.FindParentModule()
		if c.ProjectPath == "" {
			return consumedFlags, mod.ErrNoGoMod
		}
	}
	modFile, err := mod.Parse(c.ProjectPath)
	if err != nil {
		return consumedFlags, err
	}

	c.GoyaveMod = mod.FindGoyaveRequire(modFile)
	if c.GoyaveMod == nil {
		return consumedFlags, mod.ErrNotAGoyaveProject
	}

	c.GoyaveVersion, err = semver.NewVersion(c.GoyaveMod.Mod.Version)
	if err != nil {
		return consumedFlags, err
	}
	return consumedFlags, nil
}
