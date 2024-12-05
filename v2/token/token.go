package token

import (
	"unicode/utf8"
)

// Type is the type of a lexed token
type Type uint

// All of the yok token types
const (
	Invalid Type = iota
	EOF
	Identifier
	Comment
	NewLine

	// Keywords
	LetKeyword
	FnKeyword
	WhileKeyword
	ForKeyword
	AndKeyword
	OrKeyword
	SwitchKeyword
	StdoutKeyword
	StderrKeyword
	StdinKeyword
	UseKeyword
	ReturnKeyword
	ShKeyword
	InKeyword
	MxKeyword
	TestKeyword
	QuoteKeyword
	UnquoteKeyword
	BodyKeyword
	IfKeyword
	ElseKeyword

	// Literals
	StringExpression
	PatternLiteral
	StringLiteral
	Atom

	// Symbols
	Assign
	EqualEqualS
	EqualEqualI
	NotEqualS
	NotEqualI
	Comma
	Plus
	Minus
	Multiply
	Divide
	Mod
	PlusPlus
	MinusMinus
	GreaterThan
	GreaterEqual
	LessThan
	LessEqual
	Pipe
	OpenBrace
	CloseBrace
	OpenParen
	CloseParen
)

var stringerMap = map[Type]string{
	Invalid:          "invalid",
	EOF:              "eof",
	Identifier:       "identifier",
	Comment:          "comment",
	NewLine:          "new_line",
	LetKeyword:       "let",
	FnKeyword:        "fn",
	WhileKeyword:     "while",
	ForKeyword:       "for",
	AndKeyword:       "and",
	OrKeyword:        "or",
	SwitchKeyword:    "switch",
	StdoutKeyword:    "stdout",
	StderrKeyword:    "stderr",
	StdinKeyword:     "stdin",
	UseKeyword:       "use",
	ReturnKeyword:    "return",
	ShKeyword:        "sh",
	InKeyword:        "in",
	MxKeyword:        "mx",
	TestKeyword:      "test",
	QuoteKeyword:     "quote",
	UnquoteKeyword:   "unquote",
	BodyKeyword:      "body",
	IfKeyword:        "if",
	ElseKeyword:      "else",
	StringExpression: "string_expression",
	PatternLiteral:   "pattern",
	StringLiteral:    "string",
	Atom:             "atom",
	Assign:           "assign",
	EqualEqualS:      "equal_equal_s",
	EqualEqualI:      "equal_equal_i",
	NotEqualS:        "not_equal_s",
	NotEqualI:        "not_equal_i",
	Comma:            "comma",
	Plus:             "plus",
	Minus:            "minus",
	Multiply:         "multiply",
	Divide:           "divide",
	Mod:              "mod",
	PlusPlus:         "plus_plus",
	MinusMinus:       "minus_minus",
	GreaterThan:      "greater_than",
	GreaterEqual:     "greater_equal",
	LessThan:         "less_than",
	LessEqual:        "less_or_equal",
	Pipe:             "pipe",
	OpenBrace:        "open_brace",
	CloseBrace:       "close_brace",
	OpenParen:        "open_paren",
	CloseParen:       "close_paren",
}

// String implements the stringer interface for all token types
func (t Type) String() string {
	name, ok := stringerMap[t]
	if ok {
		return name
	}

	return stringerMap[Invalid]
}

// Pos is the position of the token in the src code
type Pos uint64

// FullPosition is the full position of a token in a yok file
type FullPosition struct {
	FileName   string
	LineNumber int
	ColNumber  int
}

// GetFullPosition converts a basic source file Pos into a FullPosition
func GetFullPosition(sourceFile string, source []byte, pos Pos) FullPosition {
	lineNumber := 1
	ColNumber := 0
	for i := Pos(0); i < pos; {
		r, size := utf8.DecodeRune(source[i:])
		i += Pos(size)

		if r == '\n' {
			ColNumber = 0
			lineNumber++
		}

		ColNumber++
	}

	return FullPosition{
		FileName:   sourceFile,
		LineNumber: lineNumber,
		ColNumber:  ColNumber,
	}
}

// Token represents a lexed token
type Token struct {
	Type Type
	Pos  Pos
	Len  int
}

// NewToken creates a new Token with a given length and position
func NewToken(t Type, pos, len int) Token {
	return Token{
		Type: t,
		Pos:  Pos(pos),
		Len:  len,
	}
}

// Value is the string value of the token
func (t Token) Value(source []byte) string {
	start := int(t.Pos)
	return string(source[start : start+t.Len])
}
