package inject

import (
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

	injector.Dependencies = append(injector.Dependencies, "goyave.dev/openapi3")
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
