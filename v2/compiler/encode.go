package compiler

import (
	"fmt"
	"strings"

	"github.com/bjatkin/yok/ast/shast"
)

// encodeScript converts a yok script into a json string
func encodeScript(script *shast.Script) string {
	encoded := []string{}
	for _, stmt := range script.Statements {
		node := encodeNode(stmt.(shast.Node))
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

// encodeNode encodes shast.Nodes into json strings
func encodeNode(node shast.Node) string {
	switch node := node.(type) {
	case *shast.Comment:
		safeValue := strings.ReplaceAll(node.Value, "\"", "\\\"")
		return fmt.Sprintf(`{"Node": "comment", "Value": "%s"}`, safeValue)
	case *shast.NewLine:
		return `{"Node": "new line"}`
	case *shast.Assign:
		return fmt.Sprintf(`{
"Node": "assign",
"Identifier": "%s",
"Value": %s
}`,
			node.Identifier,
			encodeNode(node.Value),
		)
	case *shast.StmtExpr:
		return encodeNode(node.Expression)
	case *shast.String:
		safeValue := strings.ReplaceAll(node.Value, "\"", "\\\"")
		return fmt.Sprintf(`{"Node": "string", "Value": "%s"}`, safeValue)
	case *shast.Exec:
		args := `"Arguments": []`
		if len(node.Arguments) > 0 {
			encoded := encodeExprs(node.Arguments)
			args = fmt.Sprintf(`"Arguments": [
%s
]`, encoded)
		}

		redirects := `"Redirects": []`
		if len(node.Redirects) > 0 {
			encoded := []string{}
			for _, redirect := range node.Redirects {
				encoded = append(encoded, "\""+redirect.String()+"\"")
			}
			redirects = fmt.Sprintf(`"Redirects": [ %s ]`, strings.Join(encoded, ", "))
		}

		return fmt.Sprintf(`{
"Node": "execute",
"Command": "%s",
%s,
%s
}`,
			node.Command,
			args,
			redirects,
		)
	case *shast.Identifier:
		return fmt.Sprintf(`{"Node": "identifier", "Token": "%s"}`, node.Value)
	case *shast.ArithmeticCommand:
		return fmt.Sprintf(`{
"Node": "arithmetic command",
"Expression": %s
}`, encodeNode(node.Expression))
	case *shast.InfixExpr:
		return fmt.Sprintf(`{
"Node": "infix expression",
"Operator": "%s",
"Left": %s,
"Right": %s
}`,
			node.Operator,
			encodeNode(node.Left),
			encodeNode(node.Right),
		)
	case *shast.GroupExpr:
		return fmt.Sprintf(`{
"Node": "group expression",
"Expression": %s
}`,
			encodeNode(node.Expression))
	case *shast.If:
		body := "[]"
		if len(node.Statements) > 0 {
			body = fmt.Sprintf(`[
%s
]`, encodeStmts(node.Statements))
		}

		elseIfs := "[]"
		if len(node.ElseIfs) > 0 {
			elseIfs = fmt.Sprintf(`[
%s
]`, encodeElseIfs(node.ElseIfs))
		}

		elseBody := "[]"
		if len(node.ElseStatements) > 0 {
			elseBody = fmt.Sprintf(`[
%s
]`, encodeStmts(node.ElseStatements))
		}

		return fmt.Sprintf(`{
"Node": "if statement",
"Test": %s,
"Body": %s,
"ElseIfs": %s,
"ElseBody": %s
}`,
			encodeNode(node.Test),
			body,
			elseIfs,
			elseBody,
		)
	case *shast.TestCommand:
		return fmt.Sprintf(`{
"Node": "test statement",
"Expression": %s
}`, encodeNode(node.Expression))
	default:
		panic(fmt.Sprintf("can not encode sh node, unknown node type %T", node))
	}
}

// encodeElseIfs encodes a slice of ElseIf nodes into a list of json strings
func encodeElseIfs(elseIfs []shast.ElseIf) string {
	encoded := []string{}
	for _, elseIf := range elseIfs {
		body := "[]"
		if len(elseIf.Statements) > 0 {
			body = fmt.Sprintf(`[
%s
]`, encodeStmts(elseIf.Statements))
		}
		got := fmt.Sprintf(`{
"Node": "elif",
"Test": %s,
"Body": %s
}`,
			encodeNode(elseIf.Test),
			body,
		)
		encoded = append(encoded, got)
	}

	return strings.Join(encoded, ",\n")
}

// encodeExprs encodes a slice of expressions into a slice of json strings
func encodeExprs(exprs []shast.Expr) string {
	encoded := []string{}
	for _, expr := range exprs {
		node := expr.(shast.Node)
		encoded = append(encoded, encodeNode(node))
	}

	return strings.Join(encoded, ",\n")
}

// encodeStmts encodes a slice of statements into a slice of json strings
func encodeStmts(stmts []shast.Stmt) string {
	encoded := []string{}
	for _, stmt := range stmts {
		node := stmt.(shast.Node)
		encoded = append(encoded, encodeNode(node))
	}

	return strings.Join(encoded, ",\n")
}
