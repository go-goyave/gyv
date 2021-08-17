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
	// Controller is the path to controller stubs
	Controller = "embed/controller"
	// Middleware is the path to middleware stubs
	Middleware = "embed/middleware"
	// Model is the path to model stubs
	Model = "embed/model"
	// Inject is the path to the inject stubs
	Inject = "embed/inject"
	// InjectOpenAPI is the path to the injected OpenAPI generator stub
	InjectOpenAPI = Inject + "/openapi.go.stub"
	// InjectSeeder is the path to the injected database seed function
	InjectSeeder = Inject + "/seed.go.stub"
	// InjectMigrate is the path to the injected database migration function
	InjectMigrate = Inject + "/migrate.go.stub"
	// InjectDBClear is the path to the injected database clear function
	InjectDBClear = Inject + "/db_clear.go.stub"
)

// Data represent the data to inject inside stub files
type Data map[string]interface{}

// Load load a stub file with injected data
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

// GenerateStubVersionPath return the path to a stub according to a version
func GenerateStubVersionPath(path string, version *semver.Version) (string, error) {
	result := fmt.Sprintf("%s%c%s.go.stub", path, os.PathSeparator, "default")
	lowerThan, err := semver.NewConstraint(fmt.Sprintf("<= %s", version.String()))
	if err != nil {
		return "", err
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
		return "", err
	}

	return result, nil
}
