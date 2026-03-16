package generator

import (
	"testing"

	"github.com/stretchr/testify/require"
)

const invalidTemplate = "{{.InvalidField"

func TestNewCodeGenerator_TemplateParseError(t *testing.T) {
	testCases := []struct {
		name     string
		codeTmpl string
		testTmpl string
		baseTmpl string
		wantErr  error
	}{
		{
			name:     "code template parse error",
			codeTmpl: invalidTemplate,
			testTmpl: testTemplate,
			baseTmpl: baseTestHelperTemplate,
			wantErr:  ErrParseCodeTemplate,
		},
		{
			name:     "test template parse error",
			codeTmpl: codeTemplate,
			testTmpl: invalidTemplate,
			baseTmpl: baseTestHelperTemplate,
			wantErr:  ErrParseTestTemplate,
		},
		{
			name:     "base template parse error",
			codeTmpl: codeTemplate,
			testTmpl: testTemplate,
			baseTmpl: invalidTemplate,
			wantErr:  ErrParseBaseTemplate,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cg, err := newCodeGeneratorWithTemplates(tc.codeTmpl, tc.testTmpl, tc.baseTmpl)
			require.ErrorIs(t, err, tc.wantErr)
			require.Nil(t, cg)
		})
	}
}
