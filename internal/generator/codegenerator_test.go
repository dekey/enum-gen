package generator_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode"

	"github.com/dekey/enums/internal/generator"
	"github.com/stretchr/testify/require"
)

func TestGenerateCode_Table(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name   string
		pkg    string
		typ    string
		consts []string
		assert func(t *testing.T, got []byte)
	}{
		{
			name:   "generates code for basic constants",
			pkg:    "foo",
			typ:    "My",
			consts: []string{"A", "B"},
			assert: func(t *testing.T, got []byte) {
				s := string(got)
				require.Contains(t, s, "package foo")
				require.Contains(t, s, "MyType")
				require.Contains(t, s, "A")
				require.Contains(t, s, "B")
				require.Contains(t, s, "MyTypes")
			},
		},
		{
			name:   "filters out underscore and empty constants",
			pkg:    "bar",
			typ:    "Thing",
			consts: []string{"_", "X", ""},
			assert: func(t *testing.T, got []byte) {
				s := string(got)
				require.Contains(t, s, "package bar")
				require.Contains(t, s, "ThingType")
				require.Contains(t, s, "X")
				require.NotContains(t, s, "_")
			},
		},
		{
			name:   "exports type name by uppercasing first rune",
			pkg:    "baz",
			typ:    "example",
			consts: []string{"One"},
			assert: func(t *testing.T, got []byte) {
				s := string(got)
				require.Contains(t, s, "package baz")
				require.Contains(t, s, fmt.Sprintf("%sType", exportName("example")))
				require.Contains(t, s, "One")
			},
		},
		{
			name:   "handles empty type name and no constants",
			pkg:    "empt",
			typ:    "",
			consts: []string{},
			assert: func(t *testing.T, got []byte) {
				s := string(got)
				require.Contains(t, s, "package empt")
				require.Contains(t, s, "Type")
			},
		},
		{
			name:   "uppercases first rune preserving underscore in type name",
			pkg:    "underscore",
			typ:    "my_type",
			consts: []string{"Alpha"},
			assert: func(t *testing.T, got []byte) {
				s := string(got)
				require.Contains(t, s, "package underscore")
				require.Contains(t, s, "My_typeType")
				require.Contains(t, s, "Alpha")
			},
		},
		{
			name:   "keeps already exported type name",
			pkg:    "exported",
			typ:    "Example",
			consts: []string{"One", "Two"},
			assert: func(t *testing.T, got []byte) {
				s := string(got)
				require.Contains(t, s, "package exported")
				require.Contains(t, s, "ExampleType")
				require.Contains(t, s, "One")
				require.Contains(t, s, "Two")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			cg := generator.NewCodeGenerator()
			got := cg.GenerateCode(tc.pkg, tc.typ, tc.consts)
			tc.assert(t, got)
		})
	}
}

func TestGenerateTests_Table(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name       string
		pkg        string
		importPath string
		typ        string
		consts     []string
		assert     func(t *testing.T, err error, files map[string]string)
	}{
		{
			name:       "generates enum and base tests for basic constants",
			pkg:        "foo",
			importPath: "github.com/dekey/enums/internal/foo",
			typ:        "My",
			consts:     []string{"A", "B"},
			assert: func(t *testing.T, err error, files map[string]string) {
				require.NoError(t, err)
				enum := files["enum"]
				base := files["base"]
				require.Contains(t, enum, "enums")
				require.Contains(t, enum, fmt.Sprintf("%sTypes", exportName("My")))
				require.Contains(t, enum, "A")
				require.Contains(t, enum, "B")
				require.Contains(t, base, "doEnumTest")
			},
		},
		{
			name:       "filters out underscore and empty constants",
			pkg:        "bar",
			importPath: "github.com/dekey/enums/internal/bar",
			typ:        "Thing",
			consts:     []string{"_", "X", ""},
			assert: func(t *testing.T, err error, files map[string]string) {
				require.NoError(t, err)
				enum := files["enum"]
				require.Contains(t, enum, fmt.Sprintf("%sTypes", exportName("Thing")))
				require.Contains(t, enum, "X")
				require.NotContains(t, enum, "\"_\"") // ensure underscore constant wasn't emitted
			},
		},
		{
			name:       "handles empty type name and no constants",
			pkg:        "empt",
			importPath: "github.com/dekey/enums/internal/empt",
			typ:        "",
			consts:     []string{},
			assert: func(t *testing.T, err error, files map[string]string) {
				require.NoError(t, err)
				enum := files["enum"]
				base := files["base"]
				require.Contains(t, enum, "Type") // generator composes "Type" even for empty name
				require.Contains(t, base, "doEnumTest")
				require.NotEmpty(t, enum)
			},
		},
		{
			name:       "uppercases first rune preserving underscore in type name",
			pkg:        "underscore",
			importPath: "github.com/dekey/enums/internal/underscore",
			typ:        "my_type",
			consts:     []string{"Alpha"},
			assert: func(t *testing.T, err error, files map[string]string) {
				require.NoError(t, err)
				enum := files["enum"]
				require.Contains(t, enum, fmt.Sprintf("%sTypes", exportName("my_type")))
				require.Contains(t, enum, "Alpha")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			pkgDir := t.TempDir()
			cg := generator.NewCodeGenerator()
			err := cg.GenerateTests(tc.pkg, pkgDir, tc.importPath, tc.typ, tc.consts)

			files := map[string]string{}
			enumFileName := fmt.Sprintf("enum_%s_gen_test.go", strings.ToLower(tc.typ))
			enumPath := filepath.Join(pkgDir, enumFileName)
			if b, e := os.ReadFile(enumPath); e == nil {
				files["enum"] = string(b)
			} else {
				files["enum"] = ""
			}
			basePath := filepath.Join(pkgDir, "base_test.go")
			if b, e := os.ReadFile(basePath); e == nil {
				files["base"] = string(b)
			} else {
				files["base"] = ""
			}

			tc.assert(t, err, files)
		})
	}
}

func exportName(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}
