package genyok

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bjatkin/yok/diff"
	"github.com/bjatkin/yok/parser"
)

func TestGenerate(t *testing.T) {
	tests := []struct {
		name     string
		yokFile  string
		wantFile string
	}{
		{
			name:     "if statment",
			yokFile:  "if_dirty.yok",
			wantFile: "if.yok",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source, err := os.ReadFile(filepath.Join("testdata", tt.yokFile))
			if err != nil {
				t.Fatal("Generate() failed to read 'before' file", err)
			}

			p := parser.New(source)
			script, err := p.Parse()
			if err != nil {
				for _, e := range p.Errors {
					t.Errorf("Generate() \terror = %v", e)
				}
				t.Fatalf("Generate() error = %v", err)
			}

			got := Generate(script, source)

			wantFile := filepath.Join("testdata", tt.wantFile)
			if diffs := diff.AgainstFile(t, got, wantFile); diffs != "" {
				t.Errorf("Generate() generated code does not match %s:\n%s", tt.wantFile, diffs)
			}
		})
	}
}
