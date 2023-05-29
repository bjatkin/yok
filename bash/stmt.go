package bash

import (
	"fmt"

	"github.com/bjatkin/yok/ast"
	"github.com/bjatkin/yok/source"
	"github.com/bjatkin/yok/sym"
)

type Stmt interface {
	Node
	stmt()
}

type Root struct {
	Stmt
	Stmts []Stmt
}

func (r Root) Bash() fmt.Stringer {
	var ret source.Block
	for _, stmt := range r.Stmts {
		ret.Lines = append(ret.Lines, stmt.Bash())
	}

	return ret
}

type NewLine struct {
	Stmt
	ID sym.ID
}

func (n NewLine) Bash() fmt.Stringer {
	return source.NewLine{}
}

func buildNewLine(table *sym.Table, node ast.Node) Stmt {
	newLine, ok := node.(*ast.NewLine)
	if !ok {
		return nil
	}

	return NewLine{ID: newLine.ID}
}

type Use struct {
	Stmt
	ID      sym.ID
	Imports []If
}

func (u Use) Bash() fmt.Stringer {
	var ret source.Block
	for i, imp := range u.Imports {
		ret.Lines = append(ret.Lines, imp.Bash())
		if i < len(u.Imports)-1 {
			ret.Lines = append(ret.Lines, source.NewLine{}) // add a newline between each import
		}
	}

	return ret
}

func buildUseImport(table *sym.Table, node ast.Node) Stmt {
	use, ok := node.(*ast.Use)
	if !ok {
		return nil
	}

	ret := Use{
		ID: use.ID,
	}

	for _, imp := range use.Imports {
		if imp.CmdName != "" {
			ret.Imports = append(ret.Imports,
				If{
					ID: imp.ID,
					Check: Test{
						Exprs: []Expr{
							FileExpr{
								Flag: "-z",
								Check: SubShell{
									Root: Root{
										Stmts: []Stmt{
											Command{
												Identifyer: "command",
												Args:       []Expr{Value{Raw: "-v"}, Value{Raw: imp.CmdName}},
											},
										},
									},
								},
							},
						},
					},
					Root: &Root{
						Stmts: []Stmt{
							// TODO: add a recirect to send this message to std error
							// echo "..." >&2
							Command{
								Identifyer: "echo",
								Args:       []Expr{Value{Raw: `"this script uses the command line tool ` + imp.CmdName + `, but it could not be found in your path"`}},
							},
							Command{
								Identifyer: "exit",
								Args:       []Expr{Value{Raw: "255"}},
							},
						},
					},
				},
			)
		}

		if imp.Path != "" {
			ret.Imports = append(ret.Imports,
				If{
					ID: imp.ID,
					Check: Test{
						Exprs: []Expr{
							FileExpr{
								Flag: "! -x",
								Check: SubShell{
									Root: Root{
										Stmts: []Stmt{Value{Raw: imp.Path}},
									},
								},
							},
						},
					},
					Root: &Root{
						Stmts: []Stmt{
							// TODO: add a recirect to send this message to std error
							// echo "..." >&2
							Command{
								Identifyer: "echo",
								Args:       []Expr{Value{Raw: `"this script uses ` + imp.Path + `, but it could not be found"`}},
							},
							Command{
								Identifyer: "exit",
								Args:       []Expr{Value{Raw: "254"}},
							},
						},
					},
				},
			)
		}
	}

	return ret
}

type Comment struct {
	Stmt
	ID  sym.ID
	Raw string
}

func (c Comment) Bash() fmt.Stringer {
	return source.Linef("# %s", c.Raw)
}

func buildComment(table *sym.Table, node ast.Node) Stmt {
	comment, ok := node.(*ast.Comment)
	if !ok {
		return nil
	}

	return Comment{
		ID:  comment.ID,
		Raw: comment.Raw,
	}
}

type Assign struct {
	Stmt
	ID         sym.ID
	Identifyer Identifyer
	SetTo      Expr
}

func (a Assign) Bash() fmt.Stringer {
	return source.Linef("%s=%s", a.Identifyer.Name, a.SetTo.Bash())
}

func buildAssign(table *sym.Table, node ast.Node) Stmt {
	assign, ok := node.(*ast.Assign)
	if !ok {
		return nil
	}

	// TODO: swap this out with client and expressions matcher
	switch v := assign.SetTo.(type) {
	case *ast.Identifyer:
		return Assign{
			ID: assign.ID,
			Identifyer: Identifyer{
				ID:   assign.ID,
				Name: assign.Identifyer,
			},
			SetTo: Identifyer{
				ID:   v.ID,
				Name: v.Name,
			},
		}
	case *ast.Value:
		return Assign{
			ID: assign.ID,
			Identifyer: Identifyer{
				ID:   assign.ID,
				Name: assign.Identifyer,
			},
			SetTo: Value{
				ID:  v.ID,
				Raw: v.Raw,
			},
		}
	case *ast.BinaryExpr:
		stmt := buildBinaryExpr(table, v)
		expr, ok := stmt.(Expr)
		if !ok {
			panic("build binary expr returned a stmt not an expr")
		}
		return Assign{
			ID: assign.ID,
			Identifyer: Identifyer{
				ID:   assign.ID,
				Name: assign.Identifyer,
			},
			SetTo: Math{Exprs: []Expr{expr}},
		}
	default:
		panic(fmt.Sprintf("unknonwn set to type %T", v))
	}
}

type If struct {
	Stmt
	ID    sym.ID
	Check Expr
	Root  *Root
}

func (i If) Bash() fmt.Stringer {
	ret := source.PrefixBlock{
		Prefix: source.Linef("if %s; then", i.Check.Bash()),
		Suffix: source.Line("fi"),
	}

	for _, stmt := range i.Root.Stmts {
		ret.Block.Lines = append(ret.Block.Lines, stmt.Bash())
	}

	return ret
}

func buildIf(table *sym.Table, node ast.Node) Stmt {
	ifBlock, ok := node.(*ast.If)
	if !ok {
		return nil
	}

	ret := If{
		ID: ifBlock.ID,
	}

	// TODO: make a top level set of builders that just match expressions
	// then use it here ot match abitrary expression
	switch v := ifBlock.Check.(type) {
	case *ast.Identifyer:
		ret.Check = Test{
			Exprs: []Expr{BinaryExpr{
				Left: Identifyer{
					ID:   v.ID,
					Name: v.Name,
				},
				Op: "==",
				Right: Value{
					Raw: "true",
				},
			}},
		}
	case *ast.Value:
		ret.Check = Test{
			Exprs: []Expr{BinaryExpr{
				Left: Value{
					ID:  v.ID,
					Raw: v.Raw,
				},
				Op: "==",
				Right: Value{
					Raw: "true",
				},
			}},
		}
	default:
		// TODO: this should probably be an error
		return nil
	}

	client := NewClient(table)
	ret.Root = client.Build(ifBlock.Root)

	return ret
}
