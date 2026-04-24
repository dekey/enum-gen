package app

import "errors"

var (
	ErrNoConstants          = errors.New("no constants found to generate from")
	ErrWriteOutput          = errors.New("write output")
	ErrDetermineModuleRoot  = errors.New("determine module root")
	ErrReadModulePath       = errors.New("read module path")
	ErrDetermineRelativeDir = errors.New("determine relative dir")
)

//go:generate mockery --name CodeGenerator --filename codegenerator.go --structname CodeGenerator --output internal/mocks
type CodeGenerator interface {
	GenerateCode(pkg, name string, consts []string) ([]byte, error)
	GenerateTests(pkg string, pkgDir string, importPath string, name string, consts []string) error
}

//go:generate mockery --name Parser --filename parser.go --structname Parser --output internal/mocks
type Parser interface {
	ParseFromFile(packageDir string, goFile string, goLine string) (string, []string, error)
}

//go:generate mockery --name Locator --filename locator.go --structname Locator --output internal/mocks
type Locator interface {
	FindRootDirFrom(startDir string, file string) (string, error)
	ReadModulePath(root string) (string, error)
	RelativePackagePath(modRoot string, fullPath string) (string, error)
}
