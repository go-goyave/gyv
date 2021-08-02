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

// Seed command for running seeders.
type Seed struct {
	command.ProjectPathCommand
}

// BuildCobraCommand builds the cobra command for this action
func (c *Seed) BuildCobraCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "seed",
		Short: "Run seeders",
		Long: `Command to run seeders.
If project-path is not specified, the nearest directory containing a go.mod file importing Goyave will be used.
`,
		RunE: command.GenerateRunFunc(c),
	}

	c.setFlags(cmd.Flags())

	return cmd
}

// BuildSurvey builds a survey for this action
func (c *Seed) BuildSurvey() ([]*survey.Question, error) {
	return []*survey.Question{}, nil
}

// Execute the command's behavior
func (c *Seed) Execute() error {

	// TODO pick the seeders that need to be run using AST exported functions of seeder package
	// TODO if seeder returns error
	// TODO handle panics
	seed, err := inject.Seeder(c.ProjectPath, []string{"Run"})
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
	fmt.Println("💾 Running seeders...")
	err = seed()
	if err1 := os.Chdir(wd); err1 != nil && err == nil {
		err = err1
	}
	if err != nil {
		return err
	}

	fmt.Println("✅ Database seeded!")

	return nil
}

// Validate checks if required flags are definded
func (c *Seed) Validate() error {
	return nil
}

func (c *Seed) setFlags(flags *pflag.FlagSet) {
	flags.StringVarP(
		&c.ProjectPath,
		"project-path",
		"p",
		"",
		"The path to the Goyave project root",
	)

}
