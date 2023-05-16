package bash

import (
	"fmt"
	"strings"

	"github.com/bjatkin/yok/ast"
	"github.com/bjatkin/yok/sym"
)

type Identifyer struct {
	Stmt
	Expr
	ID   sym.ID
	Name string
}

func (v Identifyer) Bash() []string {
	return []string{"$" + v.Name}
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

func (v Value) Bash() []string {
	return []string{v.Raw}
}

type Test struct {
	Expr
	Exprs []Expr
}

func (t Test) Bash() []string {
	var exprs []string
	for _, expr := range t.Exprs {
		exprs = append(exprs, expr.Bash()...)
	}

	return []string{fmt.Sprintf("[[ %s ]]", strings.Join(exprs, "; "))}
}

type Math struct {
	Expr
	Exprs []Expr
}

func (m Math) Bash() []string {
	var exprs []string
	for _, expr := range m.Exprs {
		exprs = append(exprs, expr.Bash()...)
	}

	return []string{fmt.Sprintf("(( %s ))", strings.Join(exprs, "; "))}
}

type BinaryExpr struct {
	Expr
	Left Expr
	// TODO: make this more strict than just an abitrary string
	Op    string
	Right Expr
}

func (b BinaryExpr) Bash() []string {
	return []string{fmt.Sprintf("%s %s %s", strings.Join(b.Left.Bash(), ""), b.Op, strings.Join(b.Right.Bash(), ""))}
}

type Command struct {
	Stmt
	Expr
	ID         sym.ID
	Identifyer string
	Args       []Expr
}

func (c Command) Bash() []string {
	var args []string
	for _, arg := range c.Args {
		args = append(args, arg.Bash()...)
	}

	return []string{fmt.Sprintf("%s %s",
		c.Identifyer,
		strings.Join(args, " "),
	)}
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

func (b FileExpr) Bash() []string {
	// TODO: this pattern of strings.Join(somthing.Bash, "") pops up alot to safely select the first
	// element of the bash list. I should probably not do it this way.
	return []string{b.Flag + " " + strings.Join(b.Check.Bash(), "")}
}

type SubShell struct {
	Expr
	Stmt
	Root Root
}

func (b SubShell) Bash() []string {
	return []string{fmt.Sprintf("\"$(%s)\"", strings.Join(b.Root.Bash(), "; "))}
}
