package ast

import (
	"fmt"
	"strings"

	"github.com/bjatkin/yok/parse"
	"github.com/bjatkin/yok/source"
	"github.com/bjatkin/yok/sym"
)

type Root struct {
	Expr
	Stmt
	Stmts []Stmt
}

func (r Root) Yok() fmt.Stringer {
	var ret source.Block
	for _, stmt := range r.Stmts {
		ret.Lines = append(ret.Lines, stmt.Yok())
	}

	return ret
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

func (n NewLine) Yok() fmt.Stringer {
	return source.NewLine{}
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

func (u Use) Yok() fmt.Stringer {
	var maxLen int
	var imports []source.Import
	for _, imp := range u.Imports {
		yok, ok := imp.Yok().(source.Import)
		if !ok {
			continue
		}

		if len(yok.Name) > maxLen {
			maxLen = len(yok.Name)
		}

		imports = append(imports, yok)
	}

	if len(imports) == 1 {
		return source.Linef("use { %s }", imports[0].String())
	}

	ret := source.PrefixBlock{
		Prefix: source.Line("use {"),
		Suffix: source.Line("}"),
	}

	for _, imp := range imports {
		imp.MaxNameLen = maxLen
		ret.Block.Lines = append(ret.Block.Lines, imp)
	}

	return ret
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

func (i Import) Yok() fmt.Stringer {
	name := i.CmdName
	if name == "" {
		name = i.Path
	}
	return source.Import{
		Name:  name,
		Alias: i.Alias,
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
	IsDecl     bool
}

func (a Assign) Yok() fmt.Stringer {
	if a.IsDecl {
		value := a.SetTo.Yok().String()
		return source.Linef("let %s %s", a.Identifyer, sym.TypeFromValue(value))
	}
	return source.Linef("%s = %s", a.Identifyer, a.SetTo.Yok())
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
		Identifyer: node.Nodes[0].Value,
	}

	switch {
	case node.Nodes[2].NodeType == parse.Value:
		ret.SetTo = Value{
			ID:  node.Nodes[2].ID,
			Raw: node.Nodes[2].Value,
		}

	case node.Nodes[2].NodeType == parse.Identifyer:
		ret.SetTo = Identifyer{
			ID:   node.Nodes[2].ID,
			Name: node.Nodes[2].Value,
		}
	}

	return []Stmt{ret}
}

func buildDecl(table *sym.Table, stmts []Stmt, node parse.Node) []Stmt {
	if node.NodeType != parse.Decl {
		return nil
	}
	if len(node.Nodes) < 3 {
		return nil
	}
	if node.Nodes[0].NodeType != parse.LetKeyword {
		return nil
	}
	if node.Nodes[1].NodeType != parse.Identifyer {
		return nil
	}
	if node.Nodes[2].NodeType != parse.TypeKeyword {
		return nil
	}

	yokType := sym.StrToType(node.Nodes[2].Value)

	return []Stmt{Assign{
		ID:         node.Nodes[1].ID,
		Identifyer: node.Nodes[1].Value,
		SetTo:      Value{Raw: sym.DefaultValue(yokType)},
		IsDecl:     true,
	}}
}

type Comment struct {
	Stmt
	ID  sym.ID
	Raw string
}

func (c Comment) Yok() fmt.Stringer {
	return source.Linef("# %s", c.Raw)
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

func (i If) Yok() fmt.Stringer {
	block, ok := i.Root.Yok().(source.Block)
	if !ok {
		return source.Linef("if %s { }", i.Check.Yok())
	}

	return source.PrefixBlock{
		Prefix: source.Line(fmt.Sprintf("if %s {", i.Check.Yok())),
		Block:  block,
		Suffix: source.Line("}"),
	}
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
