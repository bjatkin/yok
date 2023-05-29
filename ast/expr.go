package ast

import (
	"fmt"
	"strings"

	"github.com/bjatkin/yok/parse"
	"github.com/bjatkin/yok/source"
	"github.com/bjatkin/yok/sym"
)

type Expr interface {
	Node
	expr()
	stmt() // any expression can behave as a statment if the return value is ignored
	YokType() sym.YokType
}

type Value struct {
	Expr
	ID   sym.ID
	Raw  string
	Type sym.YokType
}

func (v *Value) Yok() fmt.Stringer {
	return source.Line(v.Raw)
}

func (v *Value) YokType() sym.YokType {
	return v.Type
}

func buildValue(table *sym.Table, node parse.Node) Expr {
	if node.Type != parse.Value {
		return nil
	}

	return &Value{
		ID:  node.ID,
		Raw: node.Value,
	}
}

type Identifyer struct {
	Expr
	ID   sym.ID
	Name string
	Type sym.YokType
}

func (i *Identifyer) Yok() fmt.Stringer {
	return source.Line(i.Name)
}

func (i *Identifyer) YokType() sym.YokType {
	return i.Type
}

func buildIdentifyer(table *sym.Table, node parse.Node) Expr {
	if node.Type != parse.Identifyer {
		return nil
	}

	return &Identifyer{
		ID:   node.ID,
		Name: node.Value,
	}
}

type Command struct {
	Expr
	Type       sym.YokType
	ID         sym.ID
	Identifyer string
	SubCommand []Value
	Args       []Expr
}

func (c *Command) Yok() fmt.Stringer {
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

func (c *Command) YokType() sym.YokType {
	return sym.StringType
}

func buildCommandCall(table *sym.Table, node parse.Node) Expr {
	if node.Type != parse.Call {
		return nil
	}
	if len(node.Nodes) < 1 {
		return nil
	}
	if node.Nodes[0].Type != parse.Identifyer {
		return nil
	}

	ret := &Command{
		ID:         node.Nodes[0].ID,
		Identifyer: node.Nodes[0].Value,
	}

	client := NewClient(table)
	for _, n := range node.Nodes {
		if n.Type != parse.Arg {
			continue
		}

		if len(n.Nodes) > 1 && n.Nodes[0].Type == parse.Dot && n.Nodes[1].Type == parse.Identifyer {
			ret.SubCommand = append(ret.SubCommand,
				Value{
					ID:  n.Nodes[1].ID,
					Raw: n.Nodes[1].Value,
				},
			)
			continue
		}

		if len(n.Nodes) > 0 {
			ret.Args = append(ret.Args, client.buildExpr(n.Nodes[0]))
			continue
		}

		panic(fmt.Sprintf("unknown arugment: %v", n.Nodes))
	}

	return ret
}

type Env struct {
	Expr
	ID   sym.ID
	Name string
}

func (e *Env) Yok() fmt.Stringer {
	return source.Line(fmt.Sprintf("env[%s]", e.Name))
}

func (e *Env) YokType() sym.YokType {
	return sym.StringType
}

func buildEnv(table *sym.Table, node parse.Node) Stmt {
	if node.Type != parse.EnvKeyword {
		return nil
	}
	if len(node.Nodes) < 2 {
		return nil
	}
	if node.Nodes[1].Type != parse.Value {
		return nil
	}

	return &Env{
		ID:   node.Nodes[1].ID,
		Name: node.Nodes[1].Value,
	}
}

type BinaryExpr struct {
	Expr
	Left  Expr
	Op    string
	Right Expr
	Type  sym.YokType
}

func (b *BinaryExpr) Yok() fmt.Stringer {
	return source.Linef("%s %s %s", b.Left.Yok(), b.Op, b.Right.Yok())
}

func (b *BinaryExpr) YokType() sym.YokType {
	return b.Type
}

func buildBinaryExpr(table *sym.Table, node parse.Node) Expr {
	if node.Type != parse.Expr {
		return nil
	}
	if len(node.Nodes) < 3 {
		return nil
	}
	if node.Nodes[1].Type != parse.BinaryOp {
		return nil
	}

	client := NewClient(table)
	left := client.buildExpr(node.Nodes[0])
	if left == nil {
		return nil
	}

	right := client.buildExpr(node.Nodes[2])
	if right == nil {
		return nil
	}

	return &BinaryExpr{
		Left:  left,
		Op:    node.Nodes[1].Value,
		Right: right,
	}
}
