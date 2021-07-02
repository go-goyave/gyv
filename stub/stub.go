package stub

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"strings"
	"text/template"

	"github.com/Masterminds/semver"
)

//go:embed embed/*
var stubFolder embed.FS

const (
	defaultStub = "default.go.stub"
	Controller  = "embed/controller"
	Middleware  = "embed/middleware"
	Model       = "embed/model"
)

type Data map[string]string

func Load(name string, data Data) (*bytes.Buffer, error) {
	template, err := template.ParseFS(stubFolder, name)
	var writer bytes.Buffer

	if err != nil {
		return nil, err
	}

	if err := template.Execute(&writer, data); err != nil {
		return nil, err
	}

	return &writer, nil
}

func GenerateStubVersionPath(path string, version semver.Version) (*string, error) {
	result := fmt.Sprintf("%s%c%s.go.stub", path, os.PathSeparator, "default")
	lowerThan, err := semver.NewConstraint(fmt.Sprintf("<= %s", version.String()))
	if err != nil {
		return nil, err
	}
	var upperThan *semver.Constraints = nil

	err = fs.WalkDir(stubFolder, path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		data, err := fs.Stat(stubFolder, path)
		if err != nil {
			return err
		}

		if data.Name() == defaultStub || data.IsDir() {
			return nil
		}

		fileVersion, err := semver.NewVersion(strings.Split(data.Name(), ".")[0])
		if err != nil {
			return err
		}

		if !lowerThan.Check(fileVersion) {
			return nil
		}

		if upperThan == nil {
			upperThan, err = semver.NewConstraint(fmt.Sprintf("> %s", fileVersion.String()))
			if err != nil {
				return err
			}

			result = path
			return nil
		}

		if upperThan.Check(fileVersion) {
			upperThan, err = semver.NewConstraint(fmt.Sprintf("> %s", fileVersion.String()))
			if err != nil {
				return err
			}
			result = path
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &result, nil
}
