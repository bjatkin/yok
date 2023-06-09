package bash

import (
	"fmt"
	"strings"

	"github.com/bjatkin/yok/ast"
	"github.com/bjatkin/yok/source"
	"github.com/bjatkin/yok/sym"
)

type Expr interface {
	Node
	expr()
	stmt() // any expression can behave as a statment if the return value is ignored
}

type Identifyer struct {
	Expr
	ID   sym.ID
	Name string
}

func (v Identifyer) Bash() fmt.Stringer {
	return source.Linef("$%s", v.Name)
}

func buildIdentifyer(table *sym.Table, node ast.Node) Expr {
	identifyer, ok := node.(*ast.Identifyer)
	if !ok {
		return nil
	}

	return Identifyer{
		ID:   identifyer.ID,
		Name: identifyer.Name,
	}
}

func buildEnv(table *sym.Table, node ast.Node) Expr {
	env, ok := node.(*ast.Env)
	if !ok {
		return nil
	}

	return Identifyer{
		ID:   env.ID,
		Name: strings.Trim(env.Name, `"`),
	}
}

type Value struct {
	Expr
	ID  sym.ID
	Raw string
}

func (v Value) Bash() fmt.Stringer {
	return source.Line(v.Raw)
}

func buildValue(table *sym.Table, node ast.Node) Expr {
	value, ok := node.(*ast.Value)
	if !ok {
		return nil
	}

	return &Value{
		ID:  value.ID,
		Raw: value.Raw,
	}
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
	Left Expr
	// TODO: make this more strict than just an abitrary string
	Op    string
	Right Expr
}

func (b BinaryExpr) Bash() fmt.Stringer {
	if b.Op == "string concat" {
		return source.Line(b.Left.Bash().String() + b.Right.Bash().String())
	}
	return source.Linef("%s %s %s", b.Left.Bash().String(), b.Op, b.Right.Bash().String())
}

func buildBinaryExpr(table *sym.Table, node ast.Node) Expr {
	expr, ok := node.(*ast.BinaryExpr)
	if !ok {
		return nil
	}

	ret := BinaryExpr{
		Op: expr.Op,
	}
	if expr.Left.YokType() == sym.StringType && expr.Op == "+" {
		ret.Op = "string concat"
	}

	client := NewClient(table)
	left := client.buildExpr(expr.Left)
	if left == nil {
		panic("unknown left type in bash binary expression")
	}
	ret.Left = left

	right := client.buildExpr(expr.Right)
	if right == nil {
		panic("unknown right type in bash binary expression")
	}
	ret.Right = right

	// TODO: would be nice to have a general build exprs thing here
	// switch v := expr.Left.(type) {
	// case *ast.Identifyer:
	// 	ret.Left = Identifyer{ID: v.ID, Name: v.Name}
	// case *ast.Value:
	// 	ret.Left = Value{ID: v.ID, Raw: v.Raw}
	// default:
	// 	panic("unknown left type in bash binary expression")
	// }

	// switch v := expr.Right.(type) {
	// case *ast.Identifyer:
	// 	ret.Right = Identifyer{ID: v.ID, Name: v.Name}
	// case *ast.Value:
	// 	ret.Right = Value{ID: v.ID, Raw: v.Raw}
	// case *ast.BinaryExpr:
	// 	ret.Right = buildBinaryExpr(table, v)
	// default:
	// 	panic("unknown left type in bash binary expression")
	// }

	return ret
}

type Command struct {
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

func buildCommandCall(table *sym.Table, node ast.Node) Expr {
	call, ok := node.(*ast.Command)
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
			Raw: arg.Raw,
		})
	}

	for _, arg := range call.Args {
		switch v := arg.(type) {
		case *ast.Identifyer:
			ret.Args = append(ret.Args, Identifyer{
				ID:   v.ID,
				Name: v.Name,
			})
		case *ast.Value:
			ret.Args = append(ret.Args, Value{
				ID:  v.ID,
				Raw: v.Raw,
			})
		}
	}

	return ret
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
	Root Root
}

func (b SubShell) Bash() fmt.Stringer {
	return source.Linef("\"$(%s)\"", b.Root.Bash())
}
