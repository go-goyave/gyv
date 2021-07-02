package create

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"goyave.dev/gyv/command"
	"goyave.dev/gyv/fs"
	"goyave.dev/gyv/stub"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type CreateMiddleware struct {
	MiddlewareName string
	ProjectPath    string
}

func (c *CreateMiddleware) BuildCobraCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "middleware",
		Short: "Create goyave middleware",
		Long: `Command for create goyave middleware with the help of a survey.
Only the middleware-name flag is required. The project-path is optional.
Is project-path is empty, the program going to search for a goyave project root.
And for this, the program going to move up to the parent directory and check each time if this directory is a goyave project.
If any parents directories are goyave project, an error will be raised.`,
		RunE: command.GenerateRunFunc(c),
	}

	c.setFlags(cmd.Flags())

	return cmd
}

func (c *CreateMiddleware) BuildSurvey() ([]*survey.Question, error) {
	return []*survey.Question{
		{
			Name:     "middlewareName",
			Prompt:   &survey.Input{Message: "Input the name of the middleware to generate"},
			Validate: survey.Required,
		},
		{
			Name:   "projectPath",
			Prompt: &survey.Input{Message: "Input the path to the goyave project"},
		},
	}, nil
}

func (c *CreateMiddleware) Execute() error {
	if err := fs.IsValidProject(c.ProjectPath); err != nil {
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

	stubPath, err := stub.GenerateStubVersionPath(stub.Middleware, *goyaveVersion)
	if err != nil {
		return fmt.Errorf("❌ %s", err.Error())
	}

	templateData, err := stub.Load(*stubPath, stub.Data{
		"GoyaveModVersion": *goyaveModVersion,
		"MiddlewareName":   strings.Title(c.MiddlewareName),
	})
	if err != nil {
		return fmt.Errorf("❌ %s", err.Error())
	}

	folderPath := fs.CreateMiddlewarePath(c.ProjectPath)

	err = fs.CreateResourceFile(folderPath, c.MiddlewareName, templateData.Bytes())
	if err != nil {
		return fmt.Errorf("❌ %s", err.Error())
	}

	fmt.Println("✅ File Created !")

	return nil
}

func (c *CreateMiddleware) Validate() error {
	if c.MiddlewareName == "" {
		return errors.New("❌ required flag \"middleware-name\"")
	}

	return nil
}

func (c *CreateMiddleware) UsedFlags() bool {
	for _, arg := range os.Args[1:] {
		if arg == "--middleware-name" || arg == "-n" {
			return true
		}
	}

	return false
}

func (c *CreateMiddleware) setFlags(flags *pflag.FlagSet) {
	flags.StringVarP(
		&c.MiddlewareName,
		"middleware-name",
		"n",
		"",
		"The name of the middleware to generate",
	)
	flags.StringVarP(
		&c.ProjectPath,
		"project-path",
		"p",
		"",
		"The path to the goyave project root",
	)
}
