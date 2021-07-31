package create

import (
	"errors"
	"fmt"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"goyave.dev/gyv/internal/command"
	"goyave.dev/gyv/internal/fs"
	"goyave.dev/gyv/internal/stub"
)

// Controller command for controller generation
type Controller struct {
	command.ProjectPathCommand
	ControllerName string
}

// BuildCobraCommand builds the cobra command for this action
func (c *Controller) BuildCobraCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "controller",
		Short: "Create a Goyave controller",
		Long: `Command to create Goyave controller.
Only the controller-name flag is required. The project-path is optional.
If project-path is not specified, the nearest directory containing a go.mod file importing Goyave will be used.
`,
		RunE: command.GenerateRunFunc(c),
	}

	c.setFlags(cmd.Flags())

	return cmd
}

// BuildSurvey builds a survey for this action
func (c *Controller) BuildSurvey() ([]*survey.Question, error) {
	return []*survey.Question{
		{
			Name:     "ControllerName",
			Prompt:   &survey.Input{Message: "Controller name"},
			Validate: survey.Required,
		},
		{
			Name:   "ProjectPath",
			Prompt: &survey.Input{Message: "Project path (leave empty for auto-detect)"},
		},
	}, nil
}

// Execute the command's behavior
func (c *Controller) Execute() error {

	if err := c.Setup(); err != nil {
		return err
	}

	// TODO extract actual behavior (excluding validation and visual output)
	// That would help "front-end" part of the CLI to be swapped with ease.
	folderPath, err := fs.CreateControllerPath(c.ControllerName, c.ProjectPath, c.GoyaveVersion)
	if err != nil {
		return err
	}

	stubPath, err := stub.GenerateStubVersionPath(stub.Controller, c.GoyaveVersion)
	if err != nil {
		return err
	}

	templateData, err := stub.Load(stubPath, stub.Data{
		"GoyaveImportPath": c.GoyaveMod.Mod.Path,
		"ControllerName":   c.ControllerName,
	})
	if err != nil {
		return err
	}

	if err := os.MkdirAll(folderPath, 0744); err != nil {
		return err
	}

	err = fs.CreateResourceFile(folderPath, c.ControllerName, templateData.Bytes())
	if err != nil {
		return err
	}

	fmt.Println("✅ Controller created!")

	return nil
}

// Validate checks if required flags are definded
func (c *Controller) Validate() error {
	if c.ControllerName == "" {
		return errors.New("required flag(s) \"name\"")
	}

	return nil
}

func (c *Controller) setFlags(flags *pflag.FlagSet) {
	flags.StringVarP(
		&c.ControllerName,
		"name",
		"n",
		"",
		"The name of the controller to generate",
	)
	flags.StringVarP(
		&c.ProjectPath,
		"project-path",
		"p",
		"",
		"The path to the Goyave project root",
	)

}
