package bash

import (
	"fmt"
	"strings"

	"github.com/bjatkin/yok/ast"
	"github.com/bjatkin/yok/sym"
)

type Root struct {
	Stmt
	Stmts []Stmt
}

func (r Root) Bash() []string {
	var lines []string
	for _, stmt := range r.Stmts {
		lines = append(lines, stmt.Bash()...)
	}

	return lines
}

func buildRoot(table *sym.Table, stmts []Stmt, stmt ast.Stmt) []Stmt {
	if _, ok := stmt.(ast.Root); !ok {
		return nil
	}

	return []Stmt{Root{
		Stmts: stmts,
	}}
}

type NewLine struct {
	Stmt
	ID sym.ID
}

func (n NewLine) Bash() []string {
	return []string{""}
}

func buildNewLine(table *sym.Table, stmts []Stmt, stmt ast.Stmt) []Stmt {
	newLine, ok := stmt.(ast.NewLine)
	if !ok {
		return nil
	}

	return []Stmt{NewLine{ID: newLine.ID}}
}

type Use struct {
	Stmt
	ID      sym.ID
	Imports []If
}

func (u Use) Bash() []string {
	var lines []string
	for i, imp := range u.Imports {
		lines = append(lines, imp.Bash()...)
		if i < len(u.Imports)-1 {
			lines = append(lines, "") // add a newline between each import
		}
	}

	return lines
}

func buildUseImport(table *sym.Table, stmts []Stmt, stmt ast.Stmt) []Stmt {
	use, ok := stmt.(ast.Use)
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
					Root: Root{
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
					Root: Root{
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

	return []Stmt{ret}
}

type Comment struct {
	Stmt
	ID  sym.ID
	Raw string
}

func (c Comment) Bash() []string {
	return []string{"# " + c.Raw}
}

func buildComment(table *sym.Table, stmts []Stmt, stmt ast.Stmt) []Stmt {
	comment, ok := stmt.(ast.Comment)
	if !ok {
		return nil
	}

	return []Stmt{Comment{
		ID:  comment.ID,
		Raw: comment.Raw,
	}}
}

type Assign struct {
	Stmt
	ID         sym.ID
	Identifyer Identifyer
	SetTo      Expr
}

func (a Assign) Bash() []string {
	return []string{a.Identifyer.Name + "=" + strings.Join(a.SetTo.Bash(), "")}
}

func buildAssign(table *sym.Table, stmts []Stmt, stmt ast.Stmt) []Stmt {
	assign, ok := stmt.(ast.Assign)
	if !ok {
		return nil
	}

	switch v := assign.SetTo.(type) {
	case ast.Identifyer:
		return []Stmt{Assign{
			ID: assign.ID,
			Identifyer: Identifyer{
				ID:   assign.ID,
				Name: assign.Identifyer,
			},
			SetTo: Identifyer{
				ID:   v.ID,
				Name: v.Name,
			},
		}}
	case ast.Value:
		return []Stmt{Assign{
			ID: assign.ID,
			Identifyer: Identifyer{
				ID:   assign.ID,
				Name: assign.Identifyer,
			},
			SetTo: Value{
				ID:  v.ID,
				Raw: v.Raw,
			},
		}}
	default:
		// TODO: this should probably be an error
		return nil
	}
}

type If struct {
	Stmt
	ID    sym.ID
	Check Expr
	Root  Root
}

func (i If) Bash() []string {
	lines := []string{fmt.Sprintf("if %s; then", strings.Join(i.Check.Bash(), ""))}
	for _, line := range i.Root.Bash() {
		if line == "" {
			lines = append(lines, line)
			continue
		}
		lines = append(lines, indent+line)
	}
	lines = append(lines, "fi")

	return lines
}

func buildIf(table *sym.Table, stmts []Stmt, stmt ast.Stmt) []Stmt {
	ifBlock, ok := stmt.(ast.If)
	if !ok {
		return nil
	}

	ret := If{
		ID: ifBlock.ID,
	}

	// TODO: make a top level set of builders that just match expressions
	// then use it here ot match abitrary expression
	switch v := ifBlock.Check.(type) {
	case ast.Identifyer:
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
	case ast.Value:
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

	return []Stmt{ret}
}
