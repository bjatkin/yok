package shast

import (
	"fmt"
	"strconv"
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

// Script represents an entire sh script
type Script struct {
	Node
	Statements []Stmt
}

// Comment is a line comment
type Comment struct {
	Stmt
	Value string
}

// NewLine represents a new line in an sh script
type NewLine struct {
	Stmt
}

// Assign is an variable assignment
type Assign struct {
	Stmt
	Identifier string
	Value      Expr
}

// If is an sh if statement
type If struct {
	Stmt
	Test           *TestCommand
	Statements     []Stmt
	ElseIfs        []ElseIf
	ElseStatements []Stmt
}

// ElseIf is the 'elif' fragment in an if statement
type ElseIf struct {
	Test       *TestCommand
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
	Value string
}

// Redirect is a file redirect for an exec call
type Redirect struct {
	LeftFd  int
	RightFd string
}

// String returns the redirect as a valid sh redirect
func (r Redirect) String() string {
	left := ""
	if r.LeftFd > 0 {
		left = strconv.Itoa(r.LeftFd)
	}

	return fmt.Sprintf("%s>&%s", left, r.RightFd)
}

// Exec executes a command
type Exec struct {
	Expr
	Command   string
	Arguments []Expr
	Redirects []Redirect
}

// Identifier is a sh identifier
type Identifier struct {
	Expr
	Quoted bool
	Value  string
}

// TestCommand is the sh test command
type TestCommand struct {
	Expr
	Expression Expr
}

// ArithmeticCommand represents an arithmetic expression in sh
type ArithmeticCommand struct {
	Expr
	Expression Expr
}

// InfixExpr represents an infix expression in sh
type InfixExpr struct {
	Expr
	Left     Expr
	Operator string
	Right    Expr
}

// GroupExpr represents a grouped expression in sh
type GroupExpr struct {
	Expr
	Expression Expr
}

// CommandSub represents a command substitution
type CommandSub struct {
	Expr
	Expression Expr
}

// ParamaterExpr is a valid Expression that can appear in a ParameterExpansion expression
type ParamaterExpr interface {
	Node
	parameterExpr()
}

// ParameterExpations is a paramater expasion shell call
type ParamaterExpansion struct {
	Expr
	Expression ParamaterExpr
}

// ParameterLength is a ParamaterExpr used to determine the length of the given paramater
type ParameterLength struct {
	ParamaterExpr
	Paramater *Identifier
}

// ParameterReplace is a ParamaterExpr used to do a find and replace on the given paramater
type ParamaterReplace struct {
	ParamaterExpr
	ReplaceAll bool
	Paramater  *Identifier
	Find       Expr
	Replace    Expr
}

// ParamaterRemoveFix is a ParamaterExpr used to remove the prefix or suffix of a string
type ParamaterRemoveFix struct {
	ParamaterExpr
	RemovePrefix bool
	Paramater    *Identifier
	Remove       Expr
}
