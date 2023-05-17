package ast

import (
	"strings"

	"github.com/bjatkin/yok/parse"
	"github.com/bjatkin/yok/sym"
)

const indent = "    "

type Client struct {
	table    *sym.Table
	builders []builder
}

func NewClient(table *sym.Table) *Client {
	return &Client{
		table: table,
		builders: []builder{
			buildRoot,
			buildUseImport,
			buildAssign,
			buildComment,
			buildCommandCall,
			buildNewLine,
			buildEnv,
			buildIf,
		},
	}
}

func (c *Client) Build(tree parse.Node) Root {
	ret := Root{}

	var stmts []Stmt
	for _, node := range tree.Nodes {
		tree := c.Build(node)
		stmts = append(stmts, tree.Stmts...)
	}

	for _, builder := range c.builders {
		stmts := builder(c.table, stmts, tree)
		ret.Stmts = append(ret.Stmts, stmts...)
		if len(stmts) > 0 {
			break
		}
	}

	return ret
}

func (c *Client) Yok(tree Root) []byte {
	var raw []string
	for _, stmt := range tree.Stmts {
		raw = append(raw, stmt.Yok()...)
	}

	return []byte(strings.Join(raw, "\n") + "\n")
}

type builder func(*sym.Table, []Stmt, parse.Node) []Stmt

type Stmt interface {
	stmt()
	Yok() []string
}

type Expr interface {
	expr()
	Yok() []string
}
