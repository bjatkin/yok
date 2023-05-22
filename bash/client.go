package bash

import (
	"fmt"
	"strings"

	"github.com/bjatkin/yok/ast"
	"github.com/bjatkin/yok/sym"
)

type Client struct {
	table    *sym.Table
	builders []builder
}

func NewClient(table *sym.Table) *Client {
	return &Client{
		table: table,
		builders: []builder{
			buildNewLine,
			buildAssign,
			buildComment,
			buildUseImport,
			buildEnv,
			buildRoot,
			buildCommandCall,
			buildIf,
		},
	}
}

func (c *Client) Build(tree ast.Stmt) Root {
	var stmts []Stmt

	if root, ok := tree.(ast.Root); ok {
		for _, stmt := range root.Stmts {
			root := c.Build(stmt)
			stmts = append(stmts, root.Stmts...)
		}
		return Root{Stmts: stmts}
	}

	var ret Root
	for _, builder := range c.builders {
		stmts := builder(c.table, stmts, tree)
		ret.Stmts = append(ret.Stmts, stmts...)
		if len(stmts) > 0 {
			break
		}
	}

	return ret
}

func (c *Client) Bash(tree Root) []byte {
	raw := []string{"#!/bin/bash", ""}
	for _, stmt := range tree.Stmts {
		raw = append(raw, stmt.Bash().String())
	}

	return []byte(strings.Join(raw, "\n") + "\n")
}

type builder func(*sym.Table, []Stmt, ast.Stmt) []Stmt

type Node interface {
	Bash() fmt.Stringer
}

type Stmt interface {
	Node
	stmt()
}

type Expr interface {
	Node
	expr()
}
