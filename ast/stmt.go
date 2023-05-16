package ast

import (
	"fmt"
	"strings"

	"github.com/bjatkin/yok/parse"
	"github.com/bjatkin/yok/sym"
)

type Root struct {
	Expr
	Stmt
	Stmts []Stmt
}

func (r Root) Walk(fn WalkFunc) error {
	for _, stmt := range r.Stmts {
		if walker, ok := stmt.(Walker); ok {
			err := walker.walk(fn)
			if err != nil {
				return err
			}
		}

		err := fn(stmt)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r Root) Yok() []string {
	var lines []string
	for _, stmt := range r.Stmts {
		lines = append(lines, stmt.Yok()...)
	}

	return lines
}

func buildRoot(table *sym.Table, stmts []Stmt, node parse.Node) []Stmt {
	if node.NodeType != parse.Root {
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

func (n NewLine) Yok() []string {
	return []string{""}
}

func buildNewLine(table *sym.Table, stmts []Stmt, node parse.Node) []Stmt {
	if node.NodeType == parse.NewLineGroup {
		return []Stmt{NewLine{ID: node.Nodes[0].ID}}
	}
	return nil
}

type Use struct {
	Stmt
	ID      sym.ID
	Imports []Import
}

func (u Use) Yok() []string {
	var maxLen int
	var imports int
	for _, imp := range u.Imports {
		if yok := imp.Yok(); len(yok) > 0 {
			if len(yok[0]) > maxLen {
				maxLen = len(yok[0])
			}

			imports++
		}
	}

	var lines []string
	for _, imp := range u.Imports {
		yok := imp.Yok()
		if len(yok) != 2 {
			continue
		}

		for len(yok[1]) > 0 && len(yok[0]) < maxLen {
			yok[0] += " "
		}

		if imports == 1 {
			lines = append(lines, yok[0]+yok[1])
			continue
		}

		lines = append(lines, indent+yok[0]+yok[1])
	}
	lines = append(lines, "}")

	if len(lines) == 2 {
		return []string{fmt.Sprintf("use { %s }", lines[0])}
	}

	lines = append([]string{"use {"}, lines...)
	return lines
}

func buildUseImport(table *sym.Table, stmts []Stmt, node parse.Node) []Stmt {
	if node.NodeType != parse.UseKeyword {
		return nil
	}
	if len(node.Nodes) == 0 {
		return nil
	}
	if node.Nodes[0].NodeType != parse.OpenBlock {
		return nil
	}

	ret := Use{ID: node.ID}
	for _, n := range node.Nodes {
		imp := subUseImport(table, stmts, n)
		if imp != nil {
			ret.Imports = append(ret.Imports, *imp)
		}
	}

	return []Stmt{ret}
}

type Import struct {
	Stmt
	ID      sym.ID
	CmdName string
	Path    string
	Alias   string
}

func (i Import) Yok() []string {
	switch {
	case i.CmdName != "" && i.Alias != "":
		return []string{i.CmdName, " as " + i.Alias}
	case i.Path != "" && i.Alias != "":
		return []string{i.Path, " as " + i.Alias}
	case i.CmdName != "":
		return []string{i.CmdName, ""}
	case i.Path != "":
		return []string{i.Path, ""}
	default:
		return []string{"", ""}
	}
}

func subUseImport(table *sym.Table, stmts []Stmt, node parse.Node) *Import {
	if node.NodeType != parse.ImportExpr {
		return nil
	}
	if len(node.Nodes) == 0 {
		return nil
	}

	ret := &Import{ID: node.Nodes[0].ID}
	n := node.Nodes[0]
	switch {
	case n.NodeType == parse.Identifyer && len(node.Nodes) > 2 && node.Nodes[1].NodeType == parse.AsKeyword:
		cmdName := table.MustGetSymbol(n.ID)
		ret.CmdName = cmdName.Value
		alias := table.MustGetSymbol(node.Nodes[2].ID).Value
		cmdName.Alias = alias
		ret.Alias = alias

	case n.NodeType == parse.Value && len(node.Nodes) > 2 && node.Nodes[1].NodeType == parse.AsKeyword:
		path := table.MustGetSymbol(n.ID)
		ret.Path = path.Value
		alias := table.MustGetSymbol(node.Nodes[2].ID).Value
		path.Alias = alias
		ret.Alias = alias

	case n.NodeType == parse.Identifyer:
		ret.CmdName = table.MustGetSymbol(n.ID).Value

	case n.NodeType == parse.Value:
		ret.Path = table.MustGetSymbol(n.ID).Value

	}

	return ret
}

type Assign struct {
	Stmt
	ID         sym.ID
	Identifyer string
	SetTo      Expr
}

func (a Assign) Yok() []string {
	return []string{a.Identifyer + " = " + strings.Join(a.SetTo.Yok(), "")}
}

func buildAssign(table *sym.Table, stmts []Stmt, node parse.Node) []Stmt {
	if node.NodeType != parse.Assign {
		return nil
	}
	if len(node.Nodes) < 3 {
		return nil
	}
	if node.Nodes[0].NodeType != parse.Identifyer {
		return nil
	}

	ret := Assign{
		ID:         node.Nodes[0].ID,
		Identifyer: table.MustGetSymbol(node.Nodes[0].ID).Value,
	}

	switch {
	case node.Nodes[2].NodeType == parse.Value:
		ret.SetTo = Value{
			ID:  node.Nodes[2].ID,
			Raw: table.MustGetSymbol(node.Nodes[2].ID).Value,
		}
	case node.Nodes[2].NodeType == parse.Identifyer:
		ret.SetTo = Identifyer{
			ID:   node.Nodes[2].ID,
			Name: table.MustGetSymbol(node.Nodes[2].ID).Value,
		}
	}

	return []Stmt{ret}
}

type Comment struct {
	Stmt
	ID  sym.ID
	Raw string
}

func (c Comment) Yok() []string {
	return []string{"# " + c.Raw}
}

func buildComment(table *sym.Table, stmts []Stmt, node parse.Node) []Stmt {
	if node.NodeType != parse.Comment {
		return nil
	}

	symbol := table.MustGetSymbol(node.ID)
	symbol.Value = strings.Trim(symbol.Value, "# \t\r\n")

	return []Stmt{Comment{
		ID:  node.ID,
		Raw: symbol.Value,
	}}
}

type If struct {
	Stmt
	ID    sym.ID
	Check Expr
	Root  Root
}

func (i If) Yok() []string {
	var body []string
	for _, line := range i.Root.Yok() {
		if line == "" {
			body = append(body, line)
			continue
		}
		body = append(body, indent+line)
	}
	body = append(body, "}")

	var lines []string
	check := i.Check.Yok()
	for i, line := range check {
		if i == 0 {
			line = "if " + line
		}
		if i == len(check)-1 {
			line += " {"
		}
		if i > 0 {
			line = indent + line
		}
		lines = append(lines, line)
	}

	lines = append(lines, body...)
	return lines
}

func buildIf(table *sym.Table, stmts []Stmt, node parse.Node) []Stmt {
	if node.NodeType != parse.IfKeyword {
		return nil
	}
	if len(node.Nodes) < 1 {
		return nil
	}
	if node.Nodes[0].NodeType != parse.Expr {
		return nil
	}

	ret := If{
		ID: node.ID,
	}

	// TODO: make a top level set of matchers that just match expressions
	// then use it here to match abitrary expressions
	expr := node.Nodes[0]
	switch {
	case len(expr.Nodes) > 0 && expr.Nodes[0].NodeType == parse.Identifyer:
		ret.Check = Identifyer{
			ID:   expr.Nodes[0].ID,
			Name: table.MustGetSymbol(expr.Nodes[0].ID).Value,
		}
	case len(expr.Nodes) > 0 && expr.Nodes[0].NodeType == parse.Value:
		ret.Check = Value{
			ID:  expr.Nodes[0].ID,
			Raw: table.MustGetSymbol(expr.Nodes[0].ID).Value,
		}
	}

	if len(stmts) == 0 {
		return []Stmt{ret}
	}
	root, ok := stmts[0].(Root)
	if !ok {
		return []Stmt{ret}
	}

	ret.Root = root

	return []Stmt{ret}
}
