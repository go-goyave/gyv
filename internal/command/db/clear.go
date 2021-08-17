package db

import (
	"fmt"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"goyave.dev/gyv/internal/command"
	"goyave.dev/gyv/internal/inject"
)

// Migrate command for running auto migrations.
type Clear struct {
	command.ProjectPathCommand
}

// BuildCobraCommand builds the cobra command for this action
func (c *Clear) BuildCobraCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clear",
		Short: "Clear database",
		Long: `Command to delete all records from all registered models.
If project-path is not specified, the nearest directory containing a go.mod file importing Goyave will be used.
`,
		RunE: command.GenerateRunFunc(c),
	}

	c.setFlags(cmd.Flags())

	return cmd
}

// BuildSurvey builds a survey for this action
func (c *Clear) BuildSurvey() ([]*survey.Question, error) {
	return []*survey.Question{}, nil
}

// Execute the command's behavior
func (c *Clear) Execute() error {

	seed, err := inject.DBClear(c.ProjectPath)
	if err != nil {
		return err
	}

	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	if err := os.Chdir(c.ProjectPath); err != nil {
		return err
	}
	fmt.Println("🗑️ Clearing database...")
	err = seed()
	if err1 := os.Chdir(wd); err1 != nil && err == nil {
		err = err1
	}
	if err != nil {
		return err
	}

	fmt.Println("✅ Database cleared!")

	return nil
}

// Validate checks if required flags are definded
func (c *Clear) Validate() error {
	return nil
}

func (c *Clear) setFlags(flags *pflag.FlagSet) {
	flags.StringVarP(
		&c.ProjectPath,
		"project-path",
		"p",
		"",
		"The path to the Goyave project root",
	)

}
