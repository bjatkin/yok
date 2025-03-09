package parser

import (
	"fmt"
	"strings"

	"github.com/bjatkin/yok/ast/yokast"
	"github.com/bjatkin/yok/json"
	"github.com/bjatkin/yok/token"
)

// encodeToken converts a single token into a json.Object
func encodeToken(t token.Token, source []byte) json.Object {
	value := string(source[t.Pos : int(t.Pos)+t.Len])
	value = strings.ReplaceAll(value, "\"", "\\\"")
	if value == "\r\n" {
		value = "\\r\\n"
	}
	if value == "\n" {
		value = "\\n"
	}

	return json.NewObject(
		json.NewField("Type", json.String(t.Type.String())),
		json.NewField("Pos", json.Int(t.Pos)),
		json.NewField("Value", json.String(value)),
	)
}

// encodeTokens converts a slice of tokens into a json.Array
func encodeTokens(tokens []token.Token, source []byte) json.Array {
	array := json.Array{}
	for _, t := range tokens {
		encoded := encodeToken(t, source)
		array.AddValue(encoded)
	}

	return array
}

// encodeScript converts a yok script into a json string
func encodeScript(script *yokast.Script, source []byte) string {
	array := json.Array{}
	for _, stmt := range script.Statements {
		node := encodeNode(stmt.(yokast.Node), source)
		array.AddValue(node)
	}

	return array.Render(0)
}

// newNode cretes a new json.Object where the "Node" field is set to the given name
// if additional fields are passed they will be added to the object
func newNode(name string, fields ...json.Field) json.Object {
	nodeField := json.NewField("Node", json.String(name))
	node := json.NewObject(nodeField)
	node.AddFields(fields...)
	return node
}

// encodeNode encodes a yokast.Node into a json.Value
func encodeNode(node yokast.Node, source []byte) json.Value {
	switch node := node.(type) {
	case *yokast.Comment:
		safeValue := strings.ReplaceAll(node.Value, "\"", "\\\"")
		return newNode(
			"comment",
			json.NewField("Value", json.String(safeValue)),
		)
	case *yokast.NewLine:
		return newNode("new line")
	case *yokast.Assign:
		identifier := encodeToken(node.Identifier, source)
		value := encodeNode(node.Value, source)
		return newNode(
			"assign",
			json.NewField("Identifier", identifier),
			json.NewField("Value", value),
		)
	case *yokast.StmtExpr:
		return encodeNode(node.Expression, source)
	case *yokast.String:
		safeValue := strings.ReplaceAll(node.Value, "\"", "\\\"")
		return newNode(
			"string",
			json.NewField("Value", json.String(safeValue)),
		)
	case *yokast.Atom:
		return newNode(
			"atom",
			json.NewField("Value", json.String(node.Value)),
		)
	case *yokast.Call:
		identifier := encodeNode(node.Identifier, source)
		args := encodeExprs(node.Arguments, source)
		return newNode(
			"function call",
			json.NewField("Identifier", identifier),
			json.NewField("Arguments", args),
		)
	case *yokast.Identifier:
		token := encodeToken(node.Token, source)
		return newNode(
			"identifier",
			json.NewField("Token", token),
		)
	case *yokast.InfixExpr:
		operator := encodeToken(node.Operator, source)
		left := encodeNode(node.Left, source)
		right := encodeNode(node.Right, source)
		return newNode(
			"infix expression",
			json.NewField("Operator", operator),
			json.NewField("Left", left),
			json.NewField("Right", right),
		)
	case *yokast.GroupExpr:
		expression := encodeNode(node.Expression, source)
		return newNode(
			"grouped expression",
			json.NewField("Expression", expression),
		)
	case *yokast.If:
		test := encodeNode(node.Test.(yokast.Node), source)
		body := encodeNode(node.Body, source)
		elseIfs := encodeElseIfs(node.ElseIfs, source)
		elseBody := encodeNode(node.ElseBody, source)
		return newNode(
			"if statement",
			json.NewField("Test", test),
			json.NewField("Body", body),
			json.NewField("ElseIfs", elseIfs),
			json.NewField("ElseBody", elseBody),
		)
	case *yokast.Block:
		if node == nil {
			return json.Null{}
		}

		statements := encodeStmts(node.Statements, source)
		return newNode(
			"block",
			json.NewField("Statements", statements),
		)
	default:
		panic(fmt.Sprintf("failed to encode yok ast node, unknown type %T", node))
	}
}

// encodeElseIfs encodes a slice of ElseIf nodes into a json.Array
func encodeElseIfs(elseIfs []yokast.ElseIf, source []byte) json.Array {
	array := json.Array{}
	for _, elseIf := range elseIfs {
		test := encodeNode(elseIf.Test, source)
		body := encodeNode(elseIf.Body, source)
		node := newNode(
			"else if",
			json.NewField("Test", test),
			json.NewField("Body", body),
		)
		array.AddValue(node)
	}

	return array
}

// encodeExprs encodes a slice of expressions into a json.Array
func encodeExprs(exprs []yokast.Expr, source []byte) json.Array {
	array := json.Array{}
	for _, expr := range exprs {
		node := expr.(yokast.Node)
		encoded := encodeNode(node, source)
		array.AddValue(encoded)
	}

	return array
}

// encodeStmts encodes a slice of statements into a json.Array
func encodeStmts(stmts []yokast.Stmt, source []byte) json.Array {
	array := json.Array{}
	for _, stmt := range stmts {
		node := stmt.(yokast.Node)
		encoded := encodeNode(node, source)
		array.AddValue(encoded)

	}

	return array
}
