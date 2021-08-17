package inject

import "goyave.dev/gyv/internal/stub"

// DBClear generate and return database clear function.
func DBClear(directory string) (func() error, error) {
	injector, err := NewInjector(directory)
	if err != nil {
		return nil, err
	}

	modelImportPath := injector.ModFile.Module.Mod.Path + "/database/model"
	// TODO what if there are sub-directories for models? Blank import them recursively

	injector.StubName = stub.InjectDBClear

	injector.StubData = stub.Data{
		"ModelImportPath": modelImportPath,
	}

	plug, err := injector.Inject()
	if err != nil {
		return nil, err
	}
	s, err := plug.Lookup("DBClear")
	if err != nil {
		return nil, err
	}
	return s.(func() error), nil
}
