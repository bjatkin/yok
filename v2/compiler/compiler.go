package compiler

import (
	"fmt"
	"strings"

	"github.com/bjatkin/yok/ast/shast"
	"github.com/bjatkin/yok/ast/yokast"
	"github.com/bjatkin/yok/errors"
)

// quoteIdentifiers is a shast.Visitor quotes all the identifiers
type quoteIdentifiers struct{}

// Visit implements the shast.Visitor interface
func (q *quoteIdentifiers) Visit(node shast.Node) shast.Visitor {
	if identifier, ok := node.(*shast.Identifier); ok {
		identifier.Quoted = true
	}

	return q
}

// Compiler can be used to compile code from a yok AST into an sh AST
type Compiler struct {
	errors []error
	source []byte
}

// New creates a new compiler
func New(source []byte) *Compiler {
	return &Compiler{
		source: source,
	}
}

// addError a new error to the complier
func (c *Compiler) addError(err error) {
	c.errors = append(c.errors, err)
}

// Compile creates an shast.Script from the given yokast.Script
func (c *Compiler) Compile(script *yokast.Script) (*shast.Script, error) {
	// fix the yokast before trying to complie to sh AST
	f := fixer{source: c.source}
	script.Statements = f.walkStmts(script.Statements)

	stmts := c.compileStatements(script.Statements)

	if len(c.errors) > 0 {
		return nil, errors.New("there were errors durring compilation")
	}

	return &shast.Script{
		Statements: stmts,
	}, nil
}

// compileStatements compiles a slice of yokast.Stmts into a list of shast.Stmts
func (c *Compiler) compileStatements(statements []yokast.Stmt) []shast.Stmt {
	stmts := []shast.Stmt{}
	for _, yokStmt := range statements {
		stmt := c.compileStmt(yokStmt)
		stmts = append(stmts, stmt)
	}

	return stmts
}

// complieNode converts a yokast.Stmt into it's equivilant shast.Stmt
func (c *Compiler) compileStmt(stmt yokast.Stmt) shast.Stmt {
	switch s := stmt.(type) {
	case *yokast.NewLine:
		return &shast.NewLine{}
	case *yokast.Comment:
		return &shast.Comment{
			Value: s.Token.Value(c.source),
		}
	case *yokast.Assign:
		identifier := s.Identifier.Name(c.source)
		identifier = strings.ToUpper(identifier)

		value := c.compileExpr(s.Value)
		_, ok := value.(*shast.InfixExpr)
		if ok {
			value = &shast.ArithmeticCommand{Expression: value}
		}

		return &shast.Assign{
			Identifier: identifier,
			Value:      value,
		}
	case *yokast.StmtExpr:
		expression := c.compileExpr(s.Expression)
		_, ok := expression.(*shast.InfixExpr)
		if ok {
			expression = &shast.ArithmeticCommand{Expression: expression}
		}

		return &shast.StmtExpr{Expression: expression}
	case *yokast.If:
		test := c.complieTestCommand(s.Test)
		stmts := c.compileStatements(s.Body.Statements)

		elseIfs := []shast.ElseIf{}
		for _, elseIf := range s.ElseIfs {
			test := c.complieTestCommand(elseIf.Test)
			stmts := c.compileStatements(elseIf.Body.Statements)
			elseIfs = append(elseIfs, shast.ElseIf{Test: test, Statements: stmts})
		}

		if s.ElseBody == nil {
			return &shast.If{
				Test:       test,
				Statements: stmts,
				ElseIfs:    elseIfs,
			}
		}

		elseStmts := c.compileStatements(s.ElseBody.Statements)
		return &shast.If{
			Test:           test,
			Statements:     stmts,
			ElseIfs:        elseIfs,
			ElseStatements: elseStmts,
		}
	default:
		panic(fmt.Sprintf("Unknown statement type %T", s))
	}
}

// complieExpr converts a yokast.Expr into it's equivilant shast.Expr
func (c *Compiler) compileExpr(expr yokast.Expr) shast.Expr {
	switch e := expr.(type) {
	case *yokast.String:
		value := e.Value(c.source)
		return &shast.String{Value: value}
	case *yokast.Atom:
		value := e.Token.Value(c.source)
		value = strings.TrimPrefix(value, ":")
		value = "\"" + value + "\""

		return &shast.String{Value: value}
	case *yokast.Identifier:
		value := e.Name(c.source)
		value = strings.ToUpper(value)

		return &shast.Identifier{Value: value}
	case *yokast.Call:
		return c.compileCall(e)
	case *yokast.InfixExpr:
		left := c.compileExpr(e.Left)
		right := c.compileExpr(e.Right)

		operator := e.Operator.Value(c.source)
		operator = convertOperator(operator)

		return &shast.InfixExpr{
			Left:     left,
			Operator: operator,
			Right:    right,
		}
	case *yokast.GroupExpr:
		expr := c.compileExpr(e.Expression)

		return &shast.GroupExpr{
			Expression: expr,
		}
	case *yokast.NestedCall:
		expr := c.compileExpr(e.Call)

		return &shast.CommandSub{
			Expression: expr,
		}
	default:
		panic(fmt.Sprintf("Unknown expression type %T", e))
	}
}

// complieTestCommand complies the given test into an shast.TestCommand
func (c *Compiler) complieTestCommand(test yokast.Expr) *shast.TestCommand {
	expr := c.compileExpr(test)

	v := &quoteIdentifiers{}
	shast.Walk(v, expr)

	return &shast.TestCommand{Expression: expr}
}

// convertOperator converts the yok operator to the equivalent 'sh' operator
func convertOperator(operator string) string {
	switch operator {
	case "==":
		return "="
	case "!=":
		return "!="
	case ">":
		return "-gt"
	case ">=":
		return "-ge"
	case "<":
		return "-lt"
	case "<=":
		return "-le"
	default:
		return operator
	}
}
