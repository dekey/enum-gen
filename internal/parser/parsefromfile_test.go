package parser_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dekey/enums/internal/parser"
	"github.com/stretchr/testify/require"
)

func TestParseFromFile_Table(t *testing.T) {
	testCases := []struct {
		name    string
		content string
		goFile  string
		goLine  string
		assert  func(t *testing.T, pkg string, consts []string, err error)
	}{
		{
			name: "const block after go:generate is found",
			content: `package foo
//go:generate echo
// blank line
const (
 A = iota
 B
)
`,
			goFile: "a.go",
			goLine: "2",
			assert: func(t *testing.T, pkg string, consts []string, err error) {
				require.NoError(t, err)
				require.Equal(t, "foo", pkg)
				require.Equal(t, []string{"A", "B"}, consts)
			},
		},
		{
			name: "no const after given line",
			content: `package bar
const (
 X = 1
)
//go:generate echo
`,
			goFile: "b.go",
			goLine: "4",
			assert: func(t *testing.T, pkg string, consts []string, err error) {
				require.NoError(t, err)
				require.Equal(t, "bar", pkg)
				require.Empty(t, consts)
			},
		},
		{
			name: "parse error returns error",
			content: `package
not valid go`,
			goFile: "c.go",
			goLine: "1",
			assert: func(t *testing.T, pkg string, consts []string, err error) {
				require.Error(t, err)
				require.Empty(t, pkg)
				require.Empty(t, consts)
			},
		},
		{
			name: "single const spec after line (no parens)",
			content: `package zoo
//go:generate echo
const Y = 42
`,
			goFile: "d.go",
			goLine: "2",
			assert: func(t *testing.T, pkg string, consts []string, err error) {
				require.NoError(t, err)
				require.Equal(t, "zoo", pkg)
				require.Equal(t, []string{"Y"}, consts)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			fullPath := filepath.Join(dir, tc.goFile)
			require.NoError(t, os.WriteFile(fullPath, []byte(tc.content), 0o600))

			p := parser.NewParseFromFile()
			pkg, consts, err := p.ParseFromFile(dir, tc.goFile, tc.goLine)
			tc.assert(t, pkg, consts, err)
		})
	}
}
