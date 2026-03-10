package app_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	apppkg "github.com/dekey/enums/internal/app"
	"github.com/dekey/enums/internal/app/internal/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestApp_Run(t *testing.T) {
	t.Parallel()

	type args struct {
		enumName string
		pkgDir   string
		goFile   string
		goLine   string
		goPkg    string
	}

	type testCases struct {
		name   string
		args   args
		setup  func(t *testing.T, args args) (*mocks.CodeGenerator, *mocks.Parser, *mocks.Locator, []byte)
		assert func(t *testing.T, tmp string, gen *mocks.CodeGenerator, par *mocks.Parser, loc *mocks.Locator, codeOut []byte, err error)
	}

	mkTmp := func(t *testing.T) string {
		t.Helper()
		d := t.TempDir()
		return d
	}

	cases := []testCases{
		{
			name: "success writes file and generates tests",
			args: args{
				enumName: "Role",
				pkgDir:   mkTmp(t),
				goFile:   "file.go",
				goLine:   "10",
				goPkg:    "enums",
			},
			setup: func(t *testing.T, args args) (*mocks.CodeGenerator, *mocks.Parser, *mocks.Locator, []byte) {
				gen := mocks.NewCodeGenerator(t)
				par := mocks.NewParser(t)
				loc := mocks.NewLocator(t)

				// parser returns pkg and consts
				par.On("ParseFromFile", args.pkgDir, args.goFile, args.goLine).
					Return("foo", []string{"RoleAdmin", "RoleUser"}, nil).
					Once()

				// generator returns code
				codeOut := []byte("generated code for Role")
				gen.On("GenerateCode", "foo", "Role", mock.Anything).Return(codeOut).Once()

				// locator pathing for GenerateTests
				modRoot := filepath.Join(args.pkgDir, "modroot")
				loc.On("FindRootDir", "go.mod", 1).
					Return(modRoot, nil).
					Once()
				loc.On("ReadModulePath", modRoot).
					Return("github.com/example/mod", nil).
					Once()
				loc.On("RelativePackagePath", modRoot, args.pkgDir).
					Return("internal/foo", nil).
					Once()

				// expect GenerateTests called
				gen.
					On("GenerateTests", "foo", args.pkgDir, mock.AnythingOfType("string"), "Role", mock.Anything).
					Return(nil).Once().Run(func(args mock.Arguments) {
					// no-op; we assert later
				})

				return gen, par, loc, codeOut
			},
			assert: func(t *testing.T, tmp string, gen *mocks.CodeGenerator, par *mocks.Parser, loc *mocks.Locator, codeOut []byte, err error) {
				require.NoError(t, err)
				// file must exist with correct name and contents
				out := filepath.Join(tmp, "enum_role_gen.go")
				b, readErr := os.ReadFile(out)
				require.NoError(t, readErr)
				require.Equal(t, string(codeOut), string(b))

				// parser captured args
				par.AssertCalled(t, "ParseFromFile", tmp, "file.go", "10")

				// generator called with expected import
				gen.AssertNumberOfCalls(t, "GenerateTests", 1)
				// import path should use forward slashes and append goPkg
				importPath := "github.com/example/mod/internal/foo/enums"
				// capture the argument from the call history
				calls := gen.Calls
				var gotImport string
				for _, c := range calls {
					if c.Method == "GenerateTests" {
						gotImport = c.Arguments[2].(string)
						// also assert other args
						require.Equal(t, "foo", c.Arguments[0].(string))
						require.Equal(t, "Role", c.Arguments[3].(string))
						break
					}
				}
				// Normalize to forward slashes just in case
				gotImport = strings.ReplaceAll(gotImport, "\\", "/")
				require.Equal(t, importPath, gotImport)
			},
		},
		{
			name: "no constants results in error",
			args: args{
				enumName: "Role",
				pkgDir:   mkTmp(t),
				goFile:   "file.go",
				goLine:   "10",
				goPkg:    "enums",
			},
			setup: func(t *testing.T, args args) (*mocks.CodeGenerator, *mocks.Parser, *mocks.Locator, []byte) {
				gen := mocks.NewCodeGenerator(t)
				par := mocks.NewParser(t)
				loc := mocks.NewLocator(t)

				par.On("ParseFromFile", args.pkgDir, args.goFile, args.goLine).
					Return("foo", []string{}, nil).Once()

				return gen, par, loc, nil
			},
			assert: func(t *testing.T, tmp string, gen *mocks.CodeGenerator, par *mocks.Parser, loc *mocks.Locator, _ []byte, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "no constants found")
			},
		},
		{
			name: "parser error bubbles up",
			args: args{
				enumName: "Role",
				pkgDir:   mkTmp(t),
				goFile:   "file.go",
				goLine:   "10",
				goPkg:    "enums",
			},
			setup: func(t *testing.T, args args) (*mocks.CodeGenerator, *mocks.Parser, *mocks.Locator, []byte) {
				gen := mocks.NewCodeGenerator(t)
				par := mocks.NewParser(t)
				loc := mocks.NewLocator(t)

				par.On("ParseFromFile", args.pkgDir, args.goFile, args.goLine).
					Return("", nil, errors.New("parse fail")).Once()

				return gen, par, loc, nil
			},
			assert: func(t *testing.T, tmp string, gen *mocks.CodeGenerator, par *mocks.Parser, loc *mocks.Locator, _ []byte, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "parse fail")
			},
		},
		{
			name: "write file error when pkgDir does not exist",
			args: args{
				enumName: "Role",
				pkgDir:   filepath.Join(mkTmp(t), "does-not-exist"),
				goFile:   "file.go",
				goLine:   "10",
				goPkg:    "enums",
			},
			setup: func(t *testing.T, args args) (*mocks.CodeGenerator, *mocks.Parser, *mocks.Locator, []byte) {
				gen := mocks.NewCodeGenerator(t)
				par := mocks.NewParser(t)
				loc := mocks.NewLocator(t)

				par.On("ParseFromFile", args.pkgDir, args.goFile, args.goLine).
					Return("foo", []string{"A"}, nil).Once()
				gen.On("GenerateCode", "foo", "Role", mock.Anything).
					Return([]byte("code")).Once()

				return gen, par, loc, nil
			},
			assert: func(t *testing.T, tmp string, gen *mocks.CodeGenerator, par *mocks.Parser, loc *mocks.Locator, _ []byte, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "write output")
			},
		},
		{
			name: "locator FindRootDir error surfaces",
			args: args{
				enumName: "Env",
				pkgDir:   mkTmp(t),
				goFile:   "env.go",
				goLine:   "5",
				goPkg:    "enums",
			},
			setup: func(t *testing.T, args args) (*mocks.CodeGenerator, *mocks.Parser, *mocks.Locator, []byte) {
				gen := mocks.NewCodeGenerator(t)
				par := mocks.NewParser(t)
				loc := mocks.NewLocator(t)

				par.On("ParseFromFile", args.pkgDir, args.goFile, args.goLine).
					Return("foo", []string{"EnvProd"}, nil).
					Once()
				gen.On("GenerateCode", "foo", "Env", mock.Anything).
					Return([]byte("env code")).
					Once()

				loc.On("FindRootDir", "go.mod", 1).
					Return("", errors.New("determine module root")).
					Once()

				return gen, par, loc, []byte("env code")
			},
			assert: func(t *testing.T, tmp string, gen *mocks.CodeGenerator, par *mocks.Parser, loc *mocks.Locator, _ []byte, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "determine module root")
				_, statErr := os.Stat(filepath.Join(tmp, "enum_env_gen.go"))
				require.NoError(t, statErr)
				gen.AssertNumberOfCalls(t, "GenerateTests", 0)
			},
		},
		{
			name: "locator ReadModulePath error surfaces",
			args: args{
				enumName: "Role",
				pkgDir:   mkTmp(t),
				goFile:   "file.go",
				goLine:   "10",
				goPkg:    "enums",
			},
			setup: func(t *testing.T, args args) (*mocks.CodeGenerator, *mocks.Parser, *mocks.Locator, []byte) {
				gen := mocks.NewCodeGenerator(t)
				par := mocks.NewParser(t)
				loc := mocks.NewLocator(t)

				par.On("ParseFromFile", args.pkgDir, args.goFile, args.goLine).
					Return("foo", []string{"X"}, nil).
					Once()
				gen.On("GenerateCode", "foo", "Role", mock.Anything).
					Return([]byte("code")).
					Once()

				modRoot := filepath.Join(args.pkgDir, "modroot")
				loc.On("FindRootDir", "go.mod", 1).
					Return(modRoot, nil).
					Once()
				loc.On("ReadModulePath", modRoot).
					Return("", errors.New("read module path")).
					Once()

				return gen, par, loc, nil
			},
			assert: func(t *testing.T, tmp string, gen *mocks.CodeGenerator, par *mocks.Parser, loc *mocks.Locator, _ []byte, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "read module path")
				gen.AssertNumberOfCalls(t, "GenerateTests", 0)
			},
		},
		{
			name: "locator RelativePackagePath error surfaces",
			args: args{
				enumName: "Role",
				pkgDir:   mkTmp(t),
				goFile:   "file.go",
				goLine:   "10",
				goPkg:    "enums",
			},
			setup: func(t *testing.T, args args) (*mocks.CodeGenerator, *mocks.Parser, *mocks.Locator, []byte) {
				gen := mocks.NewCodeGenerator(t)
				par := mocks.NewParser(t)
				loc := mocks.NewLocator(t)

				par.On("ParseFromFile", args.pkgDir, args.goFile, args.goLine).
					Return("foo", []string{"X"}, nil).
					Once()
				gen.On("GenerateCode", "foo", "Role", mock.Anything).
					Return([]byte("code")).
					Once()

				modRoot := filepath.Join(args.pkgDir, "modroot")
				loc.On("FindRootDir", "go.mod", 1).
					Return(modRoot, nil).
					Once()
				loc.On("ReadModulePath", modRoot).
					Return("github.com/example/mod", nil).
					Once()
				loc.On("RelativePackagePath", modRoot, args.pkgDir).
					Return("", errors.New("determine relative dir")).
					Once()

				return gen, par, loc, nil
			},
			assert: func(t *testing.T, tmp string, gen *mocks.CodeGenerator, par *mocks.Parser, loc *mocks.Locator, _ []byte, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "determine relative dir")
				gen.AssertNumberOfCalls(t, "GenerateTests", 0)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			gen, par, loc, codeOut := tc.setup(t, tc.args)

			a := apppkg.New(gen, loc, par)
			err := a.Run(tc.args.enumName, tc.args.pkgDir, tc.args.goFile, tc.args.goLine, tc.args.goPkg)
			tc.assert(t, tc.args.pkgDir, gen, par, loc, codeOut, err)
		})
	}
}
