package inject

import (
	"plugin"

	"goyave.dev/gyv/internal/stub"
)

// OpenAPI3Generator injects openapi3 generator into given
// Goyave project.
// Returns a plugin having the "GenerateOpenAPI() ([]byte, error)" function.
func OpenAPI3Generator(directory string) (*plugin.Plugin, error) {
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
	return injector.Inject()
}
