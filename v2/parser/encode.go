package parser

import (
	"fmt"
	"strings"

	"github.com/bjatkin/yok/ast/yokast"
	"github.com/bjatkin/yok/repr"
	"github.com/bjatkin/yok/token"
)

// encodeToken converts a single token into a repr.Object
func encodeToken(t token.Token, source []byte) repr.Object {
	value := string(source[t.Pos : int(t.Pos)+t.Len])
	value = strings.ReplaceAll(value, "\"", "\\\"")
	if value == "\r\n" {
		value = "\\r\\n"
	}
	if value == "\n" {
		value = "\\n"
	}

	return repr.NewObject(
		"Token",
		repr.NewField("Type", repr.String(t.Type.String())),
		repr.NewField("Pos", repr.Int(t.Pos)),
		repr.NewField("Value", repr.String(value)),
	)
}

// encodeTokens converts a slice of tokens into a repr.Array
func encodeTokens(tokens []token.Token, source []byte) repr.Array {
	array := repr.Array{}
	for _, t := range tokens {
		encoded := encodeToken(t, source)
		array.AddValue(encoded)
	}

	return array
}

// encodeScript converts a yok script into a repr string
func encodeScript(script *yokast.Script, source []byte) string {
	array := repr.Array{}
	for _, stmt := range script.Statements {
		node := encodeNode(stmt.(yokast.Node), source)
		array.AddValue(node)
	}

	return array.Render(0)
}

// encodeNode encodes a yokast.Node into a repr.Value
func encodeNode(node yokast.Node, source []byte) repr.Value {
	switch node := node.(type) {
	case *yokast.Comment:
		safeValue := strings.ReplaceAll(node.Token.Value(source), "\"", "\\\"")
		return repr.NewObject(
			"Comment",
			repr.NewField("Value", repr.String(safeValue)),
		)
	case *yokast.NewLine:
		return repr.NewObject("NewLine")
	case *yokast.Assign:
		identifier := encodeNode(node.Identifier, source)
		value := encodeNode(node.Value, source)
		return repr.NewObject(
			"Assign",
			repr.NewField("Identifier", identifier),
			repr.NewField("Value", value),
		)
	case *yokast.StmtExpr:
		return encodeNode(node.Expression, source)
	case *yokast.String:
		safeValue := strings.ReplaceAll(node.Value(source), "\"", "\\\"")
		return repr.NewObject(
			"String",
			repr.NewField("Value", repr.String(safeValue)),
		)
	case *yokast.Atom:
		return repr.NewObject(
			"Atom",
			repr.NewField("Value", repr.String(node.Token.Value(source))),
		)
	case *yokast.Call:
		identifier := encodeNode(node.Identifier, source)
		args := encodeExprs(node.Arguments, source)
		return repr.NewObject(
			"FunctionCall",
			repr.NewField("Identifier", identifier),
			repr.NewField("Arguments", args),
		)
	case *yokast.Identifier:
		token := encodeToken(node.Token, source)
		return repr.NewObject(
			"Identifier",
			repr.NewField("Token", token),
		)
	case *yokast.InfixExpr:
		operator := encodeToken(node.Operator, source)
		left := encodeNode(node.Left, source)
		right := encodeNode(node.Right, source)
		return repr.NewObject(
			"InfixExpression",
			repr.NewField("Operator", operator),
			repr.NewField("Left", left),
			repr.NewField("Right", right),
		)
	case *yokast.GroupExpr:
		expression := encodeNode(node.Expression, source)
		return repr.NewObject(
			"GroupedExpression",
			repr.NewField("Expression", expression),
		)
	case *yokast.If:
		test := encodeNode(node.Test.(yokast.Node), source)
		body := encodeNode(node.Body, source)
		elseIfs := encodeElseIfs(node.ElseIfs, source)
		elseBody := encodeNode(node.ElseBody, source)
		return repr.NewObject(
			"IfStatement",
			repr.NewField("Test", test),
			repr.NewField("Body", body),
			repr.NewField("ElseIfs", elseIfs),
			repr.NewField("ElseBody", elseBody),
		)
	case *yokast.Block:
		if node == nil {
			return repr.Nil{}
		}

		statements := encodeStmts(node.Statements, source)
		return repr.NewObject(
			"Block",
			repr.NewField("Statements", statements),
		)
	default:
		panic(fmt.Sprintf("failed to encode yok ast node, unknown type %T", node))
	}
}

// encodeElseIfs encodes a slice of ElseIf nodes into a repr.Array
func encodeElseIfs(elseIfs []yokast.ElseIf, source []byte) repr.Array {
	array := repr.Array{}
	for _, elseIf := range elseIfs {
		test := encodeNode(elseIf.Test, source)
		body := encodeNode(elseIf.Body, source)
		node := repr.NewObject(
			"ElseIf",
			repr.NewField("Test", test),
			repr.NewField("Body", body),
		)
		array.AddValue(node)
	}

	return array
}

// encodeExprs encodes a slice of expressions into a repr.Array
func encodeExprs(exprs []yokast.Expr, source []byte) repr.Array {
	array := repr.Array{}
	for _, expr := range exprs {
		node := expr.(yokast.Node)
		encoded := encodeNode(node, source)
		array.AddValue(encoded)
	}

	return array
}

// encodeStmts encodes a slice of statements into a repr.Array
func encodeStmts(stmts []yokast.Stmt, source []byte) repr.Array {
	array := repr.Array{}
	for _, stmt := range stmts {
		node := stmt.(yokast.Node)
		encoded := encodeNode(node, source)
		array.AddValue(encoded)

	}

	return array
}
