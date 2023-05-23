package ast

import (
	"fmt"

	"github.com/bjatkin/yok/sym"
)

// TODO: the pattern that I'm moving towards is visiting every node, returning structured information
// and then handing that structure at the call site. I should apply that pattern here as well.

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

func (v *validateIdentifyers) check(stmt Stmt) error {
	switch s := stmt.(type) {
	case Assign:
		var setType sym.YokType
		switch t := s.SetTo.(type) {
		case Value:
			setType = sym.TypeFromValue(t.Raw)
		case Identifyer:
			var ok bool
			setType, ok = v.GetType(t.Name)
			if !ok {
				return fmt.Errorf("identifyer '%s' has unknown type", t.Name)
			}
		case BinaryExpr:
			var ok bool
			setType, ok = v.getBinaryExprType(t)
			if !ok {
				return fmt.Errorf("expression '%s' has unknown type", t.Yok())
			}
		default:
			return fmt.Errorf("invalid set to type: %T", t)
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

func (v *validateIdentifyers) getBinaryExprType(expr BinaryExpr) (sym.YokType, bool) {
	var leftType sym.YokType
	switch t := expr.Left.(type) {
	case Identifyer:
		var ok bool
		leftType, ok = v.GetType(t.Name)
		if !ok {
			return "", false
		}
	case Value:
		leftType = sym.TypeFromValue(t.Raw)
	default:
		panic("unknown left type")
	}

	var rightType sym.YokType
	switch t := expr.Right.(type) {
	case Identifyer:
		var ok bool
		rightType, ok = v.GetType(t.Name)
		if !ok {
			return "", false
		}
	case Value:
		rightType = sym.TypeFromValue(t.Raw)
	case BinaryExpr:
		var ok bool
		rightType, ok = v.getBinaryExprType(t)
		if !ok {
			return "", false
		}
	default:
		panic("unknown right type")
	}

	if leftType != rightType {
		return "", false
	}

	return leftType, true
}
