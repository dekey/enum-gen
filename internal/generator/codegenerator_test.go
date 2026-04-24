package generator_test

import (
	"fmt"
	"go/format"
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
		assert func(t *testing.T, got []byte, err error)
	}{
		{
			name:   "generates code for basic constants",
			pkg:    "foo",
			typ:    "My",
			consts: []string{"A", "B"},
			assert: func(t *testing.T, got []byte, err error) {
				require.NoError(t, err)
				s := string(got)
				require.Contains(t, s, "package foo")
				require.Contains(t, s, "MyType")
				require.Contains(t, s, "A")
				require.Contains(t, s, "B")
				require.Contains(t, s, "MyTypes")
				_, fmtErr := format.Source(got)
				require.NoError(t, fmtErr, "output must be valid Go")
			},
		},
		{
			name:   "filters out underscore and empty constants",
			pkg:    "bar",
			typ:    "Thing",
			consts: []string{"_", "X", ""},
			assert: func(t *testing.T, got []byte, err error) {
				require.NoError(t, err)
				s := string(got)
				require.Contains(t, s, "package bar")
				require.Contains(t, s, "ThingType")
				require.Contains(t, s, "X")
				require.NotContains(t, s, "_")
				_, fmtErr := format.Source(got)
				require.NoError(t, fmtErr, "output must be valid Go")
			},
		},
		{
			name:   "exports type name by uppercasing first rune",
			pkg:    "baz",
			typ:    "example",
			consts: []string{"One"},
			assert: func(t *testing.T, got []byte, err error) {
				require.NoError(t, err)
				s := string(got)
				require.Contains(t, s, "package baz")
				require.Contains(t, s, fmt.Sprintf("%sType", exportName("example")))
				require.Contains(t, s, "One")
				_, fmtErr := format.Source(got)
				require.NoError(t, fmtErr, "output must be valid Go")
			},
		},
		{
			name:   "handles empty type name and no constants",
			pkg:    "empt",
			typ:    "",
			consts: []string{},
			assert: func(t *testing.T, got []byte, err error) {
				require.NoError(t, err)
				s := string(got)
				require.Contains(t, s, "package empt")
				require.Contains(t, s, "Type")
				_, fmtErr := format.Source(got)
				require.NoError(t, fmtErr, "output must be valid Go")
			},
		},
		{
			name:   "uppercases first rune preserving underscore in type name",
			pkg:    "underscore",
			typ:    "my_type",
			consts: []string{"Alpha"},
			assert: func(t *testing.T, got []byte, err error) {
				require.NoError(t, err)
				s := string(got)
				require.Contains(t, s, "package underscore")
				require.Contains(t, s, "My_typeType")
				require.Contains(t, s, "Alpha")
				_, fmtErr := format.Source(got)
				require.NoError(t, fmtErr, "output must be valid Go")
			},
		},
		{
			name:   "keeps already exported type name",
			pkg:    "exported",
			typ:    "Example",
			consts: []string{"One", "Two"},
			assert: func(t *testing.T, got []byte, err error) {
				require.NoError(t, err)
				s := string(got)
				require.Contains(t, s, "package exported")
				require.Contains(t, s, "ExampleType")
				require.Contains(t, s, "One")
				require.Contains(t, s, "Two")
				_, fmtErr := format.Source(got)
				require.NoError(t, fmtErr, "output must be valid Go")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			cg, err := generator.NewCodeGenerator()
			require.NoError(t, err)
			got, err := cg.GenerateCode(tc.pkg, tc.typ, tc.consts)
			tc.assert(t, got, err)
		})
	}
}

func TestGenerateTests_Table(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		pkg        string
		pkgDir     string
		importPath string
		typ        string
		consts     []string
		assert     func(t *testing.T, pkgDir, typ string, err error)
	}{
		{
			name:       "generates enum and base tests for basic constants",
			pkg:        "foo",
			pkgDir:     t.TempDir(),
			importPath: "github.com/dekey/enums/internal/foo",
			typ:        "My",
			consts:     []string{"A", "B"},
			assert: func(t *testing.T, pkgDir, typ string, err error) {
				require.NoError(t, err)
				enum := string(getEnumFileBytes(t, pkgDir, typ))
				base := readFile(t, filepath.Join(pkgDir, "base_test.go"))
				require.Contains(t, enum, "enums")
				require.Contains(t, enum, fmt.Sprintf("%sTypes", exportName(typ)))
				require.Contains(t, enum, "A")
				require.Contains(t, enum, "B")
				require.Contains(t, base, "doEnumTest")
			},
		},
		{
			name:       "filters out underscore and empty constants",
			pkg:        "bar",
			pkgDir:     t.TempDir(),
			importPath: "github.com/dekey/enums/internal/bar",
			typ:        "Thing",
			consts:     []string{"_", "X", ""},
			assert: func(t *testing.T, pkgDir, typ string, err error) {
				require.NoError(t, err)
				enum := string(getEnumFileBytes(t, pkgDir, typ))
				require.Contains(t, enum, fmt.Sprintf("%sTypes", exportName(typ)))
				require.Contains(t, enum, "X")
				require.NotContains(t, enum, "\"_\"")
			},
		},
		{
			name:       "handles empty type name and no constants",
			pkg:        "empt",
			pkgDir:     t.TempDir(),
			importPath: "github.com/dekey/enums/internal/empt",
			typ:        "",
			consts:     []string{},
			assert: func(t *testing.T, pkgDir, typ string, err error) {
				require.NoError(t, err)
				enum := string(getEnumFileBytes(t, pkgDir, typ))
				base := readFile(t, filepath.Join(pkgDir, "base_test.go"))
				require.Contains(t, enum, "Type")
				require.Contains(t, base, "doEnumTest")
				require.NotEmpty(t, enum)
			},
		},
		{
			name:       "uppercases first rune preserving underscore in type name",
			pkg:        "underscore",
			pkgDir:     t.TempDir(),
			importPath: "github.com/dekey/enums/internal/underscore",
			typ:        "my_type",
			consts:     []string{"Alpha"},
			assert: func(t *testing.T, pkgDir, typ string, err error) {
				require.NoError(t, err)
				enum := string(getEnumFileBytes(t, pkgDir, typ))
				require.Contains(t, enum, fmt.Sprintf("%sTypes", exportName(typ)))
				require.Contains(t, enum, "Alpha")
			},
		},
		{
			name:       "exports UpperType in tests when input type is lowercase",
			pkg:        "testpkg",
			pkgDir:     t.TempDir(),
			importPath: "github.com/dekey/enums/internal/testpkg",
			typ:        "example",
			consts:     []string{"A"},
			assert: func(t *testing.T, pkgDir, typ string, err error) {
				require.NoError(t, err)
				enum := string(getEnumFileBytes(t, pkgDir, typ))
				// enums.ExampleType("eggs")
				require.Contains(t, enum, "enums.ExampleType(\"eggs\")")
			},
		},
		{
			name:       "generated files have 0o600 permissions",
			pkg:        "perms",
			pkgDir:     t.TempDir(),
			importPath: "github.com/dekey/enums/internal/perms",
			typ:        "Status",
			consts:     []string{"Active", "Inactive"},
			assert: func(t *testing.T, pkgDir, typ string, err error) {
				require.NoError(t, err)

				enumFile := filepath.Join(pkgDir, fmt.Sprintf("enum_%s_gen_test.go", strings.ToLower(typ)))
				info, statErr := os.Stat(enumFile)
				require.NoError(t, statErr)
				require.Equal(t, os.FileMode(0o600), info.Mode().Perm(), "enum test file must be 0o600")

				baseInfo, statErr := os.Stat(filepath.Join(pkgDir, "base_test.go"))
				require.NoError(t, statErr)
				require.Equal(t, os.FileMode(0o600), baseInfo.Mode().Perm(), "base_test.go must be 0o600")
			},
		},
		{
			name:       "write error when directory does not exist",
			pkg:        "foo",
			pkgDir:     filepath.Join(t.TempDir(), "does-not-exist"),
			importPath: "github.com/dekey/enums/internal/foo",
			typ:        "Role",
			consts:     []string{"A"},
			assert: func(t *testing.T, _ string, _ string, err error) {
				require.ErrorIs(t, err, generator.ErrWriteTestFile)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			cg, err := generator.NewCodeGenerator()
			require.NoError(t, err)
			err = cg.GenerateTests(tc.pkg, tc.pkgDir, tc.importPath, tc.typ, tc.consts)
			tc.assert(t, tc.pkgDir, tc.typ, err)
		})
	}
}

func TestNewCodeGenerator_ReturnsNoError(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		assert func(t *testing.T, cg *generator.CodeGenerator, err error)
	}{
		{
			name: "returns non-nil generator with no error",
			assert: func(t *testing.T, cg *generator.CodeGenerator, err error) {
				require.NoError(t, err)
				require.NotNil(t, cg)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			cg, err := generator.NewCodeGenerator()
			tc.assert(t, cg, err)
		})
	}
}

func TestGenerateTests_OverwritesBaseTestOnSecondCall(t *testing.T) {
	t.Parallel()

	type call struct {
		pkg        string
		importPath string
		typ        string
		consts     []string
	}

	testCases := []struct {
		name   string
		first  call
		second call
		assert func(t *testing.T, pkgDir string, err error)
	}{
		{
			name: "second call overwrites base_test.go package declaration",
			first: call{
				pkg:        "first",
				importPath: "github.com/dekey/enums/internal/first",
				typ:        "Role",
				consts:     []string{"A"},
			},
			second: call{
				pkg:        "second",
				importPath: "github.com/dekey/enums/internal/second",
				typ:        "Env",
				consts:     []string{"B"},
			},
			assert: func(t *testing.T, pkgDir string, err error) {
				require.NoError(t, err)
				content := readFile(t, filepath.Join(pkgDir, "base_test.go"))
				require.Contains(t, content, "package second")
				require.NotContains(t, content, "package first")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			pkgDir := t.TempDir()
			cg, err := generator.NewCodeGenerator()
			require.NoError(t, err)

			err = cg.GenerateTests(tc.first.pkg, pkgDir, tc.first.importPath, tc.first.typ, tc.first.consts)
			require.NoError(t, err)

			err = cg.GenerateTests(tc.second.pkg, pkgDir, tc.second.importPath, tc.second.typ, tc.second.consts)
			tc.assert(t, pkgDir, err)
		})
	}
}

func getEnumFileBytes(t *testing.T, pkgDir string, typ string) []byte {
	t.Helper()
	enumFileName := fmt.Sprintf("enum_%s_gen_test.go", strings.ToLower(typ))
	b, err := os.ReadFile(filepath.Join(pkgDir, enumFileName))
	require.NoError(t, err)
	return b
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	require.NoError(t, err)
	return string(b)
}

func exportName(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}
