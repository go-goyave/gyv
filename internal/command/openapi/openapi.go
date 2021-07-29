package openapi

import (
	"fmt"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"goyave.dev/gyv/internal/command"
	"goyave.dev/gyv/internal/inject"
)

// OpenAPI command implementation for OpenAPI 3 specification generation.
type OpenAPI struct {
	command.ProjectPathCommand
	Output string
}

// BuildCobraCommand builds the cobra command for this action
func (c *OpenAPI) BuildCobraCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "openapi",
		Short: "Generate an OpenAPI 3 specification",
		Long: `Generate an OpenAPI 3 specification and saves it to a file named by the output flag.
If project-path is not specified, the nearest directory containing a go.mod file importing Goyave will be used.`,
		RunE: command.GenerateRunFunc(c),
	}

	c.setFlags(cmd.Flags())

	return cmd
}

// BuildSurvey builds a survey for this action
func (c *OpenAPI) BuildSurvey() ([]*survey.Question, error) {
	return []*survey.Question{
		{
			Name: "output",
			Prompt: &survey.Input{
				Message: "File output name",
				Default: "openapi.json",
			},
			Validate: survey.Required,
		},
		{
			Name:   "ProjectPath",
			Prompt: &survey.Input{Message: "Project path (leave empty for auto-detect)"},
		},
	}, nil
}

// Execute the command's behavior
func (c *OpenAPI) Execute() error {
	if c.Output == "" {
		c.Output = "openapi.json"
	}

	if err := c.Setup(); err != nil {
		return err
	}

	plug, err := inject.OpenAPI3Generator(c.ProjectPath)
	if err != nil {
		return err
	}

	s, err := plug.Lookup("GenerateOpenAPI")
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
	spec, err := s.(func() ([]byte, error))()
	if err1 := os.Chdir(wd); err1 != nil && err == nil {
		err = err1
	}
	if err != nil {
		return err
	}

	fmt.Println("✏ Writing output file")
	if err := os.WriteFile(fmt.Sprintf("%s%c%s", c.ProjectPath, os.PathSeparator, c.Output), spec, 0644); err != nil {
		return err
	}

	fmt.Println("✅ OpenAPI 3 specification generated!")

	return nil
}

// Validate is a function which check if required flags are definded
func (c *OpenAPI) Validate() error {
	return nil
}

// UsedFlags is a function which check if flags are used
func (c *OpenAPI) UsedFlags() bool {
	for _, arg := range os.Args[1:] {
		// FIXME double-dash arguments with "=" syntax are not recognized here
		if arg == "--output" || arg == "-o" || arg == "-p" || arg == "--project-path" {
			return true
		}
	}

	return false
}

func (c *OpenAPI) setFlags(flags *pflag.FlagSet) {
	flags.StringVarP(
		&c.Output,
		"output",
		"o",
		"",
		"File output name",
	)
	flags.StringVarP(
		&c.ProjectPath,
		"project-path",
		"p",
		"",
		"The path to the Goyave project root",
	)
}
