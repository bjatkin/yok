package ast

import (
	"fmt"
	"strings"

	"github.com/bjatkin/yok/parse"
	"github.com/bjatkin/yok/sym"
)

type Client struct {
	table        *sym.Table
	stmtBuilders []stmtBuilder
	exprBuilders []exprBuilder
	validators   []validator
}

func NewClient(table *sym.Table) *Client {
	return &Client{
		table: table,
		stmtBuilders: []stmtBuilder{
			buildRoot,
			buildUseImport,
			buildAssign,
			buildDecl,
			buildComment,
			buildCommandCall,
			buildNewLine,
			buildEnv,
			buildIf,
		},
		exprBuilders: []exprBuilder{
			buildBinaryExpr,
			buildIdentifyer,
			buildValue,
		},
		validators: []validator{
			NewValidateIdentifyer(),
		},
	}
}

func (c *Client) Build(tree parse.Node) Root {
	var stmts []Stmt
	for _, node := range tree.Nodes {
		tree := c.Build(node)
		stmts = append(stmts, tree.Stmts...)
	}

	ret := Root{}
	gotStmts := c.buildStmt(stmts, tree)
	if len(gotStmts) > 0 {
		ret.Stmts = gotStmts
		return ret
	}

	gotExprs := c.buildExpr(stmts, tree)
	if len(gotExprs) > 0 {
		for _, expr := range gotExprs {
			ret.Stmts = append(ret.Stmts, expr)
		}
		return ret
	}

	return ret
}

func (c *Client) buildStmt(stmts []Stmt, node parse.Node) []Stmt {
	for _, builder := range c.stmtBuilders {
		ret := builder(c.table, stmts, node)
		if len(ret) == 0 {
			continue
		}

		return ret
	}

	return nil
}

func (c *Client) buildExpr(stmts []Stmt, node parse.Node) []Expr {
	for _, builder := range c.exprBuilders {
		ret := builder(c.table, stmts, node)
		if len(ret) == 0 {
			continue
		}

		return ret
	}

	return nil
}

func (c *Client) Yok(tree Root) []byte {
	var raw []string
	for _, stmt := range tree.Stmts {
		raw = append(raw, stmt.Yok().String())
	}

	return []byte(strings.Join(raw, "\n") + "\n")
}

func (c *Client) Validate(stmt Stmt) error {
	switch v := stmt.(type) {
	case Root:
		for _, stmt := range v.Stmts {
			err := c.Validate(stmt)
			if err != nil {
				return err
			}
		}
	case If:
		for _, stmt := range v.Root.Stmts {
			err := c.Validate(stmt)
			if err != nil {
				return err
			}
		}
	}

	for _, validator := range c.validators {
		err := validator.check(stmt)
		if err != nil {
			return err
		}
	}

	return nil
}

type stmtBuilder func(*sym.Table, []Stmt, parse.Node) []Stmt

type exprBuilder func(*sym.Table, []Stmt, parse.Node) []Expr

type Node interface {
	Yok() fmt.Stringer
}

type Stmt interface {
	Node
	stmt()
}

type Expr interface {
	Node
	expr()
	stmt() // any expression can behave as a statment if the return value is ignored
}
