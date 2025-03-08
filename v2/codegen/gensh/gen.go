package gensh

import (
	"fmt"
	"strings"

	"github.com/bjatkin/yok/ast/shast"
)

const indentToken = "    "

func Generate(script *shast.Script) string {
	scriptLines := []string{"#!/bin/sh", ""}
	for _, stmt := range script.Statements {
		node := generateStmt(stmt, 0)
		scriptLines = append(scriptLines, node)
	}

	return strings.Join(scriptLines, "\n")
}

func generateExpr(expr shast.Expr) string {
	switch expr := expr.(type) {
	case *shast.String:
		value := expr.Value
		if !strings.Contains(expr.Value, " ") &&
			!strings.Contains(expr.Value, "\t") {
			// if the string has no white space, sh allows for dropping the double quotes
			value = strings.TrimPrefix(value, "\"")
			value = strings.TrimSuffix(value, "\"")
		}
		return value
	case *shast.Exec:
		args := []string{}
		for _, arg := range expr.Arguments {
			n := generateExpr(arg)
			args = append(args, n)
		}

		for _, redirect := range expr.Redirects {
			args = append(args, redirect.String())
		}

		return expr.Command + " " + strings.Join(args, " ")
	case *shast.Identifier:
		if expr.AsString {
			return "\"$" + expr.Value + "\""
		}

		return "$" + expr.Value
	case *shast.ArithmeticCommand:
		inner := generateExpr(expr.Expression)
		return "$(( " + inner + " ))"
	case *shast.InfixExpr:
		left := generateExpr(expr.Left)
		right := generateExpr(expr.Right)
		return fmt.Sprintf("%s %s %s", left, expr.Operator, right)
	case *shast.GroupExpr:
		inner := generateExpr(expr.Expression)
		return "( " + inner + " )"
	default:
		panic(fmt.Sprintf("can not gen sh code, unknown expr type %T", expr))
	}
}

func generateStmt(stmt shast.Stmt, indentDepth int) string {
	indent := strings.Repeat(indentToken, indentDepth)
	switch stmt := stmt.(type) {
	case *shast.Comment:
		return indent + stmt.Value
	case *shast.NewLine:
		return ""
	case *shast.Assign:
		value := generateExpr(stmt.Value)
		return indent + stmt.Identifier + "=" + value
	case *shast.StmtExpr:
		return indent + generateExpr(stmt.Expression)
	case *shast.If:
		// TODO: clean all this code up, there's got to be a better way to genrate code with correct
		// levels of indentation than these messy format strings!
		test := generateStmt(stmt.Test, indentDepth)

		body := []string{}
		for _, stmt := range stmt.Statements {
			line := generateStmt(stmt, indentDepth+1)
			body = append(body, line)
		}

		if len(stmt.ElseIfs) == 0 && len(stmt.ElseStatements) == 0 {
			return fmt.Sprintf(
				"%sif %s; then\n%s\n%sfi",
				indent,
				test,
				strings.Join(body, "\n"),
				indent,
			)
		}

		elseIfs := []string{}
		for _, elseIf := range stmt.ElseIfs {
			test := generateStmt(elseIf.Test, indentDepth)
			body := []string{}
			for _, stmt := range stmt.Statements {
				line := generateStmt(stmt, indentDepth+1)
				body = append(body, line)
			}

			elseIfs = append(elseIfs, fmt.Sprintf(
				"%selif %s; then\n%s",
				indent,
				test,
				strings.Join(body, "\n"),
			))
		}

		if len(stmt.ElseStatements) == 0 {
			return fmt.Sprintf(
				"%sif %s; then\n%s%s%s\n%sfi",
				indent,
				test,
				strings.Join(body, "\n"),
				indent,
				strings.Join(elseIfs, "\n"),
				indent,
			)
		}

		elseBody := []string{}
		for _, stmt := range stmt.ElseStatements {
			line := generateStmt(stmt, indentDepth+1)
			elseBody = append(elseBody, line)
		}

		if len(stmt.ElseIfs) == 0 {
			return fmt.Sprintf(
				"%sif %s; then\n%s\n%selse\n%s\n%sfi",
				indent,
				test,
				strings.Join(body, "\n"),
				indent,
				strings.Join(elseBody, "\n"),
				indent,
			)
		}

		return fmt.Sprintf(
			"%sif %s; then\n%s\n%s%s\n%selse\n%s\n%sfi",
			indent,
			test,
			strings.Join(body, "\n"),
			indent,
			strings.Join(elseIfs, "\n"),
			indent,
			strings.Join(elseBody, "\n"),
			indent,
		)
	case *shast.TestCommand:
		expr := generateExpr(stmt.Expression)
		return "[ " + expr + " ]"
	default:
		panic(fmt.Sprintf("can not gen sh code, unknown stmt type %T", stmt))
	}
}
