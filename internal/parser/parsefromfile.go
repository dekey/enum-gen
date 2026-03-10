package parser

import (
	"go/ast"
	"go/parser"
	"go/token"
	"log/slog"
	"strconv"
)

type ParseFromFile struct{}

func NewParseFromFile() *ParseFromFile {
	return &ParseFromFile{}
}

func (p *ParseFromFile) ParseFromFile(packageDir string, goFile string, goLine string) (string, []string, error) {
	var consts []string

	fullPath := packageDir + "/" + goFile

	slog.Debug(
		"Parsing file",
		slog.String("fullPath", fullPath),
		slog.String("packageDir", packageDir),
		slog.String("goFile", goFile),
		slog.String("goLine", goLine),
	)

	line, err := strconv.Atoi(goLine)
	if err != nil {
		return "", nil, err
	}

	// parse the Go file
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, fullPath, nil, parser.ParseComments)
	if err != nil {
		return "", nil, err
	}

	pkg := file.Name.Name

	// walk declarations
	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		if genDecl.Tok != token.CONST {
			continue
		}

		// find the const block *after* the //go:generate line
		start := fset.Position(genDecl.Pos()).Line
		if start <= line {
			continue
		}

		slog.Info("Found const block", slog.Int("line", start))

		// extract constants from only the first const block
		for _, spec := range genDecl.Specs {
			valSpec, isValueSpec := spec.(*ast.ValueSpec)
			if !isValueSpec {
				slog.Debug("not a ValueSpec", slog.Any("spec", spec))
				continue
			}

			for _, name := range valSpec.Names {
				consts = append(consts, name.Name)
			}
		}
		break
	}
	return pkg, consts, nil
}
