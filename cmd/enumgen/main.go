package main

import (
	"bytes"
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"unicode"

	"github.com/dekey/enums/pkg/filesystem"
)

//go:embed code.tmpl
var codeTemplate string

//go:embed test.tmpl
var testTemplate string

//go:embed base_test.tmpl
var baseTestHelperTemplate string

func main() {
	var name string
	flag.StringVar(&name, "name", "", "Enum base name (e.g., Env)")
	flag.Parse()

	if name == "" {
		fail("--name flag is required")
	}

	goline := os.Getenv("GOLINE")
	gofile := os.Getenv("GOFILE")
	if gofile == "" {
		fail("GOFILE is not set; run via `go generate`")
	}

	pkgDir := filepath.Dir(gofile)

	slog.Info(
		"Generating enum code",
		slog.String("name", name),
		slog.String("gofile", gofile),
	)

	pkg, consts, err := parseConstsFromFile(gofile, goline)
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

	out := generateCode(pkg, name, consts)

	outFile := filepath.Join(pkgDir, fmt.Sprintf("enum_%s_gen.go", strings.ToLower(name)))
	if err := os.WriteFile(outFile, out, 0o600); err != nil {
		fail(fmt.Sprintf("write output: %v", err))
	}

	// Generate tests alongside code
	modRoot, err := filesystem.FindRootDir("go.mod", 1)
	if err != nil {
		fail(fmt.Sprintf("determine module root: %v", err))
	}
	modulePath, err := readModulePath(modRoot)
	if err != nil {
		fail(fmt.Sprintf("read module path: %v", err))
	}
	enumsImport := modulePath + "/internal/pkg/enums"

	if err := generateTests(pkg, pkgDir, enumsImport, name, consts); err != nil {
		fail(fmt.Sprintf("generate tests: %v", err))
	}
}

func fail(msg string) {
	slog.Error("enum gen:", slog.String("message", msg))

	os.Exit(2)
}

func parseConstsFromFile(gofile, goline string) (string, []string, error) {
	var consts []string

	pkgdir, err := os.Getwd() // current package directory
	if err != nil {
		return "", nil, err
	}
	slog.Info("parseConstsFromFile", slog.String("pkgdir", pkgdir))
	fullpath := pkgdir + "/" + gofile

	line, err := strconv.Atoi(goline)
	if err != nil {
		return "", nil, err
	}

	// parse the Go file
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, fullpath, nil, parser.ParseComments)
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

		slog.Info("Found const block at", slog.Int("line", start))

		// extract constants only one block `const ()`
		for _, spec := range genDecl.Specs {
			valSpec, isValueSpec := spec.(*ast.ValueSpec)
			if !isValueSpec {
				slog.Debug("not a ValueSpec, got ", slog.Any("spec", spec))
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

// exportName uppercases the first rune to create an exported identifier
func exportName(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
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

func generateCode(pkg, name string, consts []string) []byte {
	upperType := fmt.Sprintf("%sType", name)
	lowerStruct := fmt.Sprintf("%sTypes", strings.ToLower(name))
	properName := exportName(name)
	lowerName := strings.ToLower(name)
	properStructVar := fmt.Sprintf("%sTypes", properName) // e.g., EnvTypes

	// filter consts
	filtered := make([]string, 0, len(consts))
	for _, c := range consts {
		if c == "_" || c == "" {
			continue
		}
		filtered = append(filtered, c)
	}

	data := map[string]any{
		"Pkg":             pkg,
		"UpperType":       upperType,
		"LowerStruct":     lowerStruct,
		"ProperStructVar": properStructVar,
		"ProperName":      properName,
		"LowerName":       lowerName,
		"Consts":          filtered,
	}

	tmpl, err := template.New("code").Funcs(template.FuncMap{"export": exportName}).Parse(codeTemplate)
	if err != nil {
		return []byte("// template parse error: " + err.Error())
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return []byte("// template execute error: " + err.Error())
	}

	// gofmt
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		// If formatting fails, return unformatted to aid debugging
		return buf.Bytes()
	}
	return formatted
}

func generateTests(pkg, pkgDir, importPath, name string, consts []string) error {
	properName := exportName(name)
	upperType := fmt.Sprintf("%sType", name)
	properStructVar := fmt.Sprintf("%sTypes", properName) // EnvTypes

	// Build the true cases for each constant
	var casesBuilder strings.Builder
	for _, c := range consts {
		if c == "_" || c == "" {
			continue
		}
		fmt.Fprintf(&casesBuilder, "\t\t%s.%s.%s(): true,\n", "enums", properStructVar, exportName(c))
	}

	// Execute enum test template
	data := map[string]any{
		"Pkg":             pkg,
		"ImportPath":      importPath,
		"ProperName":      properName,
		"EnumsPkgName":    "enums",
		"UpperType":       upperType,
		"ProperStructVar": properStructVar,
		"Cases":           casesBuilder.String(),
	}

	tmpl, err := template.New("test").Parse(testTemplate)
	if err != nil {
		return fmt.Errorf("parse test template: %w", err)
	}
	var testBuf bytes.Buffer
	if err := tmpl.Execute(&testBuf, data); err != nil {
		return fmt.Errorf("execute test template: %w", err)
	}

	// Write enum specific test
	testFile := filepath.Join(pkgDir, fmt.Sprintf("enum_%s_gen_test.go", strings.ToLower(name)))
	if err := os.WriteFile(testFile, testBuf.Bytes(), 0o600); err != nil {
		return fmt.Errorf("write test file: %w", err)
	}

	// Always (re)write base_test.go to ensure it's up-to-date
	baseData := map[string]any{
		"Pkg":          pkg,
		"ImportPath":   importPath,
		"EnumsPkgName": "enums",
	}
	baseTmpl, err := template.New("base").Parse(baseTestHelperTemplate)
	if err != nil {
		return fmt.Errorf("parse base template: %w", err)
	}
	var baseBuf bytes.Buffer
	if err := baseTmpl.Execute(&baseBuf, baseData); err != nil {
		return fmt.Errorf("execute base template: %w", err)
	}
	baseFile := filepath.Join(pkgDir, "base_test.go")
	if err := os.WriteFile(baseFile, baseBuf.Bytes(), 0o600); err != nil {
		return fmt.Errorf("write base_test: %w", err)
	}
	return nil
}
