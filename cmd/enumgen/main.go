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

func main() {
	var name string
	var debug bool
	flag.StringVar(&name, "name", "", "This variable is responsible for naming files and structures")
	flag.BoolVar(&debug, "debug", false, "Enable debug logging")
	flag.Parse()

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

	//nolint:gosec // in CLI context, we want to log the parameters for transparency
	slog.Debug(
		"Generating enum code",
		slog.String("name", name),
		slog.String("goFile", goFile),
		slog.String("pkgDir", pkgDir),
		slog.String("goLine", goLine),
		slog.String("gopackage", gopackage),
	)

	p := parser.NewParseFromFile()
	g := generator.NewCodeGenerator()
	locator := filesystem.NewLocator()

	consoleApp := app.New(g, locator, p)
	if err := consoleApp.Run(name, pkgDir, goFile, goLine, gopackage); err != nil {
		//nolint:gosec // in CLI context, logging errors from the tool is expected
		slog.Error(
			"error during code generation",
			slog.String("message", err.Error()),
		)
		os.Exit(2)
	}
}
