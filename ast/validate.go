package ast

import (
	"fmt"

	"github.com/bjatkin/yok/sym"
)

type validator interface {
	visitor
	errors() []string
}

type validateUse struct {
	useBlocks        int
	visited          int
	useIsNotFirst    bool
	duplicateImports []string
	unknownCommand   []string
	imported         map[string]int
}

func newValidateuse() *validateUse {
	return &validateUse{
		imported: map[string]int{
			"echo": 1, // echo is a built in
		},
	}
}

func (v *validateUse) visit(node Node) visitor {
	switch t := node.(type) {
	case *Comment:
		return nil
	case *NewLine:
		return nil
	case *Root:
		return v
	case *Use:
		if v.visited > 0 {
			v.useIsNotFirst = true
		}
		v.useBlocks++
		return v
	case *Import:
		name := t.Alias
		if name == "" {
			name = t.CmdName
		}
		if name == "" {
			name = t.Path
		}
		if _, ok := v.imported[name]; ok {
			v.duplicateImports = append(v.duplicateImports, name)
		}
		v.imported[name] = 0
		return nil
	case *Command:
		v.visited++
		if _, ok := v.imported[t.Identifyer]; !ok {
			v.unknownCommand = append(v.unknownCommand, t.Identifyer)
			return v
		}
		v.imported[t.Identifyer]++
		return v
	default:
		v.visited++
		return v
	}
}

func (v *validateUse) errors() []string {
	var errs []string
	if v.useIsNotFirst {
		errs = append(errs, "use import must be the first command in a yok script")
	}

	for _, imports := range v.duplicateImports {
		errs = append(errs, fmt.Sprintf("duplicate import %s", imports))
	}

	for _, cmd := range v.unknownCommand {
		errs = append(errs, fmt.Sprintf("unknown command %s", cmd))
	}

	for name, count := range v.imported {
		if count == 0 {
			errs = append(errs, fmt.Sprintf("unused command %s", name))
		}
	}

	return errs
}

// TODO: identify unused identifyers, this is a little tricky right now since we copy
// identifyers from outer scopes into inner scopes
type validateIdentifyers struct {
	unknownIdentifyer []string
	identifyers       map[string]int
	sub               []*validateIdentifyers
}

func newValidateIdentifyers() *validateIdentifyers {
	return &validateIdentifyers{
		identifyers: make(map[string]int),
	}
}

func (v *validateIdentifyers) newSub() *validateIdentifyers {
	ret := newValidateIdentifyers()
	for k, v := range v.identifyers {
		ret.identifyers[k] = v
	}
	v.sub = append(v.sub, ret)

	return ret
}

func (v *validateIdentifyers) visit(node Node) visitor {
	switch t := node.(type) {
	case *Root:
		return v.newSub()
	case *Identifyer:
		if _, ok := v.identifyers[t.Name]; !ok {
			v.unknownIdentifyer = append(v.unknownIdentifyer, t.Name)
			return nil
		}

		v.identifyers[t.Name]++
		return nil
	case *Assign:
		if w, ok := t.SetTo.(walker); ok {
			w.walk(v)
		} else {
			v.visit(t.SetTo)
		}

		v.identifyers[t.Identifyer] = 0
		return nil
	case *Command:
		for _, arg := range t.Args {
			if w, ok := arg.(walker); ok {
				w.walk(v)
			} else {
				v.visit(arg)
			}
		}
		return nil
	default:
		return v
	}
}

func (v *validateIdentifyers) errors() []string {
	var errs []string
	for _, unknown := range v.unknownIdentifyer {
		errs = append(errs, fmt.Sprintf("unknown identifyer: %s", unknown))
	}

	for _, sub := range v.sub {
		errs = append(errs, sub.errors()...)
	}

	return errs
}

type validateTypes struct {
	typeMissmatch []string
	names         map[string]sym.YokType
	sub           []*validateTypes
}

func newValidateTypes() *validateTypes {
	return &validateTypes{
		names: make(map[string]sym.YokType),
	}
}

func (v *validateTypes) newSub() *validateTypes {
	ret := newValidateTypes()
	for k, v := range v.names {
		ret.names[k] = v
	}
	v.sub = append(v.sub, ret)

	return ret
}

func (v *validateTypes) FindType(name string) sym.YokType {
	for have, t := range v.names {
		if have == name {
			return t
		}
	}

	return sym.UnknownType
}

// TODO: set types for values as well so I don't need to remember to include type information
// when building the AST. It would make sense if every node could be given a type at creation time
// however, since that's not possible, I should do it all here at once.
func (v *validateTypes) visit(node Node) visitor {
	switch t := node.(type) {
	case *Root:
		return v.newSub()
	case *Value:
		t.Type = sym.TypeFromValue(t.Raw)
	case *Assign:
		if w, ok := t.SetTo.(walker); ok {
			w.walk(v)
		} else {
			v.visit(t.SetTo)
		}

		if v.names[t.Identifyer] != sym.UnknownType && v.names[t.Identifyer] != t.SetTo.yokType() {
			v.typeMissmatch = append(
				v.typeMissmatch,
				fmt.Sprintf("missmatched types %s and %s for assignemnt %s", v.names[t.Identifyer], t.SetTo.yokType(), t.Yok()),
			)
			return nil
		}

		v.names[t.Identifyer] = t.SetTo.yokType()
		t.Type = t.SetTo.yokType()

		return nil
	case *Identifyer:
		if t.Type != sym.UnknownType && t.Type != v.names[t.Name] {
			v.typeMissmatch = append(
				v.typeMissmatch,
				fmt.Sprintf("missmatched types %s and %s for identifyer %s", t.Type, v.names[t.Name], t.Name),
			)
			return nil
		}
		t.Type = v.names[t.Name]
	case *BinaryExpr:
		if w, ok := t.Left.(walker); ok {
			w.walk(v)
		} else {
			v.visit(t.Left)
		}

		if w, ok := t.Right.(walker); ok {
			w.walk(v)
		} else {
			v.visit(t.Right)
		}

		if t.Left.yokType() != t.Right.yokType() {
			v.typeMissmatch = append(
				v.typeMissmatch,
				fmt.Sprintf("missmatched types %s and %s for binary expression %s", t.Left.yokType(), t.Right.yokType(), t.Yok()),
			)
		}

		t.Type = t.Left.yokType()
		return nil
	default:
		return v
	}
	return nil
}

func (v *validateTypes) errors() []string {
	var errs []string
	for _, miss := range v.typeMissmatch {
		errs = append(errs, fmt.Sprintf("type missmatch: %s", miss))
	}

	for _, sub := range v.sub {
		errs = append(errs, sub.errors()...)
	}

	return errs
}
