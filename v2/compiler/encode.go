package compiler

import (
	"fmt"
	"strings"

	"github.com/bjatkin/yok/ast/shast"
	"github.com/bjatkin/yok/json"
)

// encodeScript converts a yok script into a json string
func encodeScript(script *shast.Script) string {
	array := json.Array{}
	for _, stmt := range script.Statements {
		node := encodeNode(stmt.(shast.Node))
		array.AddItem(node)
	}

	return array.Render(0)
}

func newNode(nodeName string, fields ...json.Field) json.Object {
	nodeField := json.NewField("Node", json.String(nodeName))
	node := json.NewObject(nodeField)
	node.AddFields(fields...)
	return node
}

// encodeNode encodes shast.Nodes into json strings
func encodeNode(node shast.Node) json.Value {
	switch node := node.(type) {
	case *shast.Comment:
		safeValue := strings.ReplaceAll(node.Value, "\"", "\\\"")
		return newNode(
			"comment",
			json.NewField("Value", json.String(safeValue)),
		)
	case *shast.NewLine:
		return newNode("new line")
	case *shast.Assign:
		return newNode(
			"assign",
			json.NewField("Identifier", json.String(node.Identifier)),
			json.NewField("Value", encodeNode(node.Value)),
		)
	case *shast.StmtExpr:
		return encodeNode(node.Expression)
	case *shast.String:
		safeValue := strings.ReplaceAll(node.Value, "\"", "\\\"")
		return newNode(
			"string",
			json.NewField("Value", json.String(safeValue)),
		)
	case *shast.Exec:
		args := encodeExprs(node.Arguments)

		redirects := json.Array{}
		for _, r := range node.Redirects {
			redirects.AddItem(json.String(r.String()))
		}

		return newNode(
			"execute",
			json.NewField("Command", json.String(node.Command)),
			json.NewField("Arguments", args),
			json.NewField("Redirects", redirects),
		)
	case *shast.Identifier:
		return newNode(
			"identifier",
			json.NewField("Token", json.String(node.Value)),
		)
	case *shast.ArithmeticCommand:
		expression := encodeNode(node.Expression)
		return newNode(
			"arithmetic command",
			json.NewField("Expression", expression),
		)
	case *shast.InfixExpr:
		left := encodeNode(node.Left)
		right := encodeNode(node.Right)
		return newNode(
			"infix expression",
			json.NewField("Operator", json.String(node.Operator)),
			json.NewField("Left", left),
			json.NewField("Right", right),
		)
	case *shast.GroupExpr:
		expression := encodeNode(node.Expression)
		return newNode(
			"group expression",
			json.NewField("Expression", expression),
		)
	case *shast.If:
		test := encodeNode(node.Test)
		body := encodeStmts(node.Statements)
		elseIfs := encodeElseIfs(node.ElseIfs)
		elseBody := encodeStmts(node.ElseStatements)

		return newNode(
			"if statement",
			json.NewField("Test", test),
			json.NewField("Body", body),
			json.NewField("ElseIfs", elseIfs),
			json.NewField("ElseBody", elseBody),
		)
	case *shast.TestCommand:
		expression := encodeNode(node.Expression)
		return newNode(
			"test statement",
			json.NewField("Expression", expression),
		)
	default:
		panic(fmt.Sprintf("can not encode sh node, unknown node type %T", node))
	}
}

// encodeElseIfs encodes a slice of ElseIf nodes into a list of json strings
func encodeElseIfs(elseIfs []shast.ElseIf) json.Array {
	array := json.Array{}
	for _, elseIf := range elseIfs {
		body := encodeStmts(elseIf.Statements)
		test := encodeNode(elseIf.Test)
		node := newNode(
			"elif",
			json.NewField("Test", test),
			json.NewField("Body", body),
		)
		array.AddItem(node)
	}

	return array
}

// encodeExprs encodes a slice of expressions into a slice of json strings
func encodeExprs(exprs []shast.Expr) json.Array {
	array := json.Array{}
	for _, expr := range exprs {
		node := expr.(shast.Node)
		encoded := encodeNode(node)
		array.AddItem(encoded)
	}

	return array
}

// encodeStmts encodes a slice of statements into a slice of json strings
func encodeStmts(stmts []shast.Stmt) json.Array {
	array := json.Array{}
	for _, stmt := range stmts {
		node := stmt.(shast.Node)
		encoded := encodeNode(node)
		array.AddItem(encoded)
	}

	return array
}
