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
}

type Value struct {
	Expr
	ID  sym.ID
	Raw string
}

func (v Value) Yok() fmt.Stringer {
	return source.Line(v.Raw)
}

func buildValue(table *sym.Table, node parse.Node) Expr {
	if node.Type != parse.Value {
		return nil
	}

	return Value{
		ID:  node.ID,
		Raw: node.Value,
	}
}

type Identifyer struct {
	Expr
	ID   sym.ID
	Name string
}

func (i Identifyer) Yok() fmt.Stringer {
	return source.Line(i.Name)
}

func buildIdentifyer(table *sym.Table, node parse.Node) Expr {
	if node.Type != parse.Identifyer {
		return nil
	}

	return Identifyer{
		ID:   node.ID,
		Name: node.Value,
	}
}

type Command struct {
	Expr
	ID         sym.ID
	Identifyer string
	// TODO: this should probably be an array of string literals instead
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

	ret := Command{
		ID:         node.Nodes[0].ID,
		Identifyer: table.MustGetSymbol(node.Nodes[0].ID).Value,
	}

	client := NewClient(table)
	for _, n := range node.Nodes {
		if n.Type != parse.Arg {
			continue
		}

		if len(n.Nodes) > 1 && n.Nodes[0].Type == parse.Dot && n.Nodes[1].Type == parse.Identifyer {
			ret.SubCommand = append(ret.SubCommand,
				Identifyer{
					ID:   n.Nodes[1].ID,
					Name: table.MustGetSymbol(n.Nodes[1].ID).Value,
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

func (e Env) Yok() fmt.Stringer {
	return source.Line(fmt.Sprintf("env[%s]", e.Name))
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

	return Env{
		ID:   node.Nodes[1].ID,
		Name: table.MustGetSymbol(node.Nodes[1].ID).Value,
	}
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

// TODO: use the client expression matcher to make this more robusts
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

	left := node.Nodes[0]
	right := node.Nodes[2]
	ret := BinaryExpr{
		Op: node.Nodes[1].Value,
	}
	if left.Type == parse.Identifyer {
		ret.Left = Identifyer{
			ID:   left.ID,
			Name: left.Value,
		}
	}
	if left.Type == parse.Value {
		ret.Left = Value{
			ID:  left.ID,
			Raw: left.Value,
		}
	}
	if right.Type == parse.Identifyer {
		ret.Right = Identifyer{
			ID:   right.ID,
			Name: right.Value,
		}
	}
	if right.Type == parse.Value {
		ret.Right = Value{
			ID:  right.ID,
			Raw: right.Value,
		}
	}
	if right.Type == parse.Expr {
		ret.Right = buildBinaryExpr(table, right)
	}

	return ret
}
