package command

import (
	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
)

type Command interface {
	Execute() error
	Validate() error
	UsedFlags() bool
	BuildSurvey() ([]*survey.Question, error)
	BuildCobraCommand() *cobra.Command
}

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
