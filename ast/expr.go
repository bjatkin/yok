package ast

import (
	"fmt"
	"strings"

	"github.com/bjatkin/yok/parse"
	"github.com/bjatkin/yok/source"
	"github.com/bjatkin/yok/sym"
)

type Value struct {
	Stmt
	Expr
	ID  sym.ID
	Raw string
}

func (v Value) Yok() fmt.Stringer {
	return source.Line(v.Raw)
}

type Identifyer struct {
	Stmt
	Expr
	ID   sym.ID
	Name string
}

func (i Identifyer) Yok() fmt.Stringer {
	return source.Line(i.Name)
}

type Command struct {
	Stmt
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
		ID:         node.Nodes[0].ID,
		Identifyer: table.MustGetSymbol(node.Nodes[0].ID).Value,
	}
	for _, n := range node.Nodes {
		if n.NodeType != parse.Arg {
			continue
		}

		switch {
		case len(n.Nodes) > 1 && n.Nodes[0].NodeType == parse.Dot && n.Nodes[1].NodeType == parse.Identifyer:
			ret.SubCommand = append(ret.SubCommand,
				Identifyer{
					ID:   n.Nodes[1].ID,
					Name: table.MustGetSymbol(n.Nodes[1].ID).Value,
				},
			)
		case len(n.Nodes) > 0 && n.Nodes[0].NodeType == parse.Identifyer:
			ret.Args = append(ret.Args,
				Identifyer{
					ID:   n.Nodes[0].ID,
					Name: table.MustGetSymbol(n.Nodes[0].ID).Value,
				},
			)
		case len(n.Nodes) > 0 && n.Nodes[0].NodeType == parse.Value:
			ret.Args = append(ret.Args,
				Value{
					ID:  n.Nodes[0].ID,
					Raw: table.MustGetSymbol(n.Nodes[0].ID).Value,
				},
			)
		}
	}

	return []Stmt{ret}
}

type Env struct {
	Stmt
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
		ID:   node.Nodes[1].ID,
		Name: table.MustGetSymbol(node.Nodes[1].ID).Value,
	}}
}
