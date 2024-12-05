package parser

import (
	"fmt"
	"strings"

	"github.com/bjatkin/yok/ast/yokast"
	"github.com/bjatkin/yok/token"
)

// encodeToken converts a single token into a json string
func encodeToken(t token.Token, source []byte) string {
	value := string(source[t.Pos : int(t.Pos)+t.Len])
	value = strings.ReplaceAll(value, "\"", "\\\"")
	if value == "\r\n" {
		value = "\\r\\n"
	}
	if value == "\n" {
		value = "\\n"
	}

	return fmt.Sprintf(
		`{"Type": "%s", "Pos": %d, "Value": "%s"}`,
		t.Type.String(),
		t.Pos,
		value,
	)
}

// encodeTokens converts a slice of tokens into a json array
func encodeTokens(tokens []token.Token, source []byte) string {
	encoded := []string{}
	for _, t := range tokens {
		encoded = append(encoded, encodeToken(t, source))
	}

	return fmt.Sprintf("[\n    %s\n]", strings.Join(encoded, ",\n    "))
}

// encodeScript converts a yok script into a json string
func encodeScript(script *yokast.Script, source []byte) string {
	encoded := []string{}
	for _, stmt := range script.Statements {
		node := encodeNode(stmt.(yokast.Node), source)
		encoded = append(encoded, node)
	}

	encodedScript := strings.Join(encoded, ",\n")
	return indentLines("[\n" + encodedScript + "\n]")
}

// indentLines moves through a json string line by line, adding the correct indent.
// turns out it's easier to do this as a pos-processing step rather than doing it while
// generating the json structure itself.
func indentLines(str string) string {
	lines := strings.Split(str, "\n")
	depth := 0
	indentedLines := []string{}
	for _, line := range lines {
		if line == "]" ||
			line == "]," ||
			line == "}" ||
			line == "}," {
			depth--
		}

		indent := strings.Repeat("    ", depth)
		indentedLines = append(indentedLines, indent+line)

		if strings.HasSuffix(line, "[") ||
			strings.HasSuffix(line, "{") {
			depth++
		}
	}

	return strings.Join(indentedLines, "\n")
}

// encodeNode encodes yokast.Nodes into json strings
func encodeNode(node yokast.Node, source []byte) string {
	switch node := node.(type) {
	case *yokast.Comment:
		safeValue := strings.ReplaceAll(node.Value, "\"", "\\\"")
		return fmt.Sprintf(`{"Node": "comment", "Value": "%s"}`, safeValue)
	case *yokast.NewLine:
		return `{"Node": "new line"}`
	case *yokast.Assign:
		return fmt.Sprintf(`{
"Node": "assign",
"Identifier": %s,
"Value": %s
}`,
			encodeToken(node.Identifier, source),
			encodeNode(node.Value, source),
		)
	case *yokast.StmtExpr:
		return encodeNode(node.Expression, source)
	case *yokast.String:
		safeValue := strings.ReplaceAll(node.Value, "\"", "\\\"")
		return fmt.Sprintf(`{"Node": "string", "Value": "%s"}`, safeValue)
	case *yokast.Atom:
		return fmt.Sprintf(`{"Node": "atom", "Value": "%s"}`, node.Value)
	case *yokast.Call:
		identifier := encodeNode(node.Identifier, source)

		if len(node.Arguments) == 0 {
			return fmt.Sprintf(`{
"Node": "function call",
"Identifier": %s,
"Arguments": []
}`, identifier)
		}

		return fmt.Sprintf(`{
"Node": "function call",
"Identifier": %s,
"Arguments": [
%s
]
}`,
			identifier,
			encodeExprs(node.Arguments, source),
		)
	case *yokast.Identifier:
		return fmt.Sprintf(`{
"Node": "identifier",
"Token": %s
}`,
			encodeToken(node.Token, source),
		)
	case *yokast.InfixExpr:
		return fmt.Sprintf(`{
"Node": "infix expression",
"Operator": %s,
"Left": %s,
"Right": %s
}`,
			encodeToken(node.Operator, source),
			encodeNode(node.Left, source),
			encodeNode(node.Right, source),
		)
	case *yokast.GroupExpr:
		return fmt.Sprintf(`{
"Node": "grouped expression",
"Expression": %s
}`, encodeNode(node.Expression, source))
	case *yokast.If:
		return fmt.Sprintf(`{
"Node": "if statement",
"Test": %s,
"Body": %s,
"ElseBody": %s
}`,
			encodeNode(node.Test.(yokast.Node), source),
			encodeNode(node.Body, source),
			encodeNode(node.ElseBody, source),
		)
	case *yokast.Block:
		// TODO: this is kinda a hack but trying to determine if the underlying value of an
		// interface is kinda a mess. This is really the only node that can be nil in a valid
		// ast right now so I'm just gonna leave this here for now.
		// I could just make an empty *Block for the ElseBody in the if statement but a nil
		// value seems more correct...
		if node == nil {
			return "null"
		}

		if len(node.Statements) == 0 {
			return `{
"Node": "block",
"Statements": []
}`
		}

		return fmt.Sprintf(`{
"Node": "block",
"Statements": [
%s
]
}`, encodeStmts(node.Statements, source))
	default:
		panic(fmt.Sprintf("failed to encode yok ast node, unknown type %T", node))
	}
}

// encodeExprs encodes a slice of expressions into a slice of json strings
func encodeExprs(exprs []yokast.Expr, source []byte) string {
	encoded := []string{}
	for _, expr := range exprs {
		node := expr.(yokast.Node)
		encoded = append(encoded, encodeNode(node, source))
	}

	return strings.Join(encoded, ",\n")
}

// encodeStmts encodes a slice of statements into a slice of json strings
func encodeStmts(stmts []yokast.Stmt, source []byte) string {
	encoded := []string{}
	for _, stmt := range stmts {
		node := stmt.(yokast.Node)
		encoded = append(encoded, encodeNode(node, source))
	}

	return strings.Join(encoded, ",\n")
}
