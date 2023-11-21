package parser

import (
	"reflect"
	"testing"

	"github.com/bjatkin/yok/v2/token"
)

func Test_lexer_lex(t *testing.T) {
	type args struct {
		src string
	}
	tests := []struct {
		name    string
		args    args
		want    []token.Token
		wantErr bool
	}{
		{
			"value parse",
			args{
				src: "\nfunc main() {}\n",
			},
			[]token.Token{
				{
					Start:  0,
					End:    1,
					Type:   token.NewLine,
					Lexeme: "\n",
				},
				{
					Start:  1,
					End:    5,
					Type:   token.Func,
					Lexeme: "func",
				},
				{
					Start:  6,
					End:    10,
					Type:   token.Identifyer,
					Lexeme: "main",
				},
				{
					Start:  10,
					End:    11,
					Type:   token.OpenParen,
					Lexeme: "(",
				},
				{
					Start:  11,
					End:    12,
					Type:   token.CloseParen,
					Lexeme: ")",
				},
				{
					Start:  13,
					End:    14,
					Type:   token.OpenBrace,
					Lexeme: "{",
				},
				{
					Start:  14,
					End:    15,
					Type:   token.CloseBrace,
					Lexeme: "}",
				},
				{
					Start:  15,
					End:    16,
					Type:   token.NewLine,
					Lexeme: "\n",
				},
			},
			false,
		},
		{
			"unknown token",
			args{
				src: "\nfunc _main() {}\n",
			},
			[]token.Token{
				{
					Start:  0,
					End:    1,
					Type:   token.NewLine,
					Lexeme: "\n",
				},
				{
					Start:  1,
					End:    5,
					Type:   token.Func,
					Lexeme: "func",
				},
				{
					Start:  6,
					End:    7,
					Type:   token.Unknown,
					Lexeme: "_",
				},
				{
					Start:  7,
					End:    11,
					Type:   token.Identifyer,
					Lexeme: "main",
				},
				{
					Start:  11,
					End:    12,
					Type:   token.OpenParen,
					Lexeme: "(",
				},
				{
					Start:  12,
					End:    13,
					Type:   token.CloseParen,
					Lexeme: ")",
				},
				{
					Start:  14,
					End:    15,
					Type:   token.OpenBrace,
					Lexeme: "{",
				},
				{
					Start:  15,
					End:    16,
					Type:   token.CloseBrace,
					Lexeme: "}",
				},
				{
					Start:  16,
					End:    17,
					Type:   token.NewLine,
					Lexeme: "\n",
				},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := newLexer()
			got := lexer.lex(tt.args.src)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parser.Parse() = %v, want %v", got, tt.want)
			}
		})
	}
}
