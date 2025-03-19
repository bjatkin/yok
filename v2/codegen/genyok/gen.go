package genyok

import (
	"fmt"
	"strings"

	"github.com/bjatkin/yok/ast/yokast"
)

const indentToken = "    "

// TODO: I might want to use a client to generate yok code since I need to pass the source file in.
// That could technically let me do the parsing as well
func Generate(script *yokast.Script, source []byte) string {
	scriptLines := []string{}
	for _, stmt := range script.Statements {
		node := generateStmt(stmt, 0, source)
		scriptLines = append(scriptLines, node)
	}

	return strings.Join(scriptLines, "\n") + "\n"
}

func generateExpr(expr yokast.Expr, source []byte) string {
	switch expr := expr.(type) {
	case *yokast.InfixExpr:
		left := generateExpr(expr.Left, source)
		op := expr.Operator.Value(source)
		right := generateExpr(expr.Right, source)
		return fmt.Sprintf("%s %s %s", left, op, right)
	case *yokast.Identifier:
		return expr.Token.Value(source)
	case *yokast.Atom:
		return expr.Token.Value(source)
	case *yokast.Call:
		args := []string{}
		for _, arg := range expr.Arguments {
			a := generateExpr(arg, source)
			args = append(args, a)
		}
		funcName := expr.Identifier.Token.Value(source)
		return fmt.Sprintf("%s(%s)", funcName, strings.Join(args, ","))
	case *yokast.String:
		return expr.Token.Value(source)
	default:
		panic(fmt.Sprintf("can not get yok code, unknown expr type %T", expr))
	}
}

func generateStmt(stmt yokast.Stmt, indentDepth int, source []byte) string {
	indent := strings.Repeat(indentToken, indentDepth)
	switch stmt := stmt.(type) {
	case *yokast.Comment:
		return indent + stmt.Token.Value(source)
	case *yokast.If:
		test := generateExpr(stmt.Test, source)
		body := generateStmt(stmt.Body, indentDepth+1, source)

		if stmt.ElseBody == nil {
			return indent + "if " + test + " {\n" + body + "\n}"
		}

		elseBody := generateStmt(stmt.ElseBody, indentDepth+1, source)
		return indent + "if " + test + " {\n" +
			body + "\n} else {\n" +
			elseBody + "\n}"
	case *yokast.Block:
		statements := []string{}
		for _, statment := range stmt.Statements {
			s := generateStmt(statment, indentDepth, source)
			statements = append(statements, indent+s)
		}
		return strings.Join(statements, "\n")
	case *yokast.StmtExpr:
		expr := generateExpr(stmt.Expression, source)
		return expr
	default:
		panic(fmt.Sprintf("can not get yok code, unknown stmt type %T", stmt))
	}
}
