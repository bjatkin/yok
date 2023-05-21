package ast

import (
	"fmt"

	"github.com/bjatkin/yok/sym"
)

type validator interface {
	check(Stmt) error
}

type scope map[string]sym.YokType

func (s scope) Add(name string, yokType sym.YokType) error {
	setType, ok := s[name]
	if !ok {
		s[name] = yokType
		return nil
	}

	if setType == yokType {
		return nil
	}

	return fmt.Errorf("identifyer %s has type %s, but got type %s", name, setType, yokType)
}

func (s scope) GetType(name string) (sym.YokType, bool) {
	t, ok := s[name]
	return t, ok
}

type validateIdentifyers struct {
	scopeStack []scope
}

func NewValidateIdentifyer() *validateIdentifyers {
	ret := &validateIdentifyers{}
	ret.PushScope()
	return ret
}

func (v *validateIdentifyers) GetType(name string) (sym.YokType, bool) {
	for i := len(v.scopeStack) - 1; i >= 0; i++ {
		t, ok := v.scopeStack[i].GetType(name)
		if ok {
			return t, ok
		}
	}
	return sym.UnknownType, false
}

func (v *validateIdentifyers) Scope() scope {
	return v.scopeStack[len(v.scopeStack)-1]
}

func (v *validateIdentifyers) PushScope() {
	v.scopeStack = append(v.scopeStack, scope{})
}

func (v *validateIdentifyers) PopScope() {
	v.scopeStack = v.scopeStack[:len(v.scopeStack)-1]
}

// TODO: get this working (the validator probably needs to be a full struct an not just a function)
func (v *validateIdentifyers) check(stmt Stmt) error {
	fmt.Println("validing: ", stmt)

	switch s := stmt.(type) {
	case Assign:
		var setType sym.YokType
		if value, ok := s.SetTo.(Value); ok {
			setType = sym.TypeFromValue(value.Raw)
		}
		if ident, ok := s.SetTo.(Identifyer); ok {
			setType, ok = v.GetType(ident.Name)
			if !ok {
				return fmt.Errorf("identifyer '%s' has unknown type", ident.Name)
			}
		}

		return v.Scope().Add(s.Identifyer, setType)
	case If:
		if check, ok := s.Check.(Identifyer); ok {
			checkType, ok := v.GetType(check.Name)
			if !ok {
				return fmt.Errorf("identifyer '%s' has unknown type", check.Name)
			}

			if checkType != sym.BoolType {
				return fmt.Errorf("identifyer '%s' has type %s but it must be a bool type", check.Name, checkType)
			}
		}
	}

	return nil
}
