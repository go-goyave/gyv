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

	"goyave.dev/gyv/internal/fs"
)

// read main.go and find the main route registrer using AST
// can be another file that is not named main.go (find the main function and where goyave.Start is located)

// Create a temp file with current timestamp in name
// containing a function that returns the main route registrer
// Compile in plugin mode
// Use the injected function to generate a router
// Delete the tmp file
// Delete the built plugin (or output it to /tmp?)

// FunctionCall is a string representation of a function call or reference
// with its matching import. Doesn't support functions with parameters.
type FunctionCall struct {
	Package *ast.ImportSpec
	Value   string
}

// Inject a Goyave project into gyv.
// TODO better documentation
func Inject(directory string) (*plugin.Plugin, error) {
	call, err := FindRouteRegistrer(directory)
	if err != nil {
		return nil, err
	}

	// FIXME go.mod parsed twice
	goyaveImportPath, err := fs.GetGoyavePath(directory)
	if err != nil {
		return nil, err
	}
	imports := []string{fmt.Sprintf("\"%s\"", goyaveImportPath)}
	if callImport := importToString(call.Package); callImport != "" {
		imports = append(imports, callImport)
	}

	file := File{
		Package: "main",
		Imports: imports,
		Functions: []Function{
			{
				Name:         "InjectedRouteRegistrer",
				ReturnTypes:  []string{"func(*goyave.Router)"}, // TODO potential future issue for compatibility
				ReturnValues: []string{call.Value},
			},
		},
	}

	fileName := generateTempFileName(directory)
	if err := file.Save(fileName); err != nil {
		return nil, err
	}
	defer func() {
		if err := os.Remove(fileName); err != nil {
			fmt.Println("⚠️ WARNING: could not delete temporary code injection file", fileName)
		}
	}()

	pluginPath := generatePluginPath()
	if err := buildPlugin(directory, pluginPath); err != nil {
		return nil, err
	}
	defer func() {
		if err := os.Remove(pluginPath); err != nil {
			fmt.Println("⚠️ WARNING: could not delete compiled plugin at", pluginPath)
		}
	}()

	// FIXME plugin was built with a different version of package goyave.dev/goyave/v3/helper
	// maybe using reflection that may work?
	// or inject what's supposed to be executed by the CLI inside the plugin as well?
	// Can use hashicorp/go-plugin instead of std plugin
	return plugin.Open(pluginPath)
}

func importToString(i *ast.ImportSpec) string {
	if i == nil {
		return ""
	}
	str := ""
	if i.Name != nil {
		str += i.Name.Name + " "
	}

	return str + i.Path.Value
}

// FindRouteRegistrer tries to find the route registrer function from Go files
// inside the given directory using the Go AST. Sub-directories are not checked.
// If `goyave.Start()` is found, then the parameter passed to it is assumed to be
// the main route registrer function. Import aliases are supported.
func FindRouteRegistrer(directory string) (*FunctionCall, error) {
	var routeRegister *FunctionCall
	files, err := findGoFiles(directory)
	if err != nil {
		return nil, err
	}

	goyaveImportPath, err := fs.GetGoyavePath(directory)
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

			goyaveImportName := "goyave" // TODO make this more generic
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

func buildPlugin(directory, output string) error {
	cmd := exec.Command("go", "build", "-buildmode=plugin", "-o", output)
	cmd.Dir = directory
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
