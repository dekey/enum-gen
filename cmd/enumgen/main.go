package main

import (
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/dekey/enums/generator"
	"github.com/dekey/enums/parser"
	"github.com/dekey/enums/pkg/filesystem"
)

func main() {
	// config ->
	var name string
	var debug bool
	flag.StringVar(&name, "name", "", "Enum base name (e.codeGenerator., Env)")
	flag.BoolVar(&debug, "debug", false, "Enable debug logging")
	flag.Parse()

	if debug {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	if name == "" {
		fail("--name flag is required")
	}

	gopackage := os.Getenv("GOPACKAGE") // package name
	goLine := os.Getenv("GOLINE")
	goFile := os.Getenv("GOFILE")
	if goFile == "" {
		fail("GOFILE is not set; run via `go generate`")
	}

	pkgDir, err := os.Getwd() // current package directory
	if err != nil {
		fail(err.Error())
	}

	slog.Info(
		"Generating enum code",
		slog.String("name", name),
		slog.String("goFile", goFile),
		slog.String("pkgDir", pkgDir),
		slog.String("goLine", goLine),
		slog.String("gopackage", gopackage),
	)

	p := parser.NewParseFromFile()
	g := generator.NewCodeGenerator()

	pkg, consts, err := p.ParseFromFile(pkgDir, goFile, goLine)
	if err != nil {
		fail(err.Error())
	}
	if len(consts) == 0 {
		fail("no constants found to generate from")
	}

	slog.Info(
		"Found constants",
		slog.Int("count", len(consts)),
		slog.String("package", pkg),
		slog.Any("consts", consts),
	)

	out := g.GenerateCode(pkg, name, consts)

	outFile := filepath.Join(pkgDir, fmt.Sprintf("enum_%s_gen.go", strings.ToLower(name)))
	slog.Debug("Writing output", slog.String("outFile", outFile))

	if err := os.WriteFile(outFile, out, 0o600); err != nil {
		fail(fmt.Sprintf("write output: %v", err))
	}

	locator := filesystem.NewLocator()
	modRoot, err := locator.FindRootDir("go.mod", 1)
	if err != nil {
		fail(fmt.Sprintf("determine module root: %v", err))
	}
	modulePath, err := readModulePath(modRoot)
	if err != nil {
		fail(fmt.Sprintf("read module path: %v", err))
	}

	rel, err := relativeDir(modRoot, pkgDir)
	if err != nil {
		fail(fmt.Sprintf("determine relative dir: %v", err))
	}

	slog.Debug(
		"relativeDir",
		slog.String("rel", rel),
		slog.String("gopackage", gopackage),
	)

	enumsImport := modulePath + "/" + rel + "/" + gopackage

	slog.Debug(
		"Enum import",
		slog.String("modRoot", modRoot),
		slog.String("enumsImport", enumsImport),
		slog.String("modulePath", modulePath),
		slog.String("pkgDir", pkgDir),
	)

	if err := g.GenerateTests(pkg, pkgDir, enumsImport, name, consts); err != nil {
		fail(fmt.Sprintf("generate tests: %v", err))
	}
}

func fail(msg string) {
	slog.Error("enum gen:", slog.String("message", msg))

	os.Exit(2)
}

// readModulePath reads the module path from the go.mod at the given root directory.
// It also normalizes common VCS suffixes like ".git".
func readModulePath(root string) (string, error) {
	gomod := filepath.Join(root, "go.mod")
	b, err := os.ReadFile(gomod)
	if err != nil {
		return "", err
	}
	lines := strings.Split(string(b), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			mod := strings.TrimSpace(strings.TrimPrefix(line, "module "))
			// strip quotes if any
			mod = strings.Trim(mod, "\"`")
			// drop trailing .git if present
			mod = strings.TrimSuffix(mod, ".git")
			return mod, nil
		}
	}
	return "", errors.New("module path not found in go.mod")
}

func relativeDir(modRoot, fullpath string) (string, error) {
	slog.Debug("relativeDir", slog.String("modRoot", modRoot), slog.String("fullpath", fullpath))

	rel, err := filepath.Rel(modRoot, fullpath)
	if err != nil {
		return "", err
	}
	return filepath.Dir(rel), nil
}
