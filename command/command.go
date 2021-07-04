package command

import (
	"fmt"

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
			return fmt.Errorf("❌ %w", err)
		}

		return nil
	}
}
