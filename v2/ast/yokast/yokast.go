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
	Identifier token.Token
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
	Token token.Token
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
	Token token.Token
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
