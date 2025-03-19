package parser

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/bjatkin/yok/diff"
	"github.com/bjatkin/yok/token"
)

func Test_matchSingleToken(t *testing.T) {
	type args struct {
		char byte
		pos  int
	}
	tests := []struct {
		name   string
		args   args
		want   token.Token
		wantOk bool
	}{
		{
			name: "dollar",
			args: args{
				char: '$',
				pos:  11,
			},
			want:   token.Token{},
			wantOk: false,
		},
		{
			name: "assign",
			args: args{
				char: '=',
				pos:  10,
			},
			want:   token.Token{Type: token.Assign, Pos: 10, Len: 1},
			wantOk: true,
		},
		{
			name: "pipe",
			args: args{
				char: '|',
				pos:  5,
			},
			want:   token.Token{Type: token.Pipe, Pos: 5, Len: 1},
			wantOk: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotOk := matchSingleToken(tt.args.char, tt.args.pos)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("matchSingleToken() got = %v, want %v", got, tt.want)
			}
			if gotOk != tt.wantOk {
				t.Errorf("matchSingleToken() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}

func Test_matchDoubleToken(t *testing.T) {
	type args struct {
		chars []byte
		pos   int
	}
	tests := []struct {
		name   string
		args   args
		want   token.Token
		wantOk bool
	}{
		{
			name: "double bang",
			args: args{
				chars: []byte("!!"),
				pos:   5,
			},
			want:   token.Token{},
			wantOk: false,
		},
		{
			name: "greater than or equal",
			args: args{
				chars: []byte(">="),
				pos:   10,
			},
			want:   token.Token{Type: token.GreaterEqual, Pos: 10, Len: 2},
			wantOk: true,
		},
		{
			name: "windows new line",
			args: args{
				chars: []byte("\r\n"),
				pos:   8,
			},
			want:   token.Token{Type: token.NewLine, Pos: 8, Len: 2},
			wantOk: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotOk := matchDoubleToken(tt.args.chars, tt.args.pos)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("matchDoubleToken() got = %v, want %v", got, tt.want)
			}
			if gotOk != tt.wantOk {
				t.Errorf("matchDoubleToken() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}

func Test_matchIdentifierOrKeyword(t *testing.T) {
	type args struct {
		chars []byte
		pos   int
	}
	tests := []struct {
		name   string
		args   args
		want   token.Token
		wantOk bool
	}{
		{
			name: "invalid identifier",
			args: args{
				chars: []byte("_invalid_"),
				pos:   40,
			},
			want:   token.Token{},
			wantOk: false,
		},
		{
			name: "valid identifier",
			args: args{
				chars: []byte("if_test_1"),
				pos:   10,
			},
			want:   token.Token{Type: token.Identifier, Pos: 10, Len: 9},
			wantOk: true,
		},
		{
			name: "fn keyword",
			args: args{
				chars: []byte("fn"),
				pos:   15,
			},
			want:   token.Token{Type: token.FnKeyword, Pos: 15, Len: 2},
			wantOk: true,
		},
		{
			name: "while keyword",
			args: args{
				chars: []byte("while"),
				pos:   25,
			},
			want:   token.Token{Type: token.WhileKeyword, Pos: 25, Len: 5},
			wantOk: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotOk := matchIdentifierOrKeyword(tt.args.chars, tt.args.pos)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("matchIdentifierOrKeyword() got = %v, want %v", got, tt.want)
			}
			if gotOk != tt.wantOk {
				t.Errorf("matchIdentifierOrKeyword() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}

func Test_matchKeyword(t *testing.T) {
	type args struct {
		identifier []byte
		pos        int
	}
	tests := []struct {
		name   string
		args   args
		want   token.Token
		wantOk bool
	}{
		{
			name: "invalid keyword",
			args: args{
				identifier: []byte("type"),
				pos:        8,
			},
			want:   token.Token{},
			wantOk: false,
		},
		{
			name: "and keyword",
			args: args{
				identifier: []byte("and"),
				pos:        15,
			},
			want:   token.Token{Type: token.AndKeyword, Pos: 15, Len: 3},
			wantOk: true,
		},
		{
			name: "switch keyword",
			args: args{
				identifier: []byte("switch"),
				pos:        23,
			},
			want:   token.Token{Type: token.SwitchKeyword, Pos: 23, Len: 6},
			wantOk: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotOk := matchKeyword(tt.args.identifier, tt.args.pos)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("matchKeyword() got = %v, want %v", got, tt.want)
			}
			if gotOk != tt.wantOk {
				t.Errorf("matchKeyword() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}

func Test_matchPatternLiteral(t *testing.T) {
	type args struct {
		chars []byte
		pos   int
	}
	tests := []struct {
		name   string
		args   args
		want   token.Token
		wantOk bool
	}{
		{
			name: "not a pattern",
			args: args{
				chars: []byte("test"),
				pos:   35,
			},
			want:   token.Token{},
			wantOk: false,
		},
		{
			name: "unclosed pattern",
			args: args{
				chars: []byte("'*pattern"),
				pos:   8,
			},
			want:   token.Token{Type: token.Invalid, Pos: 8, Len: 9},
			wantOk: true,
		},
		{
			name: "valid pattern",
			args: args{
				chars: []byte("'!match'"),
				pos:   12,
			},
			want:   token.Token{Type: token.PatternLiteral, Pos: 12, Len: 8},
			wantOk: true,
		},
		{
			name: "valid star pattern",
			args: args{
				chars: []byte("'*'"),
				pos:   38,
			},
			want:   token.Token{Type: token.PatternLiteral, Pos: 38, Len: 3},
			wantOk: true,
		},
		{
			name: "valid empty pattern",
			args: args{
				chars: []byte("''"),
				pos:   2,
			},
			want:   token.Token{Type: token.PatternLiteral, Pos: 2, Len: 2},
			wantOk: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotOk := matchPatternLiteral(tt.args.chars, tt.args.pos)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("matchPatternLiteral() got = %v, want %v", got, tt.want)
			}
			if gotOk != tt.wantOk {
				t.Errorf("matchPatternLiteral() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}

func Test_matchStringLiteral(t *testing.T) {
	type args struct {
		chars []byte
		pos   int
	}
	tests := []struct {
		name   string
		args   args
		want   token.Token
		wantOk bool
	}{
		{
			name: "invalid string literal",
			args: args{
				chars: []byte("_invalid_"),
				pos:   60,
			},
			want:   token.Token{},
			wantOk: false,
		},
		{
			name: "unclosed string literal",
			args: args{
				chars: []byte(`"unclosed string`),
				pos:   5,
			},
			want:   token.Token{Type: token.Invalid, Pos: 5, Len: 16},
			wantOk: true,
		},
		{
			name: "string with new line",
			args: args{
				chars: []byte("\"test\n\nstring\""),
				pos:   8,
			},
			want:   token.Token{Type: token.Invalid, Pos: 8, Len: 5},
			wantOk: true,
		},
		{
			name: "string literal",
			args: args{
				chars: []byte(`"hello world"`),
				pos:   10,
			},
			want:   token.Token{Type: token.StringLiteral, Pos: 10, Len: 13},
			wantOk: true,
		},
		{
			name: "string literal with escape chars",
			args: args{
				chars: []byte(`"hello \"world\""`),
				pos:   13,
			},
			want:   token.Token{Type: token.StringLiteral, Pos: 13, Len: 17},
			wantOk: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotOk := matchStringLiteral(tt.args.chars, tt.args.pos)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("matchStringLiteral() got = %v, want %v", got, tt.want)
			}
			if gotOk != tt.wantOk {
				t.Errorf("matchStringLiteral() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}

func Test_matchAtomLiteral(t *testing.T) {
	type args struct {
		chars []byte
		pos   int
	}
	tests := []struct {
		name   string
		args   args
		want   token.Token
		wantOk bool
	}{
		{
			name: "not an atom",
			args: args{
				chars: []byte("invalid"),
				pos:   23,
			},
			want:   token.Token{},
			wantOk: false,
		},
		{
			name: "invalid atom",
			args: args{
				chars: []byte(":hello*"),
				pos:   93,
			},
			want:   token.Token{Type: token.Invalid, Pos: 93, Len: 7},
			wantOk: true,
		},
		{
			name: "valid atom",
			args: args{
				chars: []byte(":success"),
				pos:   18,
			},
			want:   token.Token{Type: token.Atom, Pos: 18, Len: 8},
			wantOk: true,
		},
		{
			name: "file path as atom",
			args: args{
				chars: []byte(":/my/file.txt"),
				pos:   10,
			},
			want:   token.Token{Type: token.Atom, Pos: 10, Len: 13},
			wantOk: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotOk := matchAtomLiteral(tt.args.chars, tt.args.pos)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("matchAtomLiteral() got = %v, want %v", got, tt.want)
			}
			if gotOk != tt.wantOk {
				t.Errorf("matchAtomLiteral() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}

func Test_matchComment(t *testing.T) {
	type args struct {
		chars []byte
		pos   int
	}
	tests := []struct {
		name   string
		args   args
		want   token.Token
		wantOk bool
	}{
		{
			name: "not a comment",
			args: args{
				chars: []byte("not a comment"),
				pos:   6,
			},
			want:   token.Token{},
			wantOk: false,
		},
		{
			name: "comment",
			args: args{
				chars: []byte("# this is a # comment \n"),
				pos:   30,
			},
			want:   token.Token{Type: token.Comment, Pos: 30, Len: 22},
			wantOk: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotOk := matchComment(tt.args.chars, tt.args.pos)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("matchComment() got = %v, want %v", got, tt.want)
			}
			if gotOk != tt.wantOk {
				t.Errorf("matchComment() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}

func TestLexFile(t *testing.T) {
	tests := []struct {
		name       string
		sourceFile string
		tokenFile  string
	}{
		{
			name:       "hello world",
			sourceFile: "hello_world.yok",
			tokenFile:  "hello_world_tokens.txt",
		},
		{
			name:       "declare variables",
			sourceFile: "declare_variables.yok",
			tokenFile:  "declare_variables_tokens.txt",
		},
		{
			name:       "math",
			sourceFile: "math.yok",
			tokenFile:  "math_tokens.txt",
		},
		{
			name:       "if",
			sourceFile: "if.yok",
			tokenFile:  "if_tokens.txt",
		},
		{
			name:       "builtin string functions",
			sourceFile: "string_builtins.yok",
			tokenFile:  "string_builtins_tokens.txt",
		},
		{
			name:       "nested expressions",
			sourceFile: "nested_expressions.yok",
			tokenFile:  "nested_expressions_tokens.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sourceFilePath := filepath.Join("..", "testdata", tt.sourceFile)
			source, err := os.ReadFile(sourceFilePath)
			if err != nil {
				t.Fatal("LexFile failed to read source file", err)
			}

			lex := newLexer(source)
			tokens := []token.Token{}
			for lex.peek().Type != token.EOF {
				if lex.peek().Len == 0 {
					t.Fatal("LexFiles token had length of 0")
				}
				tokens = append(tokens, lex.take())
			}

			gotJson := encodeTokens(tokens, source)
			got := gotJson.Render(0)

			wantFile := filepath.Join("testdata", tt.tokenFile)
			if diffs := diff.AgainstFile(t, got, wantFile); diffs != "" {
				t.Errorf("LexFile tokens do not match %s:\n%s", tt.tokenFile, diffs)
			}
		})
	}
}
