package bash

import (
	"fmt"
	"strings"

	"github.com/bjatkin/yok/ast"
	"github.com/bjatkin/yok/source"
	"github.com/bjatkin/yok/sym"
)

type Identifyer struct {
	Stmt
	Expr
	ID   sym.ID
	Name string
}

func (v Identifyer) Bash() fmt.Stringer {
	return source.Linef("$%s", v.Name)
}

func buildEnv(table *sym.Table, stmts []Stmt, stmt ast.Stmt) []Stmt {
	env, ok := stmt.(ast.Env)
	if !ok {
		return nil
	}

	return []Stmt{Identifyer{
		ID:   env.ID,
		Name: strings.Trim(env.Name, `"`),
	}}
}

type Value struct {
	Expr
	Stmt
	ID  sym.ID
	Raw string
}

func (v Value) Bash() fmt.Stringer {
	return source.Line(v.Raw)
}

type Test struct {
	Expr
	Exprs []Expr
}

func (t Test) Bash() fmt.Stringer {
	var exprs []string
	for _, expr := range t.Exprs {
		exprs = append(exprs, expr.Bash().String())
	}

	return source.Linef("[[ %s ]]", strings.Join(exprs, "; "))
}

type Math struct {
	Expr
	Exprs []Expr
}

func (m Math) Bash() fmt.Stringer {
	var exprs []string
	for _, expr := range m.Exprs {
		exprs = append(exprs, expr.Bash().String())
	}

	return source.Linef("$(( %s ))", strings.Join(exprs, "; "))
}

type BinaryExpr struct {
	Expr
	Stmt
	Left Expr
	// TODO: make this more strict than just an abitrary string
	Op    string
	Right Expr
}

func (b BinaryExpr) Bash() fmt.Stringer {
	return source.Linef("%s %s %s", b.Left.Bash().String(), b.Op, b.Right.Bash().String())
}

func buildBinaryExpr(table *sym.Table, stmts []Stmt, stmt ast.Stmt) []Stmt {
	expr, ok := stmt.(ast.BinaryExpr)
	if !ok {
		return nil
	}

	// TODO: would be nice to have a general build exprs thing here
	ret := BinaryExpr{
		Op: expr.Op,
	}
	switch v := expr.Left.(type) {
	case ast.Identifyer:
		ret.Left = Identifyer{ID: v.ID, Name: v.Name}
	case ast.Value:
		ret.Left = Value{ID: v.ID, Raw: v.Raw}
	default:
		panic("unknown left type in bash binary expression")
	}

	switch v := expr.Right.(type) {
	case ast.Identifyer:
		ret.Right = Identifyer{ID: v.ID, Name: v.Name}
	case ast.Value:
		ret.Right = Value{ID: v.ID, Raw: v.Raw}
	case ast.BinaryExpr:
		// TODO: this is really messy, clean this up
		stmt := buildBinaryExpr(table, nil, v)[0]
		if expr, ok := stmt.(Expr); ok {
			ret.Right = expr
		}
	default:
		panic("unknown left type in bash binary expression")
	}

	return []Stmt{ret}
}

type Command struct {
	Stmt
	Expr
	ID         sym.ID
	Identifyer string
	Args       []Expr
}

func (c Command) Bash() fmt.Stringer {
	var args []string
	for _, arg := range c.Args {
		args = append(args, arg.Bash().String())
	}

	return source.Linef("%s %s",
		c.Identifyer,
		strings.Join(args, " "),
	)
}

func buildCommandCall(table *sym.Table, stmts []Stmt, stmt ast.Stmt) []Stmt {
	call, ok := stmt.(ast.Command)
	if !ok {
		return nil
	}

	ret := Command{
		ID:         call.ID,
		Identifyer: call.Identifyer,
	}

	for _, arg := range call.SubCommand {
		ret.Args = append(ret.Args, Value{
			ID:  arg.ID,
			Raw: arg.Name,
		})
	}

	for _, arg := range call.Args {
		switch v := arg.(type) {
		case ast.Identifyer:
			ret.Args = append(ret.Args, Identifyer{
				ID:   v.ID,
				Name: v.Name,
			})
		case ast.Value:
			ret.Args = append(ret.Args, Value{
				ID:  v.ID,
				Raw: v.Raw,
			})
		}
	}

	return []Stmt{ret}
}

type FileExpr struct {
	Expr
	// TODO: make this more strict than just an abitrary string
	Flag  string
	Check Expr
}

func (b FileExpr) Bash() fmt.Stringer {
	return source.Linef("%s %s", b.Flag, b.Check.Bash())
}

type SubShell struct {
	Expr
	Stmt
	Root Root
}

func (b SubShell) Bash() fmt.Stringer {
	return source.Linef("\"$(%s)\"", b.Root.Bash())
}
