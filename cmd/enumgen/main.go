package main

import (
	"flag"
	"log/slog"
	"os"

	"github.com/dekey/enums/internal/app"
	"github.com/dekey/enums/internal/generator"
	"github.com/dekey/enums/internal/parser"
	"github.com/dekey/go-pkg/filesystem"
)

var version = "dev"

func main() {
	var name string
	var enumsPkgName string
	var debug bool
	var showVersion bool
	flag.StringVar(&name, "name", "", "This variable is responsible for naming files and structures")
	flag.StringVar(&enumsPkgName, "enums-pkg-name", "enums", "The package name/alias for enums in tests")
	flag.BoolVar(&debug, "debug", false, "Enable debug logging")
	flag.BoolVar(&showVersion, "version", false, "Show version and exit")
	flag.Parse()

	if showVersion {
		os.Stdout.WriteString(version + "\n")
		os.Exit(0)
	}

	if debug {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	if name == "" {
		slog.Error("error during code generation", slog.String("message", "--name flag is required"))
		os.Exit(2)
	}

	pkgDir, err := os.Getwd()
	if err != nil {
		slog.Error("error during code generation", slog.String("message", err.Error()))
		os.Exit(2)
	}

	gopackage := os.Getenv("GOPACKAGE")
	goLine := os.Getenv("GOLINE")
	goFile := os.Getenv("GOFILE")
	if goFile == "" {
		slog.Error("error during code generation", slog.String("message", "GOFILE is not set; run via `go generate`"))
		os.Exit(2)
	}

	if gopackage == "" {
		slog.Error(
			"error during code generation",
			slog.String("message", "GOPACKAGE is not set; run via `go generate`"),
		)
		os.Exit(2)
	}
	if goLine == "" {
		slog.Error("error during code generation", slog.String("message", "GOLINE is not set; run via `go generate`"))
		os.Exit(2)
	}

	slog.Debug(
		"Generating enum code",
		slog.String("name", name),
		slog.String("goFile", goFile),
		slog.String("pkgDir", pkgDir),
		slog.String("goLine", goLine),
		slog.String("gopackage", gopackage),
	)

	p := parser.NewParseFromFile()
	g, err := generator.NewCodeGenerator()
	if err != nil {
		slog.Error("error during code generation", slog.String("message", err.Error()))
		os.Exit(2)
	}
	g.EnumsPkgName = enumsPkgName
	locator := filesystem.NewLocator()

	consoleApp := app.New(g, locator, p)
	if err := consoleApp.Run(name, pkgDir, goFile, goLine, gopackage); err != nil {
		slog.Error(
			"error during code generation",
			slog.String("message", err.Error()),
		)
		os.Exit(2)
	}
}
