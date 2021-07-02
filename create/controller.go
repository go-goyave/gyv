package create

import (
	"errors"
	"fmt"
	"os"

	"goyave.dev/gyv/command"
	"goyave.dev/gyv/fs"
	"goyave.dev/gyv/stub"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type CreateController struct {
	ControllerName string
	ProjectPath    string
}

func (c *CreateController) BuildCobraCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "controller",
		Short: "Create goyave controller",
		Long: `Command for create goyave controller with the help of a survey.
Only the controller-name flag is required. The project-path is optional.
Is project-path is empty, the program going to search for a goyave project root.
And for this, the program going to move up to the parent directory and check each time if this directory is a goyave project.
If any parents directories are goyave project, an error will be raised.`,
		RunE: command.GenerateRunFunc(c),
	}

	c.setFlags(cmd.Flags())

	return cmd
}

func (c *CreateController) BuildSurvey() ([]*survey.Question, error) {
	return []*survey.Question{
		{
			Name:     "controllerName",
			Prompt:   &survey.Input{Message: "Input the name of the controller to generate"},
			Validate: survey.Required,
		},
		{
			Name:   "projectPath",
			Prompt: &survey.Input{Message: "Input the path to the goyave project"},
		},
	}, nil
}

func (c *CreateController) Execute() error {
	if err := fs.IsValidProject(c.ProjectPath); err != nil {
		return fmt.Errorf("❌ %s", err.Error())
	}

	folderPath, err := fs.CreateControllerPath(c.ControllerName, c.ProjectPath)
	if err != nil {
		return fmt.Errorf("❌ %s", err.Error())
	}

	goyaveModVersion, err := fs.GetGoyavePath(c.ProjectPath)
	if err != nil {
		return fmt.Errorf("❌ %s", err.Error())
	}

	goyaveVersion, err := fs.GetGoyaveVersion(c.ProjectPath)
	if err != nil {
		return fmt.Errorf("❌ %s", err.Error())
	}

	stubPath, err := stub.GenerateStubVersionPath(stub.Controller, *goyaveVersion)
	if err != nil {
		return fmt.Errorf("❌ %s", err.Error())
	}

	templateData, err := stub.Load(*stubPath, stub.Data{
		"GoyaveModVersion": *goyaveModVersion,
		"ControllerName":   c.ControllerName,
	})
	if err != nil {
		return fmt.Errorf("❌ %s", err.Error())
	}

	if err := fs.CreatePath(*folderPath); err != nil {
		return err
	}

	err = fs.CreateResourceFile(*folderPath, c.ControllerName, templateData.Bytes())
	if err != nil {
		return fmt.Errorf("❌ %s", err.Error())
	}

	fmt.Println("✅ File Created !")

	return nil
}

func (c *CreateController) Validate() error {
	if c.ControllerName == "" {
		return errors.New("required flag(s) \"controller-name\"")
	}

	return nil
}

func (c *CreateController) UsedFlags() bool {
	controllerNameCheck := false

	for _, arg := range os.Args[1:] {
		if arg == "--controller-name" || arg == "-n" {
			controllerNameCheck = true
		}
	}

	return controllerNameCheck
}

func (c *CreateController) setFlags(flags *pflag.FlagSet) {
	flags.StringVarP(
		&c.ControllerName,
		"controller-name",
		"n",
		"",
		"The name of the controller to generate",
	)
	flags.StringVarP(
		&c.ProjectPath,
		"project-path",
		"p",
		"",
		"The path to the goyave project root",
	)

}
