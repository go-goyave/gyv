package inject

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
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

func generateTempFileName() string {
	return fmt.Sprintf("codeinject-%d.go", time.Now().Unix())
}

func generatePluginPath() string {
	return fmt.Sprintf("%s%cgyv-code-injection.go", os.TempDir(), os.PathSeparator)
}

func buildPlugin(output string) error {
	cmd := exec.Command("go", "build", "-buildmode=plugin", "-o", output)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
