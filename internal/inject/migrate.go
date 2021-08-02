package inject

import "goyave.dev/gyv/internal/stub"

// Migrate generate and return database migration function.
func Migrate(directory string) (func() error, error) {
	injector, err := NewInjector(directory)
	if err != nil {
		return nil, err
	}

	modelImportPath := injector.ModFile.Module.Mod.Path + "/database/model"
	// TODO what if there are sub-directories for models? Blank import them recursively

	injector.StubName = stub.InjectMigrate

	// TODO if seeder returns error
	// TODO add main.go blank imports
	injector.StubData = stub.Data{
		"ModelImportPath": modelImportPath,
	}

	plug, err := injector.Inject()
	if err != nil {
		return nil, err
	}
	s, err := plug.Lookup("Migrate")
	if err != nil {
		return nil, err
	}
	return s.(func() error), nil
}
