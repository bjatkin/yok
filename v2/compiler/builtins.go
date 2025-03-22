package compiler

import (
	"fmt"

	"github.com/bjatkin/yok/ast/shast"
	"github.com/bjatkin/yok/ast/yokast"
	"github.com/bjatkin/yok/errors"
)

// compileCall compiles a yokast.Call into it's equivilant shast.Node
func (c *Compiler) compileCall(call *yokast.Call) shast.Expr {
	command := call.Identifier.Name(c.source)
	args := []shast.Expr{}
	for _, arg := range call.Arguments {
		expr := c.compileExpr(arg)
		args = append(args, expr)
	}

	switch command {
	case "print":
		return compilePrint(args)
	case "len":
		len, err := compileLen(args)
		if err != nil {
			c.addError(err)
			return nil
		}

		return &shast.ParamaterExpansion{Expression: len}
	case "remove_prefix":
		remove, err := compileRemoveFix(args)
		if err != nil {
			c.addError(err)
			return nil
		}
		remove.RemovePrefix = true

		return &shast.ParamaterExpansion{Expression: remove}
	case "remove_suffix":
		remove, err := compileRemoveFix(args)
		if err != nil {
			c.addError(err)
			return nil
		}
		remove.RemovePrefix = false

		return &shast.ParamaterExpansion{Expression: remove}
	default:
		return &shast.Exec{
			Command:   command,
			Arguments: args,
		}
	}
}

// compilePrint takes a list of arguments and complies a call to the echo command
func compilePrint(args []shast.Expr) *shast.Exec {
	return &shast.Exec{
		Command:   "echo",
		Arguments: args,
		Redirects: []shast.Redirect{{RightFd: "2"}},
	}
}

// compileLen takes in a list of arguments and complies a *shast.ParameterLength
func compileLen(args []shast.Expr) (*shast.ParameterLength, error) {
	if len(args) != 1 {
		return nil, errors.New("len() takes only a single argument")
	}

	identifier, ok := args[0].(*shast.Identifier)
	if !ok {
		return nil, errors.New(fmt.Sprintf("len() only supports identifiers, but got (%T)%#v", args[0], args[0]))
	}

	// TODO: fix this to support expressions other than identifiers
	return &shast.ParameterLength{
		Paramater: identifier,
	}, nil
}

// compileRemoveFix takes a list of arguments and compiles a shast.ParamaterRemoveFix
func compileRemoveFix(args []shast.Expr) (*shast.ParamaterRemoveFix, error) {
	if len(args) != 2 {
		return nil, errors.New("remove_suffix() takes 2 arguments")
	}

	identifier, ok := args[0].(*shast.Identifier)
	if !ok {
		return nil, errors.New("remove_suffix() first argument must be an identifier")
	}

	remove, ok := isStringOrCommand(args[1])
	if !ok {
		return nil, errors.New("remove_suffix() second argument must be string")
	}

	return &shast.ParamaterRemoveFix{
		Paramater: identifier,
		Remove:    remove,
	}, nil
}

func isStringOrCommand(expr shast.Expr) (shast.Expr, bool) {
	strExpr, ok := expr.(*shast.String)
	if ok {
		return strExpr, ok
	}
	cmdExpr, ok := expr.(*shast.CommandSub)
	if ok {
		return cmdExpr, ok
	}

	return expr, false
}
