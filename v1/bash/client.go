package bash

import (
	"fmt"
	"strings"

	"github.com/bjatkin/yok/ast"
	"github.com/bjatkin/yok/sym"
)

type astBuilder struct {
	root   *Root
	client *Client
}

func (b *astBuilder) Visit(node ast.Node) ast.Visitor {
	switch node.(type) {
	case *ast.If:
		stmt := buildIf(b.client.table, node)
		b.root.Stmts = append(b.root.Stmts, stmt)
		return nil
	case *ast.Assign:
		stmt := buildAssign(b.client.table, node)
		b.root.Stmts = append(b.root.Stmts, stmt)
		return nil
	default:
		stmt := b.client.buildStmt(node)
		if stmt != nil {
			b.root.Stmts = append(b.root.Stmts, stmt)
			return b
		}

		expr := b.client.buildExpr(node)
		if expr != nil {
			b.root.Stmts = append(b.root.Stmts, expr)
			return b
		}

		return b
	}
}

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
			buildIf,
		},
		exprBuilders: []exprBuilder{
			buildValue,
			buildEnv,
			buildIdentifyer,
			buildBinaryExpr,
			buildCommandCall,
		},
	}
}

func (c *Client) buildStmt(node ast.Node) Stmt {
	for _, builder := range c.stmtBuilders {
		stmt := builder(c.table, node)
		if stmt == nil {
			continue
		}

		return stmt
	}

	return nil
}

func (c *Client) buildExpr(node ast.Node) Expr {
	for _, builder := range c.exprBuilders {
		expr := builder(c.table, node)
		if expr == nil {
			continue
		}

		return expr
	}

	return nil
}

func (c *Client) Build(tree *ast.Root) *Root {
	builder := astBuilder{
		root:   &Root{},
		client: c,
	}

	tree.Walk(&builder)

	return builder.root
}

func (c *Client) Bash(tree *Root) []byte {
	raw := []string{"#!/bin/bash", ""}
	for _, stmt := range tree.Stmts {
		raw = append(raw, stmt.Bash().String())
	}

	return []byte(strings.Join(raw, "\n") + "\n")
}

type stmtBuilder func(*sym.Table, ast.Node) Stmt

type exprBuilder func(*sym.Table, ast.Node) Expr

type Node interface {
	Bash() fmt.Stringer
}
