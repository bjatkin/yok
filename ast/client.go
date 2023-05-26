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
			buildNewLine,
			buildEnv,
			buildIf,
		},
		exprBuilders: []exprBuilder{
			buildCommandCall,
			buildBinaryExpr,
			buildIdentifyer,
			buildValue,
		},
		validators: []validator{
			newValidateuse(),
		},
	}
}

func (c *Client) Build(tree parse.Node) Root {
	if tree.Type != parse.Root {
		// TODO: this should be an error
		panic("tree root passed to build must have type root not " + tree.Type)
	}

	root, ok := c.build(tree).(Root)
	if !ok {
		// TODO: this should be an error
		panic("invalid AST root")
	}

	return root
}

func (c *Client) build(node parse.Node) Stmt {
	stmt := c.buildStmt(node)
	if stmt != nil {
		return stmt
	}

	expr := c.buildExpr(node)
	if expr != nil {
		return expr
	}

	return nil
}

func (c *Client) buildStmt(node parse.Node) Stmt {
	for _, builder := range c.stmtBuilders {
		stmt := builder(c.table, node)
		if stmt == nil {
			continue
		}

		return stmt
	}

	return nil
}

func (c *Client) buildExpr(node parse.Node) Expr {
	for _, builder := range c.exprBuilders {
		expr := builder(c.table, node)
		if expr == nil {
			continue
		}

		return expr
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

func (c *Client) Validate(tree Root) error {
	for _, validator := range c.validators {
		tree.walk(validator)
		errs := validator.errors()
		if len(errs) > 0 {
			return fmt.Errorf("validation failure: %s", strings.Join(errs, "\n\t"))
		}
	}

	return nil
}

type Node interface {
	Yok() fmt.Stringer
}

type stmtBuilder func(*sym.Table, parse.Node) Stmt

type exprBuilder func(*sym.Table, parse.Node) Expr

type visitor interface {
	visit(Node) visitor
}

type walker interface {
	walk(visitor)
}
