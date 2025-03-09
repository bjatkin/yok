package gensh

import (
	"fmt"
	"strings"

	"github.com/bjatkin/yok/ast/shast"
)

// Generate takes a shast.Script and renderes it into a well formated shell script
func Generate(script *shast.Script) string {
	scriptBuilder := newCodeBuilder("#!/bin/sh")
	scriptBuilder.addLine("")

	bodyBuilder := generateStmts(script.Statements)
	scriptBuilder.addUnits(bodyBuilder.units)

	return scriptBuilder.render()
}

// generateExpr takes an shast.Expr and renders it into a well formated shell string
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

// generateStmt takes an shast.Stmt and converts it into a codeBuilder
func generateStmt(stmt shast.Stmt) codeBuilder {
	switch stmt := stmt.(type) {
	case *shast.Comment:
		return newCodeBuilder(stmt.Value)
	case *shast.NewLine:
		return newCodeBuilder("")
	case *shast.Assign:
		value := generateExpr(stmt.Value)
		return newCodeBuilder(stmt.Identifier + "=" + value)
	case *shast.StmtExpr:
		expr := generateExpr(stmt.Expression)
		return newCodeBuilder(expr)
	case *shast.If:
		test := generateStmt(stmt.Test)
		ifUnit := newCodeUnitf("if %s; then", test.render())

		for _, stmt := range stmt.Statements {
			line := generateStmt(stmt)
			ifUnit.addChildren(line.units)
		}

		ifBuilder := codeBuilder{}
		ifBuilder.addUnit(ifUnit)
		for _, elseIf := range stmt.ElseIfs {
			test := generateStmt(elseIf.Test)
			elseIfUnit := newCodeUnitf("elif %s; then", test.render())

			bodyBuilder := generateStmts(stmt.Statements)
			elseIfUnit.addChildren(bodyBuilder.units)

			ifBuilder.addUnit(elseIfUnit)
		}

		if stmt.ElseStatements != nil {
			elseUnit := codeUnit{line: "else"}
			bodyBuilder := generateStmts(stmt.ElseStatements)
			elseUnit.addChildren(bodyBuilder.units)

			ifBuilder.addUnit(elseUnit)
		}

		ifBuilder.addLine("fi")
		return ifBuilder

	case *shast.TestCommand:
		expr := generateExpr(stmt.Expression)
		return newCodeBuilder("[ " + expr + " ]")
	default:
		panic(fmt.Sprintf("can not gen sh code, unknown stmt type %T", stmt))
	}
}

// generateStmts takes a slice of shast.Stmt and converts it into a codeBuilder
func generateStmts(statements []shast.Stmt) codeBuilder {
	builder := codeBuilder{}
	for _, stmt := range statements {
		line := generateStmt(stmt)
		builder.addUnits(line.units)
	}
	return builder
}
