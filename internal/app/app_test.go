package app_test

import (
	"errors"
	"os"
	"path/filepath"
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

	type testCase struct {
		name    string
		args    args
		genMock func(t *testing.T, args args, codeOut []byte) *mocks.CodeGenerator
		parMock func(t *testing.T, args args) *mocks.Parser
		locMock func(t *testing.T, args args) *mocks.Locator
		codeOut []byte
		assert  func(t *testing.T, tmp string, gen *mocks.CodeGenerator, codeOut []byte, err error)
	}

	testCases := []testCase{
		{
			name: "success writes file and generates tests",
			args: args{
				enumName: "Role",
				pkgDir:   t.TempDir(),
				goFile:   "file.go",
				goLine:   "10",
				goPkg:    "enums",
			},
			codeOut: []byte("generated code for Role"),
			genMock: func(t *testing.T, args args, codeOut []byte) *mocks.CodeGenerator {
				gen := mocks.NewCodeGenerator(t)
				gen.On("GenerateCode", "foo", "Role", mock.Anything).Return(codeOut, nil).Once()
				gen.On("GenerateTests", "foo", args.pkgDir, "github.com/example/mod/internal/foo/enums", "Role", mock.Anything).
					Return(nil).
					Once()
				return gen
			},
			parMock: func(t *testing.T, args args) *mocks.Parser {
				par := mocks.NewParser(t)
				par.On("ParseFromFile", args.pkgDir, args.goFile, args.goLine).
					Return("foo", []string{"RoleAdmin", "RoleUser"}, nil).Once()
				return par
			},
			locMock: func(t *testing.T, args args) *mocks.Locator {
				loc := mocks.NewLocator(t)
				modRoot := filepath.Join(args.pkgDir, "modroot")
				loc.On("FindRootDirFrom", args.pkgDir, "go.mod").Return(modRoot, nil).Once()
				loc.On("ReadModulePath", modRoot).Return("github.com/example/mod", nil).Once()
				loc.On("RelativePackagePath", modRoot, args.pkgDir).Return("internal/foo", nil).Once()
				return loc
			},
			assert: func(t *testing.T, tmp string, _ *mocks.CodeGenerator, codeOut []byte, err error) {
				require.NoError(t, err)
				out := filepath.Join(tmp, "enum_role_gen.go")
				b, readErr := os.ReadFile(out)
				require.NoError(t, readErr)
				require.Equal(t, string(codeOut), string(b))
			},
		},
		{
			name: "no constants results in error",
			args: args{
				enumName: "Role",
				pkgDir:   t.TempDir(),
				goFile:   "file.go",
				goLine:   "10",
				goPkg:    "enums",
			},
			genMock: func(t *testing.T, _ args, _ []byte) *mocks.CodeGenerator {
				return mocks.NewCodeGenerator(t)
			},
			parMock: func(t *testing.T, args args) *mocks.Parser {
				par := mocks.NewParser(t)
				par.On("ParseFromFile", args.pkgDir, args.goFile, args.goLine).
					Return("foo", []string{}, nil).Once()
				return par
			},
			locMock: func(t *testing.T, _ args) *mocks.Locator {
				return mocks.NewLocator(t)
			},
			assert: func(t *testing.T, _ string, _ *mocks.CodeGenerator, _ []byte, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "no constants found")
			},
		},
		{
			name: "parser error bubbles up",
			args: args{
				enumName: "Role",
				pkgDir:   t.TempDir(),
				goFile:   "file.go",
				goLine:   "10",
				goPkg:    "enums",
			},
			genMock: func(t *testing.T, _ args, _ []byte) *mocks.CodeGenerator {
				return mocks.NewCodeGenerator(t)
			},
			parMock: func(t *testing.T, args args) *mocks.Parser {
				par := mocks.NewParser(t)
				par.On("ParseFromFile", args.pkgDir, args.goFile, args.goLine).
					Return("", nil, errors.New("parse fail")).Once()
				return par
			},
			locMock: func(t *testing.T, _ args) *mocks.Locator {
				return mocks.NewLocator(t)
			},
			assert: func(t *testing.T, _ string, _ *mocks.CodeGenerator, _ []byte, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "parse fail")
			},
		},
		{
			name: "write file error when pkgDir does not exist",
			args: args{
				enumName: "Role",
				pkgDir:   filepath.Join(t.TempDir(), "does-not-exist"),
				goFile:   "file.go",
				goLine:   "10",
				goPkg:    "enums",
			},
			genMock: func(t *testing.T, _ args, _ []byte) *mocks.CodeGenerator {
				gen := mocks.NewCodeGenerator(t)
				gen.On("GenerateCode", "foo", "Role", mock.Anything).
					Return([]byte("code"), nil).Once()
				return gen
			},
			parMock: func(t *testing.T, args args) *mocks.Parser {
				par := mocks.NewParser(t)
				par.On("ParseFromFile", args.pkgDir, args.goFile, args.goLine).
					Return("foo", []string{"A"}, nil).Once()
				return par
			},
			locMock: func(t *testing.T, _ args) *mocks.Locator {
				return mocks.NewLocator(t)
			},
			assert: func(t *testing.T, _ string, _ *mocks.CodeGenerator, _ []byte, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "write output")
			},
		},
		{
			name: "generate code error surfaces",
			args: args{
				enumName: "Role",
				pkgDir:   t.TempDir(),
				goFile:   "file.go",
				goLine:   "10",
				goPkg:    "enums",
			},
			genMock: func(t *testing.T, _ args, _ []byte) *mocks.CodeGenerator {
				gen := mocks.NewCodeGenerator(t)
				gen.On("GenerateCode", "foo", "Role", mock.Anything).
					Return([]byte(nil), errors.New("generate code fail")).Once()
				return gen
			},
			parMock: func(t *testing.T, args args) *mocks.Parser {
				par := mocks.NewParser(t)
				par.On("ParseFromFile", args.pkgDir, args.goFile, args.goLine).
					Return("foo", []string{"A"}, nil).Once()
				return par
			},
			locMock: func(t *testing.T, _ args) *mocks.Locator {
				return mocks.NewLocator(t)
			},
			assert: func(t *testing.T, _ string, _ *mocks.CodeGenerator, _ []byte, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "generate code fail")
			},
		},
		{
			name: "locator FindRootDir error surfaces",
			args: args{
				enumName: "Env",
				pkgDir:   t.TempDir(),
				goFile:   "env.go",
				goLine:   "5",
				goPkg:    "enums",
			},
			codeOut: []byte("env code"),
			genMock: func(t *testing.T, _ args, codeOut []byte) *mocks.CodeGenerator {
				gen := mocks.NewCodeGenerator(t)
				gen.On("GenerateCode", "foo", "Env", mock.Anything).
					Return(codeOut, nil).Once()
				return gen
			},
			parMock: func(t *testing.T, args args) *mocks.Parser {
				par := mocks.NewParser(t)
				par.On("ParseFromFile", args.pkgDir, args.goFile, args.goLine).
					Return("foo", []string{"EnvProd"}, nil).Once()
				return par
			},
			locMock: func(t *testing.T, args args) *mocks.Locator {
				loc := mocks.NewLocator(t)
				loc.On("FindRootDirFrom", args.pkgDir, "go.mod").
					Return("", errors.New("determine module root")).Once()
				return loc
			},
			assert: func(t *testing.T, tmp string, gen *mocks.CodeGenerator, _ []byte, err error) {
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
				pkgDir:   t.TempDir(),
				goFile:   "file.go",
				goLine:   "10",
				goPkg:    "enums",
			},
			genMock: func(t *testing.T, _ args, _ []byte) *mocks.CodeGenerator {
				gen := mocks.NewCodeGenerator(t)
				gen.On("GenerateCode", "foo", "Role", mock.Anything).
					Return([]byte("code"), nil).Once()
				return gen
			},
			parMock: func(t *testing.T, args args) *mocks.Parser {
				par := mocks.NewParser(t)
				par.On("ParseFromFile", args.pkgDir, args.goFile, args.goLine).
					Return("foo", []string{"X"}, nil).Once()
				return par
			},
			locMock: func(t *testing.T, args args) *mocks.Locator {
				loc := mocks.NewLocator(t)
				modRoot := filepath.Join(args.pkgDir, "modroot")
				loc.On("FindRootDirFrom", args.pkgDir, "go.mod").Return(modRoot, nil).Once()
				loc.On("ReadModulePath", modRoot).
					Return("", errors.New("read module path")).Once()
				return loc
			},
			assert: func(t *testing.T, _ string, gen *mocks.CodeGenerator, _ []byte, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "read module path")
				gen.AssertNumberOfCalls(t, "GenerateTests", 0)
			},
		},
		{
			name: "locator RelativePackagePath error surfaces",
			args: args{
				enumName: "Role",
				pkgDir:   t.TempDir(),
				goFile:   "file.go",
				goLine:   "10",
				goPkg:    "enums",
			},
			genMock: func(t *testing.T, _ args, _ []byte) *mocks.CodeGenerator {
				gen := mocks.NewCodeGenerator(t)
				gen.On("GenerateCode", "foo", "Role", mock.Anything).
					Return([]byte("code"), nil).Once()
				return gen
			},
			parMock: func(t *testing.T, args args) *mocks.Parser {
				par := mocks.NewParser(t)
				par.On("ParseFromFile", args.pkgDir, args.goFile, args.goLine).
					Return("foo", []string{"X"}, nil).Once()
				return par
			},
			locMock: func(t *testing.T, args args) *mocks.Locator {
				loc := mocks.NewLocator(t)
				modRoot := filepath.Join(args.pkgDir, "modroot")
				loc.On("FindRootDirFrom", args.pkgDir, "go.mod").Return(modRoot, nil).Once()
				loc.On("ReadModulePath", modRoot).Return("github.com/example/mod", nil).Once()
				loc.On("RelativePackagePath", modRoot, args.pkgDir).
					Return("", errors.New("determine relative dir")).Once()
				return loc
			},
			assert: func(t *testing.T, _ string, gen *mocks.CodeGenerator, _ []byte, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "determine relative dir")
				gen.AssertNumberOfCalls(t, "GenerateTests", 0)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			gen := tc.genMock(t, tc.args, tc.codeOut)
			par := tc.parMock(t, tc.args)
			loc := tc.locMock(t, tc.args)

			a := apppkg.New(gen, loc, par)
			err := a.Run(tc.args.enumName, tc.args.pkgDir, tc.args.goFile, tc.args.goLine, tc.args.goPkg)
			tc.assert(t, tc.args.pkgDir, gen, tc.codeOut, err)
		})
	}
}
