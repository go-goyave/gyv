package command

import (
	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
)

// Command is a interface which represents the actions possible for the commands
type Command interface {
	Execute() error
	Validate() error
	UsedFlags() bool
	BuildSurvey() ([]*survey.Question, error)
	BuildCobraCommand() *cobra.Command
}

// GenerateRunFunc is a function which is used by all commands to determine which behavior to take
// If all required flags are set, the main process start
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

			return c.Execute()
		}

		if err := c.Validate(); err != nil {
			return err
		}

		return c.Execute()
	}
}
