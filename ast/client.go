package ast

import (
	"strings"

	"github.com/bjatkin/yok/parse"
	"github.com/bjatkin/yok/sym"
)

const indent = "    "

type Client struct {
	table      *sym.Table
	builders   []builder
	validators []validator
}

func NewClient(table *sym.Table) *Client {
	return &Client{
		table: table,
		builders: []builder{
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
		validators: []validator{
			NewValidateIdentifyer(),
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

type builder func(*sym.Table, []Stmt, parse.Node) []Stmt

type Stmt interface {
	stmt()
	Yok() []string
}

type Expr interface {
	expr()
	Yok() []string
}
