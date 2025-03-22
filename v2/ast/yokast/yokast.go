package yokast

import (
	"github.com/bjatkin/yok/token"
)

// Node is a valid AST node
type Node interface {
	node()
}

// Stmt is a statement in an AST
type Stmt interface {
	Node
	stmt()
}

// Script represents an entire yok script
type Script struct {
	Node
	Statements []Stmt
}

// Comment is a line comment
type Comment struct {
	Stmt
	Token token.Token
}

// NewLine is a solo new line
type NewLine struct {
	Stmt
}

// Assign is a let statement
type Assign struct {
	Stmt
	Identifier *Identifier
	Value      Expr
}

type If struct {
	Stmt
	Test     Expr
	Body     *Block
	ElseIfs  []ElseIf
	ElseBody *Block
}

type ElseIf struct {
	Test Expr
	Body *Block
}

type Block struct {
	Stmt
	Statements []Stmt
}

// StmtExpr is any statement that consists of a single expression
type StmtExpr struct {
	Stmt
	Expression Expr
}

// Expr is an expression in an AST
type Expr interface {
	Node
	expr()
}

// String is a string literal
type String struct {
	Expr
	value string
	Token token.Token
}

func NewInternalString(value string, token token.Token) *String {
	return &String{
		value: value,
		Token: token,
	}
}

func (s *String) Value(source []byte) string {
	if s.value != "" {
		return s.value
	}

	return s.Token.Value(source)
}

// Atom is an atom
type Atom struct {
	Expr
	Token token.Token
}

// Call is a call expression
type Call struct {
	Expr
	Identifier *Identifier
	Arguments  []Expr
}

// Identifier is a yok identifier
type Identifier struct {
	Expr
	// only used for internal complier itendifier names, otherwise the token should be used
	name  string
	Token token.Token
}

func NewInternalIdentifier(name string, token token.Token) *Identifier {
	return &Identifier{
		name:  name,
		Token: token,
	}
}

func (i *Identifier) Name(source []byte) string {
	if i.name != "" {
		return i.name
	}

	return i.Token.Value(source)
}

// InfixExpr is a yok infix expression
type InfixExpr struct {
	Expr
	Left     Expr
	Operator token.Token
	Right    Expr
}

// GroupExpr is a yok grouped expression
type GroupExpr struct {
	Expr
	Expression Expr
}

// PrefixExpr is a yok prefix expression
type PrefixExpr struct {
	Expr
	Token      token.Token
	Expression Expr
}
