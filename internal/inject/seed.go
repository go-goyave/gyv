package inject

import (
	"path"

	"goyave.dev/gyv/internal/stub"
)

// Seeder generate and return database seed function.
func Seeder(directory string, seeders []string) (func() error, error) {
	injector, err := NewInjector(directory)
	if err != nil {
		return nil, err
	}

	seederImportPath := injector.ModFile.Module.Mod.Path + "/database/seeder"

	injector.StubName = stub.InjectSeeder

	// TODO if seeder returns error
	// TODO add main.go blank imports
	injector.StubData = stub.Data{
		"BlankImports":     []string{"goyave.dev/goyave/v3/database/dialect/sqlite"},
		"SeederImportPath": seederImportPath,
		"SeederPackage":    path.Base(seederImportPath),
		"Seeders":          seeders,
	}

	plug, err := injector.Inject()
	if err != nil {
		return nil, err
	}
	s, err := plug.Lookup("Seed")
	if err != nil {
		return nil, err
	}
	return s.(func() error), nil
}
