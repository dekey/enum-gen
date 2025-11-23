package app

type CodeGenerator interface {
	GenerateCode(pkg, name string, consts []string) []byte
	GenerateTests(pkg, pkgDir, importPath, name string, consts []string) error
}

type Parser interface {
	ParseFromFile(packageDir string, goFile string, goLine string) (string, []string, error)
}

type Locator interface {
	FindRootDir(file string, skipCaller int) (string, error)
	ReadModulePath(root string) (string, error)
	RelativePackagePath(modRoot string, fullPath string) (string, error)
}
