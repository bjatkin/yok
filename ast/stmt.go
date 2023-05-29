package ast

import (
	"fmt"
	"strings"

	"github.com/bjatkin/yok/parse"
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

func (r *Root) Yok() fmt.Stringer {
	var ret source.Block
	for _, stmt := range r.Stmts {
		ret.Lines = append(ret.Lines, stmt.Yok())
	}

	return ret
}

func (r *Root) walk(v visitor) {
	v = v.visit(r)
	if v == nil {
		return
	}

	for _, stmt := range r.Stmts {
		w, ok := stmt.(walker)
		if ok {
			w.walk(v)
			continue
		}

		v.visit(stmt)
	}
}

func buildRoot(table *sym.Table, node parse.Node) Stmt {
	if node.Type != parse.Root {
		return nil
	}

	client := NewClient(table)

	ret := &Root{}
	for _, node := range node.Nodes {
		stmt := client.build(node)
		if stmt != nil {
			ret.Stmts = append(ret.Stmts, stmt)
		}
	}

	return ret
}

type NewLine struct {
	Stmt
	ID sym.ID
}

func (n *NewLine) Yok() fmt.Stringer {
	return source.NewLine{}
}

func buildNewLine(table *sym.Table, node parse.Node) Stmt {
	if node.Type == parse.NewLineGroup {
		return &NewLine{ID: node.Nodes[0].ID}
	}
	return nil
}

type Use struct {
	Stmt
	ID      sym.ID
	Imports []*Import
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

func (u *Use) walk(v visitor) {
	v = v.visit(u)
	if v == nil {
		return
	}

	for _, imp := range u.Imports {
		v.visit(imp)
	}
}

func buildUseImport(table *sym.Table, node parse.Node) Stmt {
	if node.Type != parse.UseKeyword {
		return nil
	}
	if len(node.Nodes) == 0 {
		return nil
	}
	if node.Nodes[0].Type != parse.OpenBlock {
		return nil
	}

	ret := &Use{ID: node.ID}
	for _, n := range node.Nodes {
		imp := subUseImport(table, n)
		if imp != nil {
			ret.Imports = append(ret.Imports, imp)
		}
	}

	return ret
}

type Import struct {
	Stmt
	ID      sym.ID
	CmdName string
	Path    string
	Alias   string
}

func (i *Import) Yok() fmt.Stringer {
	name := i.CmdName
	if name == "" {
		name = i.Path
	}
	return source.Import{
		Name:  name,
		Alias: i.Alias,
	}
}

func subUseImport(table *sym.Table, node parse.Node) *Import {
	if node.Type != parse.ImportExpr {
		return nil
	}
	if len(node.Nodes) == 0 {
		return nil
	}

	ret := &Import{ID: node.Nodes[0].ID}
	n := node.Nodes[0]
	switch {
	case n.Type == parse.Identifyer && len(node.Nodes) > 2 && node.Nodes[1].Type == parse.AsKeyword:
		cmdName := table.MustGetSymbol(n.ID)
		ret.CmdName = cmdName.Value
		alias := table.MustGetSymbol(node.Nodes[2].ID).Value
		cmdName.Alias = alias
		ret.Alias = alias

	case n.Type == parse.Value && len(node.Nodes) > 2 && node.Nodes[1].Type == parse.AsKeyword:
		path := table.MustGetSymbol(n.ID)
		ret.Path = path.Value
		alias := table.MustGetSymbol(node.Nodes[2].ID).Value
		path.Alias = alias
		ret.Alias = alias

	case n.Type == parse.Identifyer:
		ret.CmdName = table.MustGetSymbol(n.ID).Value

	case n.Type == parse.Value:
		ret.Path = table.MustGetSymbol(n.ID).Value

	}

	return ret
}

type Assign struct {
	Stmt
	ID         sym.ID
	Type       sym.YokType
	Identifyer string
	SetTo      Expr
	IsDecl     bool
}

func (a *Assign) Yok() fmt.Stringer {
	if a.IsDecl {
		value := a.SetTo.Yok().String()
		return source.Linef("let %s %s", a.Identifyer, sym.TypeFromValue(value))
	}
	return source.Linef("%s = %s", a.Identifyer, a.SetTo.Yok())
}

func (a *Assign) walk(v visitor) {
	v = v.visit(a)
	if v == nil {
		return
	}

	w, ok := a.SetTo.(walker)
	if ok {
		w.walk(v)
		return
	}

	v.visit(a.SetTo)
}

func buildAssign(table *sym.Table, node parse.Node) Stmt {
	if node.Type != parse.Assign {
		return nil
	}
	if len(node.Nodes) < 3 {
		return nil
	}
	if node.Nodes[0].Type != parse.Identifyer {
		return nil
	}

	ret := &Assign{
		ID:         node.Nodes[0].ID,
		Identifyer: node.Nodes[0].Value,
	}

	client := NewClient(table)
	right := client.buildExpr(node.Nodes[2])
	if right == nil {
		return nil
	}

	ret.SetTo = right

	return ret
}

func buildDecl(table *sym.Table, node parse.Node) Stmt {
	if node.Type != parse.Decl {
		return nil
	}
	if len(node.Nodes) < 3 {
		return nil
	}
	if node.Nodes[0].Type != parse.LetKeyword {
		return nil
	}
	if node.Nodes[1].Type != parse.Identifyer {
		return nil
	}
	if node.Nodes[2].Type != parse.TypeKeyword {
		return nil
	}

	yokType := sym.StrToType(node.Nodes[2].Value)

	return &Assign{
		ID:         node.Nodes[1].ID,
		Identifyer: node.Nodes[1].Value,
		SetTo:      &Value{Raw: sym.DefaultValue(yokType)},
		Type:       yokType,
		IsDecl:     true,
	}
}

type Comment struct {
	Stmt
	ID  sym.ID
	Raw string
}

func (c *Comment) Yok() fmt.Stringer {
	return source.Linef("# %s", c.Raw)
}

func buildComment(table *sym.Table, node parse.Node) Stmt {
	if node.Type != parse.Comment {
		return nil
	}

	symbol := table.MustGetSymbol(node.ID)
	symbol.Value = strings.Trim(symbol.Value, "# \t\n")

	return &Comment{
		ID:  node.ID,
		Raw: symbol.Value,
	}
}

type If struct {
	Stmt
	ID    sym.ID
	Check Expr
	Root  *Root
}

func (i *If) Yok() fmt.Stringer {
	block, ok := i.Root.Yok().(source.Block)
	if !ok {
		return source.Linef("if %s { }", i.Check.Yok())
	}

	// TODO: Seems like the indentation on this block is incorrect...
	// either that or the indnetation on the use block is wrong
	return source.PrefixBlock{
		Prefix: source.Line(fmt.Sprintf("if %s {", i.Check.Yok())),
		Block:  block,
		Suffix: source.Line("}"),
	}
}

func (i *If) walk(v visitor) {
	v = v.visit(i)
	if v == nil {
		return
	}

	w, ok := i.Check.(walker)
	if ok {
		w.walk(v)
	} else {
		v.visit(i.Check)
	}

	for _, stmt := range i.Root.Stmts {
		w, ok := stmt.(walker)
		if ok {
			w.walk(v)
			continue
		}

		v.visit(stmt)
	}
}

func buildIf(table *sym.Table, node parse.Node) Stmt {
	if node.Type != parse.IfKeyword {
		return nil
	}
	if len(node.Nodes) < 2 {
		return nil
	}
	if node.Nodes[1].Type != parse.OpenBlock {
		return nil
	}

	client := NewClient(table)
	check := client.buildExpr(node.Nodes[0])
	if check == nil {
		return nil
	}

	root, ok := client.build(node.Nodes[3]).(*Root)
	if !ok {
		return nil
	}

	return &If{
		ID:    node.ID,
		Check: check,
		Root:  root,
	}
}
