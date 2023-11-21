package parser

import (
	"github.com/bjatkin/yok/v2/token"
)

type Lexer struct {
	matchers []matcher
}

// TODO: do we even need a lexer client?
func newLexer() *Lexer {
	return &Lexer{
		matchers: []matcher{
			newStringMatch(token.NewLine, "\n"),
			newStringMatch(token.Func, "func"),
			newStringMatch(token.It, "it"),
			newStringMatch(token.Loop, "loop"),
			newStringMatch(token.Switch, "switch"),
			newStringMatch(token.If, "if"),
			newStringMatch(token.Else, "else"),
			newStringMatch(token.Int, "int"),
			newStringMatch(token.Bool, "bool"),
			newStringMatch(token.String, "string"),
			newStringMatch(token.Path, "path"),
			newStringMatch(token.Error, "error"),
			newStringMatch(token.Struct, "struct"),
			newStringMatch(token.OpenSquare, "["),
			newStringMatch(token.CloseSquare, "]"),
			newStringMatch(token.OpenBrace, "{"),
			newStringMatch(token.CloseBrace, "}"),
			newStringMatch(token.OpenParen, "("),
			newStringMatch(token.CloseParen, ")"),
			newStringMatch(token.Assign, "="),
			newStringMatch(token.Equal, "=="),
			newStringMatch(token.NotEqual, "!="),
			newStringMatch(token.Greater, ">"),
			newStringMatch(token.GreaterEqual, ">="),
			newStringMatch(token.Less, "<"),
			newStringMatch(token.LessEqual, "<="),
			newStringMatch(token.Bang, "!"),
			newStringMatch(token.Return, "return"),
			newStringMatch(token.Break, "break"),
			newStringMatch(token.Comma, ","),
			newStringMatch(token.Minus, "-"),
			newStringMatch(token.Add, "+"),
			newStringMatch(token.Star, "*"),
			newStringMatch(token.Divide, "/"),
			newStringMatch(token.Var, "var"),
			newStringMatch(token.Or, "||"),
			newStringMatch(token.And, "&&"),
			newRegexMatch(token.Comment, "#[^\n]*\n"),
			newRegexMatch(token.IntLiteral, "[0-9]+"),
			newRegexMatch(token.BoolLiteral, "(true|false)"),
			newRegexMatch(token.PathLiteral, `(\.\.?|~){0,1}\/[^ \(\)\[\]\{\}\n]*`),
			newRegexMatch(token.ErrorLiteral, "e[0-9]{1,3}"),
			newRegexMatch(token.Identifyer, "[a-zA-Z][a-zA-Z0-9_]*"),
			newFuncMatch(token.StringLiteral, func(src *stream[rune]) (string, bool) {
				if src.peek() != '"' {
					return "", false
				}

				value := string(src.take())

				var done, escape bool
				for !done {
					switch src.peek() {
					case '"':
						// if we're not in an escape sequence, set done to true
						done = !escape
						value += string(src.take())
					case '\n':
						// strings can not be multi-line
						return "", false
					case '\\':
						escape = true
						value += string(src.take())
						continue
					default:
						value += string(src.take())
					}

					escape = false
				}

				return value, true
			}),
		},
	}
}

func (p *Lexer) lex(src string) []token.Token {
	stream := newStream([]rune(src))

	var tokens []token.Token
	for !stream.isEmpty() {
		// eat all the white space characters
		p.eatWhiteSpace(stream)

		// match the lexeme
		match, ok := p.match(stream)
		if ok {
			tokens = append(tokens, match)
			continue
		}

		// we failed to match, eat runes until we can match again
		tokens = append(tokens, p.getUnknownToken(stream)...)
	}

	return tokens
}

func (p *Lexer) eatWhiteSpace(stream *stream[rune]) {
	for !stream.isEmpty() {
		t := stream.peek()
		if t != ' ' && t != '\t' {
			break
		}

		stream.take()
	}
}

func (p *Lexer) getUnknownToken(stream *stream[rune]) []token.Token {
	if stream.isEmpty() {
		return nil
	}

	unknownToken := token.Token{
		Start: stream.current,
		End:   stream.current,
		Type:  token.Unknown,
	}

	var tokens []token.Token
	for !stream.isEmpty() {
		unknownToken.Lexeme += string(stream.take())
		unknownToken.End += 1

		if match, ok := p.match(stream); ok {
			// we found a match, let's get back to normal lexing
			tokens = append(tokens, unknownToken)
			tokens = append(tokens, match)
			return tokens
		}
	}

	// the stream is empty, return the tokens
	tokens = append(tokens, unknownToken)
	return tokens
}

func (p *Lexer) match(stream *stream[rune]) (token.Token, bool) {
	if stream.isEmpty() {
		return token.Empty, false
	}

	var match token.Token
	var found bool
	for _, matcher := range p.matchers {
		token, ok := matcher.match(stream)
		// look for the longest match, not nessisarily the first
		// though the first match does take presidence for longer sequences
		if ok && len(token.Lexeme) > len(match.Lexeme) {
			match = token
			found = true
		}
	}

	stream.takeN(len(match.Lexeme))
	return match, found
}
