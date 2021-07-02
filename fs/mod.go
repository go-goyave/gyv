package fs

import (
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/Masterminds/semver"
	"golang.org/x/mod/modfile"
)

const (
	goModFilename = "go.mod"
)

func IsValidProject(projectPath string) error {
	if projectPath == "" {
		return SetRootWorkingDirectory()
	}

	isGoyaveProject, err := isGoyaveProject(projectPath)
	if err != nil {
		return err
	}

	if !isGoyaveProject {
		return errors.New("The root isn't a goyave project")
	}

	return nil
}

func SetRootWorkingDirectory() error {
	sep := string(os.PathSeparator)
	context, err := os.Getwd()
	if err != nil {
		return err
	}

	haveGomod := false

	for !haveGomod {
		_, err := os.Stat(context)

		if os.IsPermission(err) {
			return errors.New("No go.mod found")
		}
		if err != nil {
			return err
		}

		err = filepath.WalkDir(context, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			info, err := os.Stat(path)
			if err != nil {
				return err
			}

			if info.Name() == "go.mod" {
				haveGomod = true
				return nil
			}

			return nil
		})
		if err != nil {
			return err
		}

		if !haveGomod {
			splitedContext := strings.Split(context, sep)
			contextLength := len(splitedContext) - 1
			context = strings.Join(splitedContext[:contextLength], sep)

			if contextLength <= 1 {
				return errors.New("No go.mod found")
			}
		}

	}

	return os.Chdir(context)
}

func isGoyaveProject(projectPath string) (bool, error) {
	modfile, err := dataFromGoMod(projectPath)
	if err != nil {
		return false, nil
	}

	for _, require := range modfile.Require {
		for _, url := range getGoyaveUrls() {
			if strings.Contains(require.Mod.Path, url) {
				return true, nil
			}
		}
	}

	return false, nil
}

func dataFromGoMod(projectPath string) (*modfile.File, error) {
	var goModPath string

	if projectPath == "" {
		goModPath = goModFilename
	} else {
		goModPath = fmt.Sprintf("%s%c%s", projectPath, os.PathSeparator, goModFilename)
	}

	data, err := ioutil.ReadFile(goModPath)
	if err != nil {
		return nil, err
	}

	modfile, err := modfile.Parse(fmt.Sprintf("%s%c%s", goModPath, os.PathSeparator, goModFilename), data, nil)
	if err != nil {
		return nil, err
	}

	return modfile, nil
}

func getGoyaveUrls() []string {
	return []string{"goyave.dev/goyave", "github.com/System-Glitch/goyave"}
}

func GetGoyaveVersion(projectPath string) (*semver.Version, error) {
	modfile, err := dataFromGoMod(projectPath)
	if err != nil {
		return nil, err
	}

	for _, require := range modfile.Require {
		for _, url := range getGoyaveUrls() {
			if strings.Contains(require.Mod.Path, url) {
				version, err := semver.NewVersion(require.Mod.Version)
				if err != nil {
					return nil, err
				}

				return version, nil
			}
		}
	}

	return nil, errors.New("The root isn't a goyave project")
}

func GetGoyavePath(projectPath string) (*string, error) {
	modfile, err := dataFromGoMod(projectPath)
	if err != nil {
		return nil, err
	}

	for _, require := range modfile.Require {
		for _, url := range getGoyaveUrls() {
			if strings.Contains(require.Mod.Path, url) {
				return &require.Mod.Path, nil
			}
		}
	}

	return nil, errors.New("The root isn't a goyave project")
}
