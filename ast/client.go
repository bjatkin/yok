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

// TODO: return error here?
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

func (c *Client) Yok(tree Root) ([]byte, error) {
	var raw []string
	err := tree.Walk(func(s Stmt) error {
		raw = append(raw, s.Yok()...)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return []byte(strings.Join(raw, "\n")), nil
}

type builder func(*sym.Table, []Stmt, parse.Node) []Stmt

type WalkFunc func(Stmt) error

type Walker interface {
	walk(WalkFunc) error
}

type Stmt interface {
	stmt()
	// TODO: remove this and favor the walk function
	Yok() []string
}

type Expr interface {
	expr()
	// TODO: remove this and favor the walk function
	Yok() []string
}
