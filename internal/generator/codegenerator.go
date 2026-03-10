package generator

import (
	"bytes"
	_ "embed"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"unicode"
)

const defaultEnumsPkgName = "enums"

//go:embed templates/code.tmpl
var codeTemplate string

//go:embed templates/test.tmpl
var testTemplate string

//go:embed templates/base_test.tmpl
var baseTestHelperTemplate string

type CodeGenerator struct {
	codeTemplate           string
	testTemplate           string
	baseTestHelperTemplate string
	EnumsPkgName           string
}

func NewCodeGenerator() *CodeGenerator {
	return &CodeGenerator{
		codeTemplate:           codeTemplate,
		testTemplate:           testTemplate,
		baseTestHelperTemplate: baseTestHelperTemplate,
		EnumsPkgName:           defaultEnumsPkgName,
	}
}

// GenerateCode generates the enum code based on the provided package name, type name, and constants
func (cg *CodeGenerator) GenerateCode(pkg, name string, consts []string) ([]byte, error) {
	upperType := fmt.Sprintf("%sType", cg.exportName(name))
	lowerStruct := fmt.Sprintf("%sTypes", strings.ToLower(name))

	properName := cg.exportName(name)

	lowerName := strings.ToLower(name)
	properStructVar := fmt.Sprintf("%sTypes", properName)

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

	tmpl, err := template.
		New("code").
		Funcs(template.FuncMap{"export": cg.exportName}).
		Parse(cg.codeTemplate)
	if err != nil {
		return nil, fmt.Errorf("parse code template: %w", err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("execute code template: %w", err)
	}

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return nil, fmt.Errorf("format source: %w", err)
	}
	return formatted, nil
}

// GenerateTests generates test code for the enum type and writes it to a file
func (cg *CodeGenerator) GenerateTests(pkg, pkgDir, importPath, name string, consts []string) error {
	properName := cg.exportName(name)
	upperType := fmt.Sprintf("%sType", cg.exportName(name))
	properStructVar := fmt.Sprintf("%sTypes", properName)

	// Build the true cases for each constant
	var casesBuilder strings.Builder
	for _, c := range consts {
		if c == "_" || c == "" {
			continue
		}
		fmt.Fprintf(
			&casesBuilder, "\t\t%s.%s.%s(): true,\n", cg.EnumsPkgName,
			properStructVar,
			cg.exportName(c),
		)
	}

	// Execute enum test template
	data := map[string]any{
		"Pkg":             pkg,
		"ImportPath":      importPath,
		"ProperName":      properName,
		"EnumsPkgName":    cg.EnumsPkgName,
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
		"EnumsPkgName": cg.EnumsPkgName,
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

// exportName uppercases the first rune to create an exported identifier
func (*CodeGenerator) exportName(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}
