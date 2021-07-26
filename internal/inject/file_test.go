package inject

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParameterString(t *testing.T) {
	p := Parameter{
		Name: "param",
		Type: "int",
	}
	assert.Equal(t, "param int", p.String())
	assert.Equal(t, 9, p.len())
}
func TestFunctionString(t *testing.T) {
	f := Function{
		Name:         "Inject",
		Parameters:   []Parameter{},
		ReturnTypes:  []string{},
		ReturnValues: []string{},
	}
	assert.Equal(t, "func Inject() {}\n", f.String())
	assert.Equal(t, 17, f.len())

	f.ReturnTypes = []string{"int"}
	f.ReturnValues = []string{"5"}
	assert.Equal(t, "func Inject() int {return 5}\n", f.String())
	assert.Equal(t, 29, f.len())

	f.ReturnTypes = []string{"int", "bool"}
	f.ReturnValues = []string{"5", "true"}
	assert.Equal(t, "func Inject() (int, bool) {return 5, true}\n", f.String())
	assert.Equal(t, 43, f.len())

	f.Parameters = []Parameter{
		{Name: "a", Type: "int"},
		{Name: "b", Type: "bool"},
		{Name: "c", Type: "uint64"},
	}
	assert.Equal(t, "func Inject(a int, b bool, c uint64) (int, bool) {return 5, true}\n", f.String())
	assert.Equal(t, 66, f.len())
}

func TestFileString(t *testing.T) {
	f := File{
		Package:   "main",
		Imports:   []string{},
		Functions: []Function{},
	}

	assert.Equal(t, "package main\n\n", f.String())
	assert.Equal(t, 14, f.len())

	f.Imports = []string{"\"goyave.dev/pak\"", "\"net/http\""}
	f.Functions = []Function{
		{
			Name: "InjectA",
			Parameters: []Parameter{
				{Name: "a", Type: "int"},
				{Name: "b", Type: "bool"},
				{Name: "c", Type: "uint64"},
			},
			ReturnTypes:  []string{"int", "bool"},
			ReturnValues: []string{"5", "true"},
		},
		{
			Name: "InjectB",
			Parameters: []Parameter{
				{Name: "a", Type: "int"},
			},
			ReturnTypes:  []string{"func()"},
			ReturnValues: []string{"pack.AFunction"},
		},
	}

	expected := `package main

import (
	"goyave.dev/pak"
	"net/http"
)

func InjectA(a int, b bool, c uint64) (int, bool) {return 5, true}
func InjectB(a int) func() {return pack.AFunction}
`
	assert.Equal(t, expected, f.String())
	assert.Equal(t, 174, f.len())
}
