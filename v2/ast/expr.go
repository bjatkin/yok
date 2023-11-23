package ast

import (
	"github.com/bjatkin/yok/v2/token"
)

type Expr interface {
	expr()
}

type Binary struct {
	Expr

	Left  Expr
	Op    token.Token
	Right Expr
}

type Unary struct {
	Expr

	Op    token.Token
	Right Expr
}

type IntLiteral struct {
	Expr

	Value int64
}

type BoolLiteral struct {
	Expr

	Value bool
}

type StringLiteral struct {
	Expr

	Value string
}

type ErrorLiteral struct {
	Expr

	Value int
}

type PathLiteral struct {
	Expr

	Value string
}

type Identifyer struct {
	Expr

	Name string
}
