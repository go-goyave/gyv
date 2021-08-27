package inject

import (
	"github.com/Masterminds/semver"
	"goyave.dev/gyv/internal/stub"
)

// OpenAPI3Generator injects openapi3 generator into given Goyave project.
// Returns the injected function.
func OpenAPI3Generator(directory string) (func() ([]byte, error), error) {
	injector, err := NewInjector(directory)
	if err != nil {
		return nil, err
	}

	call, err := FindRouteRegistrer(directory, injector.GoyaveImportPath)
	if err != nil {
		return nil, err
	}

	libVersion := ""
	if c, _ := semver.NewConstraint("< v4.0.0-rc1"); c.Check(injector.GoyaveVersion) {
		libVersion = "v0.1.0"
	}
	injector.Dependencies = append(injector.Dependencies, Dependency{"goyave.dev/openapi3", libVersion})
	injector.StubName = stub.InjectOpenAPI
	injector.StubData = stub.Data{"RouteRegistrerImportPath": ImportToString(call.Package)}

	plug, err := injector.Inject()
	if err != nil {
		return nil, err
	}
	s, err := plug.Lookup("GenerateOpenAPI")
	if err != nil {
		return nil, err
	}
	return s.(func() ([]byte, error)), nil
}
