package generator

import (
	"testing"

	"github.com/stretchr/testify/require"
)

const invalidTemplate = "{{.InvalidField"

func TestNewCodeGenerator_TemplateParseError(t *testing.T) {
	testCases := []struct {
		name    string
		setup   func()
		wantErr error
	}{
		{
			name:    "code template parse error",
			setup:   func() { codeTemplate = invalidTemplate },
			wantErr: ErrParseCodeTemplate,
		},
		{
			name:    "test template parse error",
			setup:   func() { testTemplate = invalidTemplate },
			wantErr: ErrParseTestTemplate,
		},
		{
			name:    "base template parse error",
			setup:   func() { baseTestHelperTemplate = invalidTemplate },
			wantErr: ErrParseBaseTemplate,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			origCode := codeTemplate
			origTest := testTemplate
			origBase := baseTestHelperTemplate
			defer func() {
				codeTemplate = origCode
				testTemplate = origTest
				baseTestHelperTemplate = origBase
			}()

			tc.setup()

			cg, err := NewCodeGenerator()
			require.ErrorIs(t, err, tc.wantErr)
			require.Nil(t, cg)
		})
	}
}
