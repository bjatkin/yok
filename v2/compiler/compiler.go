package compiler

import (
	"fmt"
	"strings"

	"github.com/bjatkin/yok/ast/shast"
	"github.com/bjatkin/yok/ast/yokast"
	"github.com/bjatkin/yok/errors"
)

// Compiler can be used to compile code from a yok AST into an sh AST
type Compiler struct {
	// TODO: should there be an errors field here like the one we have in parse?
	source          []byte
	arithmeticDepth int
	inTest          bool
}

// New creates a new compiler
func New(source []byte) *Compiler {
	return &Compiler{
		source:          source,
		arithmeticDepth: 0,
		inTest:          false,
	}
}

// Compile creates an shast.Script from the given yokast.Script
func (c *Compiler) Compile(script *yokast.Script) (*shast.Script, error) {
	ret := &shast.Script{}
	for _, stmt := range script.Statements {
		node, err := c.compileNode(stmt.(yokast.Node))
		if err != nil {
			return nil, err
		}

		ret.Statements = append(ret.Statements, node.(shast.Stmt))
	}

	return ret, nil
}

// complieNode converts a yokast.Node into it's equivilant shast.Node
func (c *Compiler) compileNode(node yokast.Node) (shast.Node, error) {
	switch node := node.(type) {
	case *yokast.Script:
		script := &shast.Script{}
		for _, stmt := range node.Statements {
			stmt, err := c.compileNode(stmt)
			if err != nil {
				return nil, err
			}
			if stmt, ok := stmt.(shast.Stmt); ok {
				script.Statements = append(script.Statements, stmt)
			}
		}

		return script, nil
	case *yokast.NewLine:
		return &shast.NewLine{}, nil
	case *yokast.Comment:
		return &shast.Comment{
			Value: node.Value,
		}, nil
	case *yokast.Assign:
		identifier := node.Identifier.Value(c.source)
		identifier = strings.ToUpper(identifier)
		value, err := c.compileNode(node.Value)
		if err != nil {
			return nil, err
		}

		expr, ok := value.(shast.Expr)
		if !ok {
			return nil, errors.New("assignment value must be an expression")
		}

		return &shast.Assign{
			Identifier: identifier,
			Value:      expr,
		}, nil
	case *yokast.StmtExpr:
		value, err := c.compileNode(node.Expression)
		if err != nil {
			return nil, err
		}

		expr, ok := value.(shast.Expr)
		if !ok {
			return nil, errors.New("statement expression value must be an expression")
		}

		return &shast.StmtExpr{
			Expression: expr,
		}, nil
	case *yokast.String:
		return &shast.String{
			Value: node.Value,
		}, nil
	case *yokast.Atom:
		value := strings.TrimPrefix(node.Value, ":")
		value = "\"" + value + "\""
		return &shast.String{
			Value: value,
		}, nil
	case *yokast.Call:
		return c.compileCall(node)
	case *yokast.Identifier:
		value := node.Token.Value(c.source)
		value = strings.ToUpper(value)
		return &shast.Identifier{
			AsString: c.inTest,
			Value:    value,
		}, nil
	case *yokast.InfixExpr:
		c.arithmeticDepth++

		leftNode, err := c.compileNode(node.Left)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Arithmetic Commands failed to compile %v", err))
		}

		left, ok := leftNode.(shast.Expr)
		if !ok {
			return nil, errors.New("Arithmetic Commands must be expression")
		}

		rightNode, err := c.compileNode(node.Right)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Arithmetic Commands failed to compile %v", err))
		}

		right, ok := rightNode.(shast.Expr)
		if !ok {
			return nil, errors.New("Arithmetic Commands must be expressions")
		}

		c.arithmeticDepth--

		operator := node.Operator.Value(c.source)
		operator = convertOperator(operator)

		switch {
		case c.inTest && c.arithmeticDepth == 0:
			return &shast.TestCommand{
				Expression: &shast.InfixExpr{
					Left:     left,
					Operator: operator,
					Right:    right,
				},
			}, nil
		case c.inTest && c.arithmeticDepth == 1:
			fallthrough
		case !c.inTest && c.arithmeticDepth == 0:
			return &shast.ArithmeticCommand{
				Expression: &shast.InfixExpr{
					Left:     left,
					Operator: operator,
					Right:    right,
				},
			}, nil
		default:
			return &shast.InfixExpr{
				Left:     left,
				Operator: operator,
				Right:    right,
			}, nil
		}
	case *yokast.GroupExpr:
		c.arithmeticDepth++

		exprNode, err := c.compileNode(node.Expression)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("failed to compile expression %v", err))
		}

		expr, ok := exprNode.(shast.Expr)
		if !ok {
			return nil, errors.New("can only group expressions")
		}

		c.arithmeticDepth--

		// TODO: hmm I really feel like a lot of this could be generalized more...
		// I know that I'll have to change this to treat some operators as non-arithmetic
		// (e.g. ==, or >= ) as those should probably assume 'test' context by default.
		// I'm not sure if that will ever make it to the compiler though as I'll probably
		// weed out weird stuff like ( 10 == 20 ) + 30 in the yok ast validation
		switch {
		case c.inTest && c.arithmeticDepth == 0:
			return &shast.TestCommand{
				Expression: &shast.GroupExpr{
					Expression: expr,
				},
			}, nil
		case c.inTest && c.arithmeticDepth == 1:
			fallthrough
		case !c.inTest && c.arithmeticDepth == 0:
			return &shast.ArithmeticCommand{
				Expression: &shast.GroupExpr{
					Expression: expr,
				},
			}, nil
		default:
			return &shast.GroupExpr{
				Expression: expr,
			}, nil
		}
	case *yokast.If:
		test, err := c.complieTestCommand(node.Test)
		if err != nil {
			return nil, err
		}

		stmts, err := c.compileStatements(node.Body.Statements)
		if err != nil {
			return nil, err
		}

		elseIfs := []shast.ElseIf{}
		for _, elseIf := range node.ElseIfs {
			test, err := c.complieTestCommand(elseIf.Test)
			if err != nil {
				return nil, err
			}

			stmts, err := c.compileStatements(elseIf.Body.Statements)
			if err != nil {
				return nil, err
			}

			elseIfs = append(elseIfs, shast.ElseIf{Test: test, Statements: stmts})
		}

		if node.ElseBody == nil {
			return &shast.If{
				Test:       test,
				Statements: stmts,
				ElseIfs:    elseIfs,
			}, nil

		}

		elseStmts, err := c.compileStatements(node.ElseBody.Statements)
		if err != nil {
			return nil, err
		}

		return &shast.If{
			Test:           test,
			Statements:     stmts,
			ElseIfs:        elseIfs,
			ElseStatements: elseStmts,
		}, nil
	default:
		return nil, errors.New(fmt.Sprintf("Unknown node %T", node))
	}
}

// compileStatements compiles a slice of yokast.Stmts into a list of shast.Stmts
func (c *Compiler) compileStatements(statements []yokast.Stmt) ([]shast.Stmt, error) {
	stmts := []shast.Stmt{}
	for _, yokStmt := range statements {
		stmtNode, err := c.compileNode(yokStmt)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("failed to complie stmt in elif body %v", err))
		}

		stmt, ok := stmtNode.(shast.Stmt)
		if !ok {
			return nil, errors.New("line in body was not a stmt")
		}

		stmts = append(stmts, stmt)
	}

	return stmts, nil
}

// complieTestCommand complies the given test into an shast.TestCommand
func (c *Compiler) complieTestCommand(test yokast.Expr) (*shast.TestCommand, error) {
	c.inTest = true

	testNode, err := c.compileNode(test)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("failed to compile test for if stmt %v", err))
	}

	c.inTest = false

	testCommand, ok := testNode.(*shast.TestCommand)
	if !ok {
		return nil, errors.New("test must be a valid test command")
	}

	return testCommand, nil
}

// compileCall compiles a yokast.Call into it's equivilant shast.Node
func (c *Compiler) compileCall(call *yokast.Call) (shast.Expr, error) {
	command := call.Identifier.Token.Value(c.source)
	args := []shast.Expr{}
	for _, arg := range call.Arguments {
		value, err := c.compileNode(arg)
		if err != nil {
			return nil, err
		}

		expr, ok := value.(shast.Expr)
		if !ok {
			return nil, errors.New("argument must be an expression")
		}
		args = append(args, expr)
	}

	switch command {
	case "print":
		return &shast.Exec{
			Command:   "echo",
			Arguments: args,
			Redirects: []shast.Redirect{{RightFd: "2"}},
		}, nil
	case "len":
		return compileLen(args)
	case "replace":
		return compileReplace(args, false)
	case "replace_all":
		return compileReplace(args, true)
	case "remove_prefix":
		return compileRemoveFix(args, true)
	case "remove_suffix":
		return compileRemoveFix(args, false)
	default:
		return &shast.Exec{
			Command:   command,
			Arguments: args,
		}, nil
	}
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

// compileLen takes in a list of arguments and complies a call to the len() builtin
func compileLen(args []shast.Expr) (shast.Expr, error) {
	if len(args) != 1 {
		return nil, errors.New("len() takes only a single argument")
	}

	identifier, ok := args[0].(*shast.Identifier)
	if !ok {
		return nil, errors.New("len() only supports identifiers")
	}

	// TODO: fix this to support expressions other than identifiers
	return &shast.ParamaterExpansion{
		Expression: &shast.ParameterLength{
			Paramater: identifier,
		},
	}, nil
}

// compileReplace takes in a list of arguments and compiles a call to the replace() builtin
func compileReplace(args []shast.Expr, replaceAll bool) (shast.Expr, error) {
	if len(args) != 3 {
		return nil, errors.New("replace() takes 3 arguments")
	}

	identifier, ok := args[0].(*shast.Identifier)
	if !ok {
		return nil, errors.New("replace() first argument must be an identifier")
	}

	find, ok := args[1].(*shast.String)
	if !ok {
		return nil, errors.New("replace() second argument must be a string")
	}

	replace, ok := args[2].(*shast.String)
	if !ok {
		return nil, errors.New("replace() third arguments must be a string")
	}

	return &shast.ParamaterExpansion{
		Expression: &shast.ParamaterReplace{
			ReplaceAll: replaceAll,
			Paramater:  identifier,
			Find:       find,
			Replace:    replace,
		},
	}, nil

}

// compileRemoveFix takes a list of arguments and compiles a call to the remove_suffix() builtin
func compileRemoveFix(args []shast.Expr, prefix bool) (shast.Expr, error) {
	if len(args) != 2 {
		return nil, errors.New("remove_suffix() takes 2 arguments")
	}

	identifier, ok := args[0].(*shast.Identifier)
	if !ok {
		return nil, errors.New("remove_suffix() first argument must be an identifier")
	}

	remove, ok := args[1].(*shast.String)
	if !ok {
		return nil, errors.New("remove_suffix() second argument must be string")
	}

	return &shast.ParamaterExpansion{
		Expression: &shast.ParamaterRemoveFix{
			RemovePrefix: prefix,
			Paramater:    identifier,
			Remove:       remove,
		},
	}, nil
}
