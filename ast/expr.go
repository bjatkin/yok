package ast

import (
	"fmt"
	"strings"

	"github.com/bjatkin/yok/parse"
	"github.com/bjatkin/yok/source"
	"github.com/bjatkin/yok/sym"
)

type Value struct {
	Expr
	ID  sym.ID
	Raw string
}

func (v Value) Yok() fmt.Stringer {
	return source.Line(v.Raw)
}

func buildValue(table *sym.Table, stmts []Stmt, node parse.Node) []Expr {
	if node.NodeType != parse.Value {
		return nil
	}

	return []Expr{Value{
		ID:  node.Token.ID,
		Raw: node.Token.Value,
	}}
}

type Identifyer struct {
	Expr
	ID   sym.ID
	Name string
}

func (i Identifyer) Yok() fmt.Stringer {
	return source.Line(i.Name)
}

func buildIdentifyer(table *sym.Table, stmts []Stmt, node parse.Node) []Expr {
	if node.NodeType != parse.Identifyer {
		return nil
	}

	return []Expr{Identifyer{
		ID:   node.Token.ID,
		Name: node.Token.Value,
	}}
}

type Command struct {
	Expr
	ID         sym.ID
	Identifyer string
	SubCommand []Identifyer
	Args       []Expr
}

func (c Command) Yok() fmt.Stringer {
	var subCommands []string
	for _, sub := range c.SubCommand {
		subCommands = append(subCommands, sub.Yok().String())
	}

	var args []string
	for _, arg := range c.Args {
		args = append(args, arg.Yok().String())
	}

	if len(subCommands) > 0 {
		return source.Line(fmt.Sprintf("%s.%s(%s)",
			c.Identifyer,
			strings.Join(subCommands, "."),
			strings.Join(args, ", "),
		))
	}

	return source.Line(fmt.Sprintf("%s(%s)",
		c.Identifyer,
		strings.Join(args, ", "),
	))
}

func buildCommandCall(table *sym.Table, stmts []Stmt, node parse.Node) []Stmt {
	if node.NodeType != parse.Call {
		return nil
	}
	if len(node.Nodes) < 1 {
		return nil
	}
	if node.Nodes[0].NodeType != parse.Identifyer {
		return nil
	}

	ret := Command{
		ID:         node.Nodes[0].Token.ID,
		Identifyer: table.MustGetSymbol(node.Nodes[0].Token.ID).Value,
	}

	client := NewClient(table)
	for _, n := range node.Nodes {
		if n.NodeType != parse.Arg {
			continue
		}

		if len(n.Nodes) > 1 && n.Nodes[0].NodeType == parse.Dot && n.Nodes[1].NodeType == parse.Identifyer {
			ret.SubCommand = append(ret.SubCommand,
				Identifyer{
					ID:   n.Nodes[1].Token.ID,
					Name: table.MustGetSymbol(n.Nodes[1].Token.ID).Value,
				},
			)
			continue
		}

		if len(n.Nodes) > 0 {
			ret.Args = append(ret.Args, client.buildExpr(stmts, n.Nodes[0])...)
			continue
		}

		panic(fmt.Sprintf("unknown arugment: %v", n.Nodes))
	}

	return []Stmt{ret}
}

type Env struct {
	Expr
	ID   sym.ID
	Name string
}

func (e Env) Yok() fmt.Stringer {
	return source.Line(fmt.Sprintf("env[%s]", e.Name))
}

func buildEnv(table *sym.Table, stmts []Stmt, node parse.Node) []Stmt {
	if node.NodeType != parse.EnvKeyword {
		return nil
	}
	if len(node.Nodes) < 2 {
		return nil
	}
	if node.Nodes[1].NodeType != parse.Value {
		return nil
	}

	return []Stmt{Env{
		ID:   node.Nodes[1].Token.ID,
		Name: table.MustGetSymbol(node.Nodes[1].Token.ID).Value,
	}}
}

type BinaryExpr struct {
	Expr
	Left  Expr
	Op    string
	Right Expr
}

func (b BinaryExpr) Yok() fmt.Stringer {
	return source.Linef("%s %s %s", b.Left.Yok(), b.Op, b.Right.Yok())
}

func buildBinaryExpr(table *sym.Table, stmts []Stmt, node parse.Node) []Expr {
	if node.NodeType != parse.Expr {
		return nil
	}
	if len(node.Nodes) < 3 {
		return nil
	}
	if node.Nodes[1].NodeType != parse.BinaryOp {
		return nil
	}

	left := node.Nodes[0]
	right := node.Nodes[2]
	ret := BinaryExpr{
		Op: node.Nodes[1].Token.Value,
	}
	if left.NodeType == parse.Identifyer {
		ret.Left = Identifyer{
			ID:   left.Token.ID,
			Name: left.Token.Value,
		}
	}
	if left.NodeType == parse.Value {
		ret.Left = Value{
			ID:  left.Token.ID,
			Raw: left.Token.Value,
		}
	}
	if right.NodeType == parse.Identifyer {
		ret.Right = Identifyer{
			ID:   right.Token.ID,
			Name: right.Token.Value,
		}
	}
	if right.NodeType == parse.Value {
		ret.Right = Value{
			ID:  right.Token.ID,
			Raw: right.Token.Value,
		}
	}
	if right.NodeType == parse.Expr {
		ret.Right = buildBinaryExpr(table, nil, right)[0]
	}

	return []Expr{ret}
}
