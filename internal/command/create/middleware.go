package create

import (
	"errors"
	"fmt"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"goyave.dev/gyv/internal/command"
	"goyave.dev/gyv/internal/fs"
	"goyave.dev/gyv/internal/stub"
)

// Middleware command for model generation
type Middleware struct {
	command.ProjectPathCommand
	MiddlewareName string
}

// BuildCobraCommand builds the cobra command for this action
func (c *Middleware) BuildCobraCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "middleware",
		Short: "Create a Goyave middleware",
		Long: `Command to create Goyave middleware.
Only the middleware-name flag is required. The project-path is optional.
If project-path is not specified, the nearest directory containing a go.mod file importing Goyave will be used.`,
		RunE: command.GenerateRunFunc(c),
	}

	c.setFlags(cmd.Flags())

	return cmd
}

// BuildSurvey builds a survey for this action
func (c *Middleware) BuildSurvey() ([]*survey.Question, error) {
	return []*survey.Question{
		{
			Name:     "MiddlewareName",
			Prompt:   &survey.Input{Message: "Middleware name"},
			Validate: survey.Required,
		},
	}, nil
}

// Execute the command's behavior
func (c *Middleware) Execute() error {
	stubPath, err := stub.GenerateStubVersionPath(stub.Middleware, c.GoyaveVersion)
	if err != nil {
		return err
	}

	templateData, err := stub.Load(stubPath, stub.Data{
		"GoyaveImportPath": c.GoyaveMod.Mod.Path,
		"MiddlewareName":   strings.Title(c.MiddlewareName),
	})
	if err != nil {
		return err
	}

	folderPath := fs.CreateMiddlewarePath(c.ProjectPath)

	err = fs.CreateResourceFile(folderPath, c.MiddlewareName, templateData.Bytes())
	if err != nil {
		return err
	}

	fmt.Println("✅ Middleware created!")

	return nil
}

// Validate is a function which check if required flags are definded
func (c *Middleware) Validate() error {
	if c.MiddlewareName == "" {
		return errors.New("❌ required flag \"name\"")
	}

	return nil
}

func (c *Middleware) setFlags(flags *pflag.FlagSet) {
	flags.StringVarP(
		&c.MiddlewareName,
		"name",
		"n",
		"",
		"The name of the middleware to generate",
	)
	flags.StringVarP(
		&c.ProjectPath,
		"project-path",
		"p",
		"",
		"The path to the Goyave project root",
	)
}
