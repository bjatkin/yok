package compiler

import (
	"fmt"
	"strings"

	"github.com/bjatkin/yok/ast/shast"
	"github.com/bjatkin/yok/repr"
)

// encodeScript converts a yok script into a repr string
func encodeScript(script *shast.Script) string {
	array := repr.Array{}
	for _, stmt := range script.Statements {
		node := encodeNode(stmt.(shast.Node))
		array.AddValue(node)
	}

	return array.Render(0)
}

// encodeNode encodes a shast.Node into a repr.Value
func encodeNode(node shast.Node) repr.Value {
	switch node := node.(type) {
	case *shast.Comment:
		safeValue := strings.ReplaceAll(node.Value, "\"", "\\\"")
		return repr.NewObject(
			"Comment",
			repr.NewField("Value", repr.String(safeValue)),
		)
	case *shast.NewLine:
		return repr.NewObject("NewLine")
	case *shast.Assign:
		return repr.NewObject(
			"Assign",
			repr.NewField("Identifier", repr.String(node.Identifier)),
			repr.NewField("Value", encodeNode(node.Value)),
		)
	case *shast.StmtExpr:
		return encodeNode(node.Expression)
	case *shast.String:
		safeValue := strings.ReplaceAll(node.Value, "\"", "\\\"")
		return repr.NewObject(
			"String",
			repr.NewField("Value", repr.String(safeValue)),
		)
	case *shast.Exec:
		args := encodeExprs(node.Arguments)

		redirects := repr.Array{}
		for _, r := range node.Redirects {
			redirects.AddValue(repr.String(r.String()))
		}

		return repr.NewObject(
			"Execute",
			repr.NewField("Command", repr.String(node.Command)),
			repr.NewField("Arguments", args),
			repr.NewField("Redirects", redirects),
		)
	case *shast.Identifier:
		return repr.NewObject(
			"Identifier",
			repr.NewField("Token", repr.String(node.Value)),
		)
	case *shast.ArithmeticCommand:
		expression := encodeNode(node.Expression)
		return repr.NewObject(
			"ArithmeticCommand",
			repr.NewField("Expression", expression),
		)
	case *shast.InfixExpr:
		left := encodeNode(node.Left)
		right := encodeNode(node.Right)
		return repr.NewObject(
			"InfixExpression",
			repr.NewField("Operator", repr.String(node.Operator)),
			repr.NewField("Left", left),
			repr.NewField("Right", right),
		)
	case *shast.GroupExpr:
		expression := encodeNode(node.Expression)
		return repr.NewObject(
			"GroupExpression",
			repr.NewField("Expression", expression),
		)
	case *shast.If:
		test := encodeNode(node.Test)
		body := encodeStmts(node.Statements)
		elseIfs := encodeElseIfs(node.ElseIfs)
		elseBody := encodeStmts(node.ElseStatements)

		return repr.NewObject(
			"IfStatement",
			repr.NewField("Test", test),
			repr.NewField("Body", body),
			repr.NewField("ElseIfs", elseIfs),
			repr.NewField("ElseBody", elseBody),
		)
	case *shast.TestCommand:
		expression := encodeNode(node.Expression)
		return repr.NewObject(
			"TestStatement",
			repr.NewField("Expression", expression),
		)
	case *shast.ParamaterExpansion:
		expression := encodeNode(node.Expression)
		return repr.NewObject(
			"ParamaterExpansion",
			repr.NewField("Expression", expression),
		)
	case *shast.ParameterLength:
		paramater := encodeNode(node.Paramater)
		return repr.NewObject(
			"ParamaterLenght",
			repr.NewField("Paramater", paramater),
		)
	case *shast.ParamaterReplace:
		paramater := encodeNode(node.Paramater)
		find := encodeNode(node.Find)
		replace := encodeNode(node.Replace)
		return repr.NewObject(
			"ParamaterReplace",
			repr.NewField("ReplaceAll", repr.Bool(node.ReplaceAll)),
			repr.NewField("Paramater", paramater),
			repr.NewField("Find", find),
			repr.NewField("Replace", replace),
		)
	case *shast.ParamaterRemoveFix:
		paramater := encodeNode(node.Paramater)
		remove := encodeNode(node.Remove)
		return repr.NewObject(
			"ParamaterRemoveFix",
			repr.NewField("RemovePrefix", repr.Bool(node.RemovePrefix)),
			repr.NewField("Paramater", paramater),
			repr.NewField("Remove", remove),
		)
	default:
		panic(fmt.Sprintf("can not encode sh node, unknown node type %T", node))
	}
}

// encodeElseIfs encodes a slice of ElseIf nodes into a repr.Array
func encodeElseIfs(elseIfs []shast.ElseIf) repr.Array {
	array := repr.Array{}
	for _, elseIf := range elseIfs {
		body := encodeStmts(elseIf.Statements)
		test := encodeNode(elseIf.Test)
		node := repr.NewObject(
			"Elif",
			repr.NewField("Test", test),
			repr.NewField("Body", body),
		)
		array.AddValue(node)
	}

	return array
}

// encodeExprs encodes a slice of expressions into a repr.Array
func encodeExprs(exprs []shast.Expr) repr.Array {
	array := repr.Array{}
	for _, expr := range exprs {
		node := expr.(shast.Node)
		encoded := encodeNode(node)
		array.AddValue(encoded)
	}

	return array
}

// encodeStmts encodes a slice of statements into a repr.Array
func encodeStmts(stmts []shast.Stmt) repr.Array {
	array := repr.Array{}
	for _, stmt := range stmts {
		node := stmt.(shast.Node)
		encoded := encodeNode(node)
		array.AddValue(encoded)
	}

	return array
}
