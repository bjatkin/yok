package compiler

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bjatkin/yok/diff"
	"github.com/bjatkin/yok/parser"
)

func TestCompiler_Compile(t *testing.T) {
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
			name:       "math",
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
			name:       "nested expressions",
			sourceFile: "nested_expressions.yok",
			astFile:    "nested_expressions_ast.txt",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sourceFilePath := filepath.Join("..", "testdata", tt.sourceFile)
			source, err := os.ReadFile(sourceFilePath)
			if err != nil {
				t.Fatal("Compiler.Compile() failed to read source file")
			}

			parser := parser.New(source)
			yokScript, err := parser.Parse()
			if err != nil {
				for _, err := range parser.Errors {
					t.Error("Compiler.Compile() parse errors", err)
				}
				t.Fatal("Compiler.Compile() failed to parse source file")
			}

			compiler := New(source)
			shScript, err := compiler.Compile(yokScript)
			if err != nil {
				t.Error("Compiler.Compile() failed to compile from yok to sh ast", err)
				for _, err := range compiler.errors {
					t.Error("\tcompliation error: ", err)
				}
				return
			}

			got := encodeScript(shScript)
			wantFile := filepath.Join("testdata", tt.astFile)
			if diffs := diff.AgainstFile(t, got, wantFile); diffs != "" {
				t.Errorf("Compiler.Compile() ast does not match %s:\n%s", tt.astFile, diffs)
			}
		})
	}
}
