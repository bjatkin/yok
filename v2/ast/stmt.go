package ast

import "github.com/bjatkin/yok/v2/token"

type Stmt interface {
	stmt()
}

type Program struct {
	Stmt

	Stmts []Stmt
}

type Block struct {
	Stmt

	Stmts []Stmt
}

type If struct {
	Stmt

	Check    Expr
	Body     *Block
	ElseBody *Block
}

type Decl struct {
	Stmt

	Name  token.Token
	Value Expr
}

type Assign struct {
	Stmt

	Name  token.Token
	Value Expr
}

type NewLine struct {
	Stmt
}

type Comment struct {
	Stmt

	Value string
}
