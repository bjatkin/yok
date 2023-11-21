package token

type Type int

const (
	Unknown = Type(iota)
	Func
	It
	Loop
	Switch
	If
	Else
	Int
	Bool
	String
	Path
	Error
	Struct
	OpenSquare
	CloseSquare
	OpenBrace
	CloseBrace
	OpenParen
	CloseParen
	Assign
	Equal
	NotEqual
	Greater
	Less
	GreaterEqual
	LessEqual
	Bang
	Return
	Break
	Comma
	Minus
	Add
	Star
	Divide
	Var
	Or
	And
	Comment
	NewLine
	Identifyer
	IntLiteral
	StringLiteral
	BoolLiteral
	PathLiteral
	ErrorLiteral
)

type Token struct {
	Start int
	End   int

	Type   Type
	Lexeme string
}

// Empty is an empty token
var Empty = Token{}
