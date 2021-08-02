package db

import (
	"fmt"
	"os"
	"sort"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"goyave.dev/gyv/internal/command"
	"goyave.dev/gyv/internal/inject"
)

// Seed command for running seeders.
type Seed struct {
	command.ProjectPathCommand
	ExportedFunctions []string
	Seeders           []string
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

// Setup parse AST to find seeder functions.
func (c *Seed) Setup() (int, error) {
	consumedFlags, err := c.ProjectPathCommand.Setup()
	if err != nil {
		return consumedFlags, err
	}

	functions, err := inject.FindExportedFunctions(c.ProjectPath + "/database/seeder")
	if err != nil {
		return consumedFlags, err
	}

	if len(functions) == 0 {
		return consumedFlags, fmt.Errorf("No seeder function found")
	}

	sort.SliceStable(functions, func(i, j int) bool {
		if functions[i] == "Run" {
			return true
		}

		return false
	})
	c.ExportedFunctions = functions
	return consumedFlags, nil
}

// BuildSurvey builds a survey for this action
func (c *Seed) BuildSurvey() ([]*survey.Question, error) {
	defaultOptions := []string{}
	for _, v := range c.ExportedFunctions {
		if v == "Run" {
			defaultOptions = append(defaultOptions, v)
			break
		}
	}

	return []*survey.Question{
		{
			Name: "Seeders",
			Prompt: &survey.MultiSelect{
				Message: "Select seeders to run",
				Options: c.ExportedFunctions,
				Default: defaultOptions,
			},
			Validate: survey.Required,
		},
	}, nil
}

// Execute the command's behavior
func (c *Seed) Execute() error {

	for _, s := range c.Seeders {
		if !functionExists(s, c.ExportedFunctions) {
			return fmt.Errorf("Seeder function %q does not exist", s)
		}
	}

	// TODO if seeder returns error
	// TODO handle panics
	seed, err := inject.Seeder(c.ProjectPath, c.Seeders)
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
	flags.StringSliceVarP(
		&c.Seeders,
		"seeders",
		"s",
		[]string{},
		"A list of seeder functions to run",
	)
	flags.StringVarP(
		&c.ProjectPath,
		"project-path",
		"p",
		"",
		"The path to the Goyave project root",
	)

}

func functionExists(funcName string, functions []string) bool {
	for _, f := range functions {
		if f == funcName {
			return true
		}
	}
	return false
}
