package main

import (
	"flag"
	"log/slog"
	"os"
	"strings"

	"github.com/dekey/enums/internal/app"
	"github.com/dekey/enums/internal/generator"
	"github.com/dekey/enums/internal/parser"
	"github.com/dekey/go-pkg/filesystem"
)

func main() {
	gopackage := os.Getenv("GOPACKAGE")
	goLine := os.Getenv("GOLINE")
	goFile := os.Getenv("GOFILE")

	var name string
	var enumsPkgName string
	var debug bool
	flag.StringVar(&name, "name", "", "This variable is responsible for naming files and structures")
	flag.StringVar(
		&enumsPkgName,
		"enums-pkg-name",
		gopackage,
		"The package name/alias for the generated package in tests",
	)
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

	p := parser.NewParseFromFile()
	g, err := generator.NewCodeGenerator()
	if err != nil {
		slog.Error("error during code generation", slog.String("message", err.Error()))
		os.Exit(2)
	}
	g.EnumsPkgName = enumsPkgName
	locator := filesystem.NewLocator()

	consoleApp := app.New(g, locator, p)
	if err := consoleApp.Run(name, pkgDir, goFile, goLine); err != nil {
		msg := strings.ReplaceAll(err.Error(), "\n", "")
		msg = strings.ReplaceAll(msg, "\r", "")

		slog.Error(
			"error during code generation",
			slog.String("message", msg),
		)
		os.Exit(2)
	}
}
