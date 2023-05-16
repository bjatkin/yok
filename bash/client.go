package bash

import (
	"strings"

	"github.com/bjatkin/yok/ast"
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

	root, ok := tree.(ast.Root)
	if ok {
		for _, stmt := range root.Stmts {
			root := c.Build(stmt)
			stmts = append(stmts, root.Stmts...)
		}
	}

	ret := Root{}
	for _, builder := range c.builders {
		stmts := builder(c.table, stmts, tree)
		ret.Stmts = append(ret.Stmts, stmts...)
		if len(stmts) > 0 {
			break
		}
	}

	return ret
}

func (c *Client) Bash(tree Root) ([]byte, error) {
	var raw []string
	err := tree.Walk(func(s Stmt) error {
		raw = append(raw, s.Bash()...)
		return nil
	})
	if err != nil {
		return nil, err
	}

	raw = append([]string{"#!/bin/bash", ""}, raw...)
	return []byte(strings.Join(raw, "\n")), nil
}

// TODO: this really seems like more of a converter or a transposer
type builder func(*sym.Table, []Stmt, ast.Stmt) []Stmt

type WalkFunc func(Stmt) error

type Walker interface {
	walk(WalkFunc) error
}

type Stmt interface {
	stmt()
	Bash() []string
}

type Expr interface {
	expr()
	Bash() []string
}
