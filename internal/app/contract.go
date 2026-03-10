package app

//go:generate mockery --name CodeGenerator --filename codegenerator.go --structname CodeGenerator --output internal/mocks
type CodeGenerator interface {
	GenerateCode(pkg, name string, consts []string) ([]byte, error)
	GenerateTests(pkg, pkgDir, importPath, name string, consts []string) error
}

//go:generate mockery --name Parser --filename parser.go --structname Parser --output internal/mocks
type Parser interface {
	ParseFromFile(packageDir string, goFile string, goLine string) (string, []string, error)
}

//go:generate mockery --name Locator --filename locator.go --structname Locator --output internal/mocks
type Locator interface {
	FindRootDir(file string, skipCaller int) (string, error)
	ReadModulePath(root string) (string, error)
	RelativePackagePath(modRoot string, fullPath string) (string, error)
}
