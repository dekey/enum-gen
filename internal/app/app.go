package app

import (
	"fmt"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type App struct {
	generator CodeGenerator
	parser    Parser
	locator   Locator
}

func New(generator CodeGenerator, locator Locator, parser Parser) *App {
	return &App{
		generator: generator,
		locator:   locator,
		parser:    parser,
	}
}

func (a *App) Run(name, pkgDir, goFile, goLine, gopackage string) error {
	pkg, consts, err := a.parser.ParseFromFile(pkgDir, goFile, goLine)
	if err != nil {
		return err
	}
	if len(consts) == 0 {
		return fmt.Errorf("no constants found to generate from")
	}

	out, err := a.generator.GenerateCode(pkg, name, consts)
	if err != nil {
		return fmt.Errorf("generate code: %w", err)
	}

	outFile := filepath.Join(pkgDir, fmt.Sprintf("enum_%s_gen.go", strings.ToLower(name)))
	slog.Debug("Writing output", slog.String("outFile", outFile))

	if err := os.WriteFile(outFile, out, 0o600); err != nil {
		return fmt.Errorf("write output: %w", err)
	}

	modRoot, err := a.locator.FindRootDir("go.mod", 1)
	if err != nil {
		return fmt.Errorf("determine module root: %w", err)
	}
	modulePath, err := a.locator.ReadModulePath(modRoot)
	if err != nil {
		return fmt.Errorf("read module path: %w", err)
	}

	rel, err := a.locator.RelativePackagePath(modRoot, pkgDir)
	if err != nil {
		return fmt.Errorf("determine relative dir: %w", err)
	}

	rel = filepath.ToSlash(rel)
	enumsImport := path.Join(modulePath, rel, gopackage)
	slog.Debug(
		"Enum import",
		slog.String("modRoot", modRoot),
		slog.String("enumsImport", enumsImport),
		slog.String("modulePath", modulePath),
		slog.String("pkgDir", pkgDir),
	)

	if err := a.generator.GenerateTests(pkg, pkgDir, enumsImport, name, consts); err != nil {
		return fmt.Errorf("generate tests: %w", err)
	}

	return nil
}
