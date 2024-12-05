package gensh

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bjatkin/yok/compiler"
	"github.com/bjatkin/yok/diff"
	"github.com/bjatkin/yok/parser"
)

func TestGenerate(t *testing.T) {
	tests := []struct {
		name    string
		yokFile string
		shFile  string
	}{
		{
			name:    "hello world",
			yokFile: "hello_world.yok",
			shFile:  "hello_world.sh",
		},
		{
			name:    "declare variables",
			yokFile: "declare_variables.yok",
			shFile:  "declare_variables.sh",
		},
		{
			name:    "math",
			yokFile: "math.yok",
			shFile:  "math.sh",
		},
		{
			name:    "if",
			yokFile: "if.yok",
			shFile:  "if.sh",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source, err := os.ReadFile(filepath.Join("..", "..", "testdata", tt.yokFile))
			if err != nil {
				t.Fatal("Generate() failed to read source file", err)
			}

			p := parser.New(source)
			script, err := p.Parse()
			if err != nil {
				for _, e := range p.Errors {
					t.Errorf("Generate() \terror = %v", e)
				}
				t.Fatalf("Generate() error = %v", err)
			}

			c := compiler.New(source)
			shAst, err := c.Compile(script)
			if err != nil {
				t.Fatal("Generate() failed to compile source code", err)
			}

			got := Generate(shAst)
			wantFile := filepath.Join("testdata", tt.shFile)
			if diffs := diff.AgainstFile(t, got, wantFile); diffs != "" {
				t.Errorf("Compiler.Compile() generated code does not match %s:\n%s", tt.shFile, diffs)
			}
		})
	}
}
