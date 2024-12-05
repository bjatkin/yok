package parser

import (
	"slices"

	"github.com/bjatkin/yok/token"
)

// lexer is a lexer that lexes tokens from a source file
type lexer struct {
	source    []byte
	pos       int
	nextToken token.Token
}

// newLexer creates a new lexer from source code
func newLexer(source []byte) lexer {
	lexer := lexer{
		source:    source,
		pos:       0,
		nextToken: token.Token{},
	}

	// take the first token so that the lexer is ready to use
	lexer.take()

	return lexer
}

// peek returns the next token in the source
func (l *lexer) peek() token.Token {
	return l.nextToken
}

// take lexes the next token in the stream and returns it
func (l *lexer) take() token.Token {
	currentToken := l.nextToken

	// given there are no more tokens to consume, this is the end of the file
	if len(l.source[l.pos:]) == 0 {
		l.nextToken = token.NewToken(token.EOF, l.pos, 0)
		return currentToken
	}

	for isWhitespace(l.source[l.pos]) {
		l.pos++
	}

	// check for tokens that match a single byte
	singleTok, foundSingle := matchSingleToken(l.source[l.pos], l.pos)

	// check for tokens that match exactly two bytes
	if len(l.source) >= 2 {
		tok, found := matchDoubleToken(l.source[l.pos:l.pos+2], l.pos)
		if found {
			l.nextToken = tok
			l.pos += 2
			return currentToken
		}
	}

	// check for tokens that match exactly three bytes
	if len(l.source) >= 3 {
		tok, found := matchTripleToken(l.source[l.pos:l.pos+3], l.pos)
		if found {
			l.nextToken = tok
			l.pos += 3
			return currentToken
		}
	}

	// we have to wait till after checking the double token case in order
	// to correctly match tokens like '==' instead of '='
	if foundSingle {
		l.nextToken = singleTok
		l.pos += 1
		return currentToken
	}

	tok, found := matchIdentifierOrKeyword(l.source[l.pos:], l.pos)
	if found {
		l.nextToken = tok
		l.pos += tok.Len
		return currentToken
	}

	tok, found = matchPatternLiteral(l.source[l.pos:], l.pos)
	if found {
		l.nextToken = tok
		l.pos += tok.Len
		return currentToken
	}

	tok, found = matchStringLiteral(l.source[l.pos:], l.pos)
	if found {
		l.nextToken = tok
		l.pos += tok.Len
		return currentToken
	}

	tok, found = matchAtomLiteral(l.source[l.pos:], l.pos)
	if found {
		l.nextToken = tok
		l.pos += tok.Len
		return currentToken
	}

	tok, found = matchComment(l.source[l.pos:], l.pos)
	if found {
		l.nextToken = tok
		l.pos += tok.Len
		return currentToken
	}

	tok = matchUnknownToken(l.source[l.pos:], l.pos)
	l.nextToken = tok
	l.pos += tok.Len
	return currentToken
}

// isAlpha returns true if the character is a valid alphabetic char
func isAlpha(char byte) bool {
	return (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z')
}

// isNumeric returns true if the character is a valid number char
func isNumeric(char byte) bool {
	return char >= '0' && char <= '9'
}

// isWhitespace returns true if the character is a whitespace character
// does not include newlines since those must be lexed independently
func isWhitespace(char byte) bool {
	return char == ' ' || char == '\t'
}

// matchSingleToken matches all the tokens that are a single byte in length
func matchSingleToken(char byte, pos int) (token.Token, bool) {
	var t token.Type
	switch char {
	case '=':
		t = token.Assign
	case ',':
		t = token.Comma
	case '+':
		t = token.Plus
	case '-':
		t = token.Minus
	case '*':
		t = token.Multiply
	case '/':
		t = token.Divide
	case '%':
		t = token.Mod
	case '>':
		t = token.GreaterThan
	case '<':
		t = token.LessThan
	case '|':
		t = token.Pipe
	case '{':
		t = token.OpenBrace
	case '}':
		t = token.CloseBrace
	case '(':
		t = token.OpenParen
	case ')':
		t = token.CloseParen
	case '\n':
		t = token.NewLine
	default:
		return token.Token{}, false
	}

	return token.NewToken(t, pos, 1), true
}

// matchDoubleToken matches all the tokens that are exactly 2 bytes in length
func matchDoubleToken(chars []byte, pos int) (token.Token, bool) {
	var t token.Type
	switch {
	case slices.Equal(chars, []byte("++")):
		t = token.PlusPlus
	case slices.Equal(chars, []byte("--")):
		t = token.MinusMinus
	case slices.Equal(chars, []byte(">=")):
		t = token.GreaterEqual
	case slices.Equal(chars, []byte("<=")):
		t = token.LessEqual
	case slices.Equal(chars, []byte("\r\n")):
		t = token.NewLine
	default:
		return token.Token{}, false
	}

	return token.NewToken(t, pos, 2), true
}

// matchTripleToken matches all the tokens that are exactly 3 bytes in length
func matchTripleToken(chars []byte, pos int) (token.Token, bool) {
	var t token.Type
	switch {
	case slices.Equal(chars, []byte("==s")):
		t = token.EqualEqualS
	case slices.Equal(chars, []byte("==i")):
		t = token.EqualEqualI
	case slices.Equal(chars, []byte("!=s")):
		t = token.NotEqualS
	case slices.Equal(chars, []byte("!=i")):
		t = token.NotEqualI
	default:
		return token.Token{}, false
	}

	return token.NewToken(t, pos, 3), true
}

// matchIdentifierOrKeyword matches identifiers and keywords in the case the the
// identifier is an exact match for a given keyword
func matchIdentifierOrKeyword(chars []byte, pos int) (token.Token, bool) {
	// identifiers must start with an alphabetic character
	if !isAlpha(chars[0]) {
		return token.Token{}, false
	}

	checkKeyword := true
	i := 1
	for ; i < len(chars); i++ {
		// this byte is an upper or lower case letter
		if isAlpha(chars[i]) {
			continue
		}

		// this byte is a digit between 0 and 9
		if isNumeric(chars[i]) {
			continue
		}

		if chars[i] == '_' {
			checkKeyword = false
			continue
		}

		break
	}

	// we know this can't be a keyword so just return it as an identifier
	if !checkKeyword {
		return token.NewToken(token.Identifier, pos, i), true
	}

	// identifiers might actually be a keyword
	if keyword, ok := matchKeyword(chars[:i], pos); ok {
		return keyword, true
	}

	return token.NewToken(token.Identifier, pos, i), true
}

// matchKeyword takes a valid identifier and checks if the identifier is actually
// a keyword. If so, this function returns the keyword token
func matchKeyword(identifier []byte, pos int) (token.Token, bool) {
	var t token.Type
	// TODO: if we use slices.Equal it will save a string allocation per call. It will be more ugly though.
	// heres and example slices.Equal(identifier, []byte{'l', 'e', 't'})
	switch string(identifier) {
	case "let":
		t = token.LetKeyword
	case "fn":
		t = token.FnKeyword
	case "while":
		t = token.WhileKeyword
	case "for":
		t = token.ForKeyword
	case "and":
		t = token.AndKeyword
	case "or":
		t = token.OrKeyword
	case "switch":
		t = token.SwitchKeyword
	case "stdout":
		t = token.StdoutKeyword
	case "stderr":
		t = token.StderrKeyword
	case "stdin":
		t = token.StdinKeyword
	case "use":
		t = token.UseKeyword
	case "return":
		t = token.ReturnKeyword
	case "sh":
		t = token.ShKeyword
	case "in":
		t = token.InKeyword
	case "mx":
		t = token.MxKeyword
	case "test":
		t = token.TestKeyword
	case "quote":
		t = token.QuoteKeyword
	case "unquote":
		t = token.UnquoteKeyword
	case "body":
		t = token.BodyKeyword
	case "if":
		t = token.IfKeyword
	case "else":
		t = token.ElseKeyword
	default:
		return token.Token{}, false
	}

	return token.NewToken(t, pos, len(identifier)), true
}

// matchPatternLiteral returns a pattern literal token if one is found
// it can also return an Invalid token if a pattern literal is started but contains
// invalid characters or is
func matchPatternLiteral(chars []byte, pos int) (token.Token, bool) {
	if chars[0] != '\'' {
		return token.Token{}, false
	}

	i := 1
	for ; i < len(chars); i++ {
		// this byte is an upper or lower case letter or a number
		if isAlpha(chars[i]) || isNumeric(chars[i]) {
			continue
		}

		// these are the special characters supported by `sh` according to `man sh` on Ubuntu 24
		if chars[i] == '!' ||
			chars[i] == '*' ||
			chars[i] == '?' ||
			chars[i] == '[' ||
			chars[i] == '-' ||
			chars[i] == ']' {
			continue
		}

		if chars[i] == '\'' {
			// TODO: I could validate that pattern here and detect possible issues
			// (e.g. unclosed [) is this the right place to do that or should I defer until the parser?
			return token.NewToken(token.PatternLiteral, pos, i+1), true
		}

		break
	}

	// TODO: I could detect common issues here (e.g. it's a regex, not a pattern)
	// is this the right place to do that or should I defer until the parser?

	// invalid token, pattern was opened but was either not closed
	// or contained an illegal token before being closed
	return token.NewToken(token.Invalid, pos, i), true
}

// matchStringLiteral returns a string literal token if one is found
// it can also return an invalid token if a string literal is started but not closed
func matchStringLiteral(chars []byte, pos int) (token.Token, bool) {
	if chars[0] != '"' {
		return token.Token{}, false
	}

	i := 1
	escape := false
	for ; i < len(chars); i++ {
		if chars[i] == '\\' {
			escape = true
			continue
		}

		if !escape && chars[i] == '"' {
			return token.NewToken(token.StringLiteral, pos, i+1), true
		}

		if escape {
			escape = false
		}

		if chars[i] == '\r' || chars[i] == '\n' {
			break
		}
	}

	// invalid token, string was started but was not closed or it
	// contained an invalid character like \n
	return token.NewToken(token.Invalid, pos, i), true
}

// matchAtomLiteral returns an atom literal token if one if found
// it can also return an invalid token if an atom is started but contains invalid characters
func matchAtomLiteral(chars []byte, pos int) (token.Token, bool) {
	if chars[0] != ':' {
		return token.Token{}, false
	}

	i := 1
	for ; i < len(chars); i++ {
		// valid special characters that can show up in an atom. Most of these are supported
		// so you can use atoms for basic file paths
		if chars[i] == '/' ||
			chars[i] == '.' ||
			chars[i] == '_' ||
			chars[i] == '(' ||
			chars[i] == ')' {
			continue
		}

		// this byte is an upper or lower case letter
		if isAlpha(chars[i]) || isNumeric(chars[i]) {
			continue
		}

		if chars[i] == ' ' || chars[i] == '\r' || chars[i] == '\n' {
			break
		}

		return token.NewToken(token.Invalid, pos, i+1), true
	}

	// invalid token, atom was stared but contained invalid characters
	return token.NewToken(token.Atom, pos, i), true
}

// matchComment returns a comment token if one is found
func matchComment(chars []byte, pos int) (token.Token, bool) {
	if chars[0] != '#' {
		return token.Token{}, false
	}

	i := 1
	for ; i < len(chars); i++ {
		if chars[i] == '\r' || chars[i] == '\n' {
			break
		}
	}

	return token.NewToken(token.Comment, pos, i), true
}

// matchUnknownToken matches any series of bytes until it hits a separator token.
// it's used to move past unknown tokens while lexing
func matchUnknownToken(chars []byte, pos int) token.Token {
	i := 0
	for ; i < len(chars); i++ {
		if chars[i] == ' ' ||
			chars[i] == '\r' ||
			chars[i] == '\n' ||
			chars[i] == '\t' ||
			chars[i] == '#' ||
			chars[i] == '"' ||
			chars[i] == '\'' ||
			chars[i] == ',' {
			break
		}
	}

	return token.NewToken(token.Invalid, pos, i)
}
