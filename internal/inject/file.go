package inject

import (
	"fmt"
	"os"
	"strings"
)

// File representation of temporary file used for code injection.
type File struct {
	Package   string
	Imports   []string
	Functions []Function
}

func (f File) String() string {
	var builder strings.Builder
	builder.Grow(f.len())

	builder.WriteString("package ")
	builder.WriteString(f.Package)
	builder.WriteString("\n\n")

	// Imports
	if len(f.Imports) > 0 {
		builder.WriteString("import (\n")
		for _, i := range f.Imports {
			builder.WriteString("\t\"")
			builder.WriteString(i)
			builder.WriteString("\"\n")
		}
		builder.WriteString(")\n\n")
	}

	for _, function := range f.Functions {
		builder.WriteString(function.String())
	}

	return builder.String()
}

func (f File) len() int {
	length := 10 + len(f.Package) // "package " and "\n\n"

	if len(f.Imports) > 0 {
		length += 12 // "import (\n" and ")\n\n"
		for _, i := range f.Imports {
			length += len(i) + 4 // '\t"' and '"\n'
		}
	}

	for _, function := range f.Functions {
		length += function.len()
	}

	return length
}

// Save write this file to filesystem. If an error occurs during writing,
// the file will be deleted.
func (f File) Save(dest string) error {
	file, err := os.OpenFile(dest, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := file.WriteString(f.String()); err != nil {
		os.Remove(dest)
		return err
	}

	return file.Close()
}

// Function simple representation of a function for code injection.
type Function struct {
	Name         string
	Parameters   []Parameter
	ReturnTypes  []string
	ReturnValues []string
}

func (f Function) String() string {
	var builder strings.Builder
	builder.Grow(f.len())

	builder.WriteString("func ")
	builder.WriteString(f.Name)
	builder.WriteString("(")
	for i, p := range f.Parameters {
		builder.WriteString(p.String())
		if i < len(f.Parameters)-1 {
			builder.WriteString(", ")
		}
	}
	builder.WriteString(") ")

	hasMultipleReturnTypes := len(f.ReturnTypes) > 1
	if hasMultipleReturnTypes {
		builder.WriteString("(")
	}

	for i, t := range f.ReturnTypes {
		builder.WriteString(t)
		if i < len(f.ReturnTypes)-1 {
			builder.WriteString(", ")
		}
	}

	if hasMultipleReturnTypes {
		builder.WriteString(")")
	}

	if len(f.ReturnTypes) > 0 {
		builder.WriteString(" ")
	}
	builder.WriteString("{")
	if len(f.ReturnValues) > 0 {
		builder.WriteString("return ")
	}
	for i, v := range f.ReturnValues {
		builder.WriteString(v)
		if i < len(f.ReturnValues)-1 {
			builder.WriteString(", ")
		}
	}
	builder.WriteString("}\n")

	return builder.String()
}

func (f Function) len() int {
	length := 11 + len(f.Name) // "func " + name + "() {}\n"
	for _, p := range f.Parameters {
		length += p.len()
	}
	if len(f.Parameters) > 1 {
		length += (len(f.Parameters) - 1) * 2 // Commas
	}
	if len(f.ReturnTypes) > 0 {
		length += 8 // "return " and the space after the return types
	}
	for _, t := range f.ReturnTypes {
		length += len(t)
	}

	if len(f.ReturnTypes) > 1 {
		length += (len(f.ReturnTypes)-1)*2 + 2 // Commas and parenthesis
	}

	for _, t := range f.ReturnValues {
		length += len(t)
	}
	if len(f.ReturnValues) > 1 {
		length += (len(f.ReturnValues) - 1) * 2 // Commas
	}

	return length
}

// Parameter simple representation of a function parameter used with functions
// involved in code injection.
type Parameter struct {
	Name string
	Type string
}

func (p Parameter) String() string {
	return fmt.Sprintf("%s %s", p.Name, p.Type)
}

func (p Parameter) len() int {
	return len(p.Name) + len(p.Type) + 1
}