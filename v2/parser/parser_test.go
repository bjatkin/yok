package parser

import (
	"fmt"
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
			astFile:    "hello_world_ast.txt",
		},
		{
			name:       "declare variables",
			sourceFile: "declare_variables.yok",
			astFile:    "declare_variables_ast.txt",
		},
		{
			name:       "math variables",
			sourceFile: "math.yok",
			astFile:    "math_ast.txt",
		},
		{
			name:       "if",
			sourceFile: "if.yok",
			astFile:    "if_ast.txt",
		},
		{
			name:       "builtin string functions",
			sourceFile: "string_builtins.yok",
			astFile:    "string_builtins_ast.txt",
		},
		{
			name:       "nested expressiosn",
			sourceFile: "nested_expressions.yok",
			astFile:    "nested_expressions_ast.txt",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fmt.Println("running test...", tt.name)
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
