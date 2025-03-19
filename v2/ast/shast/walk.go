package shast

import "fmt"

type Visitor interface {
	Visit(Node) Visitor
}

func walkSlice[N Node](v Visitor, slice []N) {
	for _, node := range slice {
		Walk(v, node)
	}
}

func Walk(v Visitor, node Node) {
	if v = v.Visit(node); v == nil {
		return
	}

	switch n := node.(type) {
	case *Script:
		walkSlice(v, n.Statements)
	case *Comment:
		// nothing to walk
	case *NewLine:
		// nothing to walk
	case *Assign:
		Walk(v, n.Value)
	case *If:
		Walk(v, n.Test)
		walkSlice(v, n.Statements)
		for _, elseIf := range n.ElseIfs {
			Walk(v, elseIf.Test)
			walkSlice(v, elseIf.Statements)
		}
		if n.ElseStatements != nil {
			walkSlice(v, n.ElseStatements)
		}
	case *StmtExpr:
		Walk(v, n.Expression)
	case *String:
		// nothing to walk
	case *Exec:
		walkSlice(v, n.Arguments)
	case *Identifier:
		// nothing to walk
	case *TestCommand:
		Walk(v, n.Expression)
	case *ArithmeticCommand:
		Walk(v, n.Expression)
	case *InfixExpr:
		Walk(v, n.Left)
		Walk(v, n.Right)
	case *GroupExpr:
		Walk(v, n.Expression)
	default:
		panic(fmt.Sprintf("failed to walk the AST, uknown node %T", n))
	}

	v.Visit(nil)
}
