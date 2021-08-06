package inject

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"plugin"
	"strings"
	"time"

	"github.com/Masterminds/semver"
	"golang.org/x/mod/modfile"
	"goyave.dev/gyv/internal/mod"
	"goyave.dev/gyv/internal/stub"
)

var (
	minimumGoyaveVersion = semver.MustParse("v3.9.1")

	// ErrUnsupportedGoyaveVersion returned when using NewInjector
	// with a Goyave project using an outdated version of the framework
	ErrUnsupportedGoyaveVersion = fmt.Errorf("Unsupported Goyave version. Minimum version: %s", minimumGoyaveVersion.Original())
)

// FunctionCall is a string representation of a function call or reference
// with its matching import. Doesn't support functions with parameters.
type FunctionCall struct {
	Package *ast.ImportSpec
	Value   string
}

// Injector code injector for Goyave projects. Builds a temporary source file at
// the project's root, build the project in plugin mode and return a Plugin instance.
type Injector struct {
	directory        string
	ModFile          *modfile.File
	GoyaveImportPath string
	GoyaveVersion    *semver.Version

	// StubName the path to the embedded stub used for
	// the temporary source file generation.
	StubName string
	// StubData the data to inject into the stub.
	StubData stub.Data

	// Dependencies list of libraries that need to be imported
	// for the planned injection. These libraries will be added
	// automatically using "go get" and removed after the build is complete.
	Dependencies []string
}

// NewInjector create a new injector for the project in the given directory.
// Can return an error if the given directory is not a Goyave project or its
// version is not supported for injection.
func NewInjector(directory string) (*Injector, error) {
	injector := &Injector{
		directory: directory,
	}
	modFile, err := mod.Parse(directory)
	if err != nil {
		return nil, err
	}

	goyaveMod := mod.FindGoyaveRequire(modFile)
	if goyaveMod == nil {
		return nil, mod.ErrNotAGoyaveProject
	}
	goyaveVersion, err := semver.NewVersion(goyaveMod.Mod.Version)
	if err != nil {
		return nil, err
	}
	if goyaveVersion.LessThan(minimumGoyaveVersion) {
		return nil, ErrUnsupportedGoyaveVersion
	}

	injector.GoyaveVersion = goyaveVersion
	injector.GoyaveImportPath = goyaveMod.Mod.Path
	injector.ModFile = modFile
	return injector, nil
}

// Inject writes temporary source file, compiles plugin, loads it and
// cleans temporary source file and compiled plugin.
func (i *Injector) Inject() (*plugin.Plugin, error) {
	pluginPath := generatePluginPath()
	if err := i.build(pluginPath); err != nil {
		return nil, err
	}
	defer func() {
		if err := os.Remove(pluginPath); err != nil {
			fmt.Println("⚠️ WARNING: could not delete compiled plugin at", pluginPath)
		}
	}()
	return plugin.Open(pluginPath)
}

func (i *Injector) build(output string) error {
	fmt.Println("⚙️ Building plugin")
	fileName := generateTempFileName(i.directory)
	if err := i.writeTemporaryFile(fileName); err != nil {
		return err
	}

	dependencies := i.getDependencies()

	defer func() {
		fmt.Println("🧹 Cleanup")
		if err := os.Remove(fileName); err != nil {
			fmt.Println("⚠️ WARNING: could not delete temporary code injection file", fileName)
			return
		}
		// Remove go.sum unused entries
		if err := i.executeCommand("go", "mod", "tidy"); err != nil {
			fmt.Println("⚠️ WARNING: \"go mod tidy\" failed")
		}
	}()

	for _, d := range dependencies {
		if err := i.executeCommand("go", "get", d); err != nil {
			return err
		}
	}

	cmd := exec.Command("go", "build", "-ldflags", "-w -s", "-buildmode=plugin", "-o", output)
	cmd.Dir = i.directory
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (i *Injector) getDependencies() []string {
	dependencies := make([]string, 0, len(i.Dependencies))
	for _, d := range i.Dependencies {
		if mod.FindDependency(i.ModFile, d) == nil {
			dependencies = append(dependencies, d)
		}
	}
	return dependencies
}

func (i *Injector) executeCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = i.directory
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (i *Injector) writeTemporaryFile(dest string) error {
	data := stub.Data{"GoyaveImportPath": i.GoyaveImportPath}
	for k, v := range i.StubData {
		data[k] = v
	}
	s, err := stub.Load(i.StubName, data)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(dest, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	if _, err := file.WriteString(s.String()); err != nil {
		_ = os.Remove(dest)
		return err
	}

	return file.Close()
}

// ImportToString convert *ast.ImportSpec to a valid import
// statement (without prepending "import "), supporting aliases.
// If the given *ast.ImportSpec is nil, an empty string is returned.
func ImportToString(i *ast.ImportSpec) string {
	if i == nil {
		return ""
	}
	str := ""
	if i.Name != nil {
		str += i.Name.Name + " "
	}

	return str + i.Path.Value
}

// FindExportedFunctions find all exported functions in given directory.
func FindExportedFunctions(directory string) ([]string, error) {
	functions := make([]string, 0, 5)
	files, err := findGoFiles(directory)
	if err != nil {
		return nil, err
	}

	for _, f := range files {
		src, err := os.ReadFile(f)
		if err != nil {
			return nil, err
		}

		fset := token.NewFileSet()

		astFile, err := parser.ParseFile(fset, filepath.Base(f), src, parser.ParseComments)
		if err != nil {
			return nil, err
		}

		ast.Inspect(astFile, func(n ast.Node) bool {
			fn, ok := n.(*ast.FuncDecl)
			if ok && fn.Recv == nil && fn.Type.Results == nil && fn.Type.Params.List == nil && fn.Name.IsExported() {
				functions = append(functions, fn.Name.Name)
				return false
			}
			return true
		})
	}

	return functions, nil
}

func GetBlankImports(directory string) ([]string, error) {
	imports := make([]string, 0, 5)
	files, err := findGoFiles(directory)
	if err != nil {
		return nil, err
	}

	for _, f := range files {
		src, err := os.ReadFile(f)
		if err != nil {
			return nil, err
		}

		fset := token.NewFileSet()

		astFile, err := parser.ParseFile(fset, filepath.Base(f), src, parser.ParseComments)
		if err != nil {
			return nil, err
		}

		for _, i := range astFile.Imports {
			if i.Name != nil && i.Name.Name == "_" {
				fmt.Println(i.Path.Value)
				imports = append(imports, i.Path.Value)
			}
		}
	}

	return imports, nil
}

// FindRouteRegistrer tries to find the route registrer function from Go files
// inside the given directory using the Go AST. Sub-directories are not checked.
// If `goyave.Start()` is found, then the parameter passed to it is assumed to be
// the main route registrer function. Import aliases are supported. To properly identify
// `goyave.Start()`, this function needs the Goyave import path specified in `go.mod`.
func FindRouteRegistrer(directory string, goyaveImportPath string) (*FunctionCall, error) {
	var routeRegister *FunctionCall
	files, err := findGoFiles(directory)
	if err != nil {
		return nil, err
	}

	for _, f := range files {
		src, err := os.ReadFile(f)
		if err != nil {
			return nil, err
		}

		fset := token.NewFileSet()

		astFile, err := parser.ParseFile(fset, filepath.Base(f), src, parser.ParseComments)
		if err != nil {
			return nil, err
		}

		ast.Inspect(astFile, func(n ast.Node) bool {
			if routeRegister != nil {
				return false
			}

			goyaveImportName := "goyave" // TODO make this more generic, it is likely we are going to use other functions such as "seeder.Run()"
			if n := findImportAlias(astFile.Imports, goyaveImportPath); n != "" {
				goyaveImportName = n
			}

			call, ok := n.(*ast.CallExpr)
			if ok {
				if fn, ok := call.Fun.(*ast.SelectorExpr); ok {
					selector, okSelector := fn.X.(*ast.Ident)
					if !okSelector || selector.Name != goyaveImportName || len(call.Args) == 0 || fn.Sel.Name != "Start" {
						return false
					}
					routeRegister = argToFunctionCall(astFile, call.Args[0])
					return false
				}
			}
			return true
		})
	}

	if routeRegister == nil {
		return nil, fmt.Errorf("Could not find any valid call of \"goyave.Start()\"")
	}
	return routeRegister, nil
}

func findImportAlias(imports []*ast.ImportSpec, importPath string) string {
	for _, i := range imports {
		if i.Name != nil && i.Path.Value == importPath {
			return i.Name.Name
		}
	}
	return ""
}

func argToFunctionCall(astFile *ast.File, argExpr ast.Expr) *FunctionCall {
	switch arg := argExpr.(type) {
	case *ast.Ident:
		return &FunctionCall{
			Value: arg.Name,
		}
	case *ast.SelectorExpr:
		if selector, ok := arg.X.(*ast.Ident); ok {
			selectorName := selector.Name

			return &FunctionCall{
				Value:   fmt.Sprintf("%s.%s", selectorName, arg.Sel.Name),
				Package: findImport(astFile.Imports, selectorName),
			}
		}
	case *ast.CallExpr:
		if len(arg.Args) == 0 {
			// If the call has arguments, then we cannot ensure copying these arguments will work
			switch fun := arg.Fun.(type) {
			case *ast.Ident:
				return &FunctionCall{
					Value: fmt.Sprintf("%s()", fun.Name),
				}
			case *ast.SelectorExpr:
				if selector, ok := fun.X.(*ast.Ident); ok {
					selectorName := selector.Name
					return &FunctionCall{
						Value:   fmt.Sprintf("%s.%s()", selectorName, fun.Sel.Name),
						Package: findImport(astFile.Imports, selectorName),
					}
				}
			}
		}
	}
	return nil
}

func findImport(imports []*ast.ImportSpec, name string) *ast.ImportSpec {
	for _, i := range imports {
		path := i.Path.Value[1 : len(i.Path.Value)-1]
		if (i.Name != nil && i.Name.Name == name) || strings.HasSuffix(path, name) {
			// FIXME Suffix method doesn't always work (for example if import path differs from actual package name)
			// This solution is acceptable because this case is supposed to be rare.
			return i
		}
	}
	return nil
}

func findGoFiles(directory string) ([]string, error) {
	files := []string{}
	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if info != nil && info.IsDir() && path != directory {
			return filepath.SkipDir
		}
		if filepath.Ext(path) == ".go" {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func generateTempFileName(parent string) string {
	return fmt.Sprintf("%s%ccodeinject-%d.go", parent, os.PathSeparator, time.Now().Unix())
}

func generatePluginPath() string {
	return fmt.Sprintf("%s%cgyv-code-injection-%d.go", os.TempDir(), os.PathSeparator, time.Now().Unix())
}
