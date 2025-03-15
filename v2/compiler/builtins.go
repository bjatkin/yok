package compiler

import (
	"github.com/bjatkin/yok/ast/shast"
	"github.com/bjatkin/yok/ast/yokast"
	"github.com/bjatkin/yok/errors"
)

// compileCall compiles a yokast.Call into it's equivilant shast.Node
func (c *Compiler) compileCall(call *yokast.Call) shast.Expr {
	command := call.Identifier.Token.Value(c.source)
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
		}

		return &shast.ParamaterExpansion{Expression: len}
	case "replace":
		replace, err := compileReplace(args)
		if err != nil {
			c.addError(err)
		}
		replace.ReplaceAll = false

		return &shast.ParamaterExpansion{Expression: replace}
	case "replace_all":
		replace, err := compileReplace(args)
		if err != nil {
			c.addError(err)
		}
		replace.ReplaceAll = true

		return &shast.ParamaterExpansion{Expression: replace}
	case "remove_prefix":
		remove, err := compileRemoveFix(args)
		if err != nil {
			c.addError(err)
		}
		remove.RemovePrefix = true

		return &shast.ParamaterExpansion{Expression: remove}
	case "remove_suffix":
		remove, err := compileRemoveFix(args)
		if err != nil {
			c.addError(err)
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
		return nil, errors.New("len() only supports identifiers")
	}

	// TODO: fix this to support expressions other than identifiers
	return &shast.ParameterLength{
		Paramater: identifier,
	}, nil
}

// compileReplace takes in a list of arguments and compiles a *shast.ParamaterReplace
func compileReplace(args []shast.Expr) (*shast.ParamaterReplace, error) {
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

	return &shast.ParamaterReplace{
		Paramater: identifier,
		Find:      find,
		Replace:   replace,
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

	remove, ok := args[1].(*shast.String)
	if !ok {
		return nil, errors.New("remove_suffix() second argument must be string")
	}

	return &shast.ParamaterRemoveFix{
		Paramater: identifier,
		Remove:    remove,
	}, nil
}
