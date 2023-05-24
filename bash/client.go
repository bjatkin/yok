package bash

import (
	"fmt"
	"strings"

	"github.com/bjatkin/yok/ast"
	"github.com/bjatkin/yok/sym"
)

type Client struct {
	table        *sym.Table
	stmtBuilders []stmtBuilder
	exprBuilders []exprBuilder
}

func NewClient(table *sym.Table) *Client {
	return &Client{
		table: table,
		stmtBuilders: []stmtBuilder{
			buildNewLine,
			buildAssign,
			buildComment,
			buildUseImport,
			buildEnv,
			buildRoot,
			buildIf,
		},
		exprBuilders: []exprBuilder{
			buildBinaryExpr,
			buildCommandCall,
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
	for _, builder := range c.stmtBuilders {
		built := builder(c.table, stmts, tree)
		if len(built) == 0 {
			continue
		}

		ret.Stmts = append(ret.Stmts, built...)
		return ret
	}

	for _, builder := range c.exprBuilders {
		built := builder(c.table, stmts, tree)
		if len(built) == 0 {
			continue
		}

		for _, expr := range built {
			ret.Stmts = append(ret.Stmts, expr)
		}
		break
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

type stmtBuilder func(*sym.Table, []Stmt, ast.Node) []Stmt

type exprBuilder func(*sym.Table, []Stmt, ast.Node) []Expr

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
	stmt() // any expression can behave as a statment if the return value is ignored
}
