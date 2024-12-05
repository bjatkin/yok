package parser

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bjatkin/yok/diff"
)

func TestParser_Parse(t *testing.T) {
	tests := []struct {
		name       string
		sourceFile string
		astFile    string
	}{
		{
			name:       "hello world",
			sourceFile: "hello_world.yok",
			astFile:    "hello_world_ast.json",
		},
		{
			name:       "declare variables",
			sourceFile: "declare_variables.yok",
			astFile:    "declare_variables_ast.json",
		},
		{
			name:       "math variables",
			sourceFile: "math.yok",
			astFile:    "math_ast.json",
		},
		{
			name:       "if",
			sourceFile: "if.yok",
			astFile:    "if_ast.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source, err := os.ReadFile(filepath.Join("..", "testdata", tt.sourceFile))
			if err != nil {
				t.Fatal("Parser.Parse() failed to read source file", err)
			}

			parser := New(source)
			script, err := parser.Parse()
			if err != nil {
				for _, e := range parser.Errors {
					t.Errorf("Parser.Parse() \terror = %v", e)
				}
				t.Fatalf("Parser.Parse() error = %v", err)
			}

			got := encodeScript(script, source)
			wantFile := filepath.Join("testdata", tt.astFile)
			if diffs := diff.AgainstFile(t, got, wantFile); diffs != "" {
				t.Errorf("Parser.Parse() ast does not match %s:\n%s", tt.astFile, diffs)
			}
		})
	}
}
