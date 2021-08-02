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
type Migrate struct {
	command.ProjectPathCommand
}

// BuildCobraCommand builds the cobra command for this action
func (c *Migrate) BuildCobraCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Run migrations",
		Long: `Command to run database migrations.
If project-path is not specified, the nearest directory containing a go.mod file importing Goyave will be used.
`,
		RunE: command.GenerateRunFunc(c),
	}

	c.setFlags(cmd.Flags())

	return cmd
}

// BuildSurvey builds a survey for this action
func (c *Migrate) BuildSurvey() ([]*survey.Question, error) {
	return []*survey.Question{}, nil
}

// Execute the command's behavior
func (c *Migrate) Execute() error {

	seed, err := inject.Migrate(c.ProjectPath)
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
	fmt.Println("💾 Running migrations...")
	err = seed()
	if err1 := os.Chdir(wd); err1 != nil && err == nil {
		err = err1
	}
	if err != nil {
		return err
	}

	fmt.Println("✅ Database migrated!")

	return nil
}

// Validate checks if required flags are definded
func (c *Migrate) Validate() error {
	return nil
}

func (c *Migrate) setFlags(flags *pflag.FlagSet) {
	flags.StringVarP(
		&c.ProjectPath,
		"project-path",
		"p",
		"",
		"The path to the Goyave project root",
	)

}
