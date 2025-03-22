package compiler

import (
	"fmt"

	"github.com/bjatkin/yok/ast/yokast"
	"github.com/bjatkin/yok/token"
)

// fixer fixes up the yok AST so it's closer to the sh AST that will be generated.
// It does de-sugaring and simplifies complex code that can't be represented directly in sh
type fixer struct {
	source             []byte
	errors             []error
	internalIdentifier int
}

func (f *fixer) walkStmts(statements []yokast.Stmt) []yokast.Stmt {
	stmts := []yokast.Stmt{}
	for _, yokStmt := range statements {
		stmt := f.fixStmt(yokStmt)
		stmts = append(stmts, stmt...)
	}

	return stmts
}

func (f *fixer) fixStmt(stmt yokast.Stmt) []yokast.Stmt {
	switch s := stmt.(type) {
	case *yokast.Assign:
		stmts, expr := f.fixExpr(s.Value, 0)
		s.Value = expr
		return append(stmts, s)
	case *yokast.StmtExpr:
		stmts, expr := f.fixExpr(s.Expression, 0)
		s.Expression = expr
		return append(stmts, s)
	default:
		// TODO: this is not going to work, we actually need to walk
		// the whole tree here
		return []yokast.Stmt{s}
	}
}

func (f *fixer) fixExpr(expr yokast.Expr, depth int) ([]yokast.Stmt, yokast.Expr) {
	switch e := expr.(type) {
	case *yokast.Call:
		callName := e.Identifier.Token.Value(f.source)
		switch callName {
		case "len":
			if len(e.Arguments) == 0 {
				return nil, e
			}

			prefix, target := f.simplifyToIdent(e.Arguments[0], depth+1)
			e.Arguments[0] = target

			return prefix, e
		case "replace":
			if len(e.Arguments) != 3 {
				return nil, e
			}

			targetStmts, target := f.simplifyToIdent(e.Arguments[0], depth+1)
			e.Arguments[0] = target
			findStmts, find := f.simplifyToStringOrCommand(e.Arguments[1], depth+1)
			e.Arguments[1] = find
			replaceStmts, replace := f.simplifyToStringOrCommand(e.Arguments[2], depth+1)
			e.Arguments[2] = replace

			prefix := append(targetStmts, findStmts...)
			prefix = append(prefix, replaceStmts...)
			return prefix, e
		case "replace_all":
			if len(e.Arguments) != 3 {
				return nil, e
			}

			targetStmts, target := f.simplifyToIdent(e.Arguments[0], depth+1)
			e.Arguments[0] = target
			findStmts, find := f.simplifyToStringOrCommand(e.Arguments[1], depth+1)
			e.Arguments[1] = find
			replaceStmts, replace := f.simplifyToStringOrCommand(e.Arguments[2], depth+1)
			e.Arguments[2] = replace

			prefix := append(targetStmts, findStmts...)
			prefix = append(prefix, replaceStmts...)
			return prefix, e
		case "remove_suffix":
			if len(e.Arguments) != 2 {
				return nil, e
			}

			targetStmts, target := f.simplifyToIdent(e.Arguments[0], depth+1)
			e.Arguments[0] = target
			removeStmts, remove := f.simplifyToStringOrCommand(e.Arguments[1], depth+1)
			e.Arguments[1] = remove

			prefix := append(targetStmts, removeStmts...)
			return prefix, e
		case "remove_prefix":
			if len(e.Arguments) != 2 {
				return nil, e
			}

			targetStmts, target := f.simplifyToIdent(e.Arguments[0], depth+1)
			e.Arguments[0] = target
			removeStmts, remove := f.simplifyToStringOrCommand(e.Arguments[1], depth+1)
			e.Arguments[1] = remove

			prefix := append(targetStmts, removeStmts...)
			return prefix, e
		default:
			prefix := []yokast.Stmt{}
			for i, arg := range e.Arguments {
				stmts, a := f.fixExpr(arg, depth+1)
				e.Arguments[i] = a
				prefix = append(prefix, stmts...)
			}

			if depth == 0 {
				return prefix, e
			}

			return prefix, &yokast.NestedCall{Depth: depth, Call: e}
		}
	default:
		return nil, e
	}
}

func (f *fixer) nextTmpIdentifier() string {
	f.internalIdentifier += 1
	ident := fmt.Sprintf("_TMP%d", f.internalIdentifier)
	return ident
}

// TODO: we should consider looking for asignment expressiosn that already match the literal value so we don't get
// duplicate identifiers that map to the same value
func (f *fixer) simplifyToIdent(expr yokast.Expr, depth int) ([]yokast.Stmt, *yokast.Identifier) {
	if ident, ok := expr.(*yokast.Identifier); ok {
		return nil, ident
	}

	prefix, fixedExpr := f.fixExpr(expr, depth)
	tok := token.Token{Type: token.Identifier}
	tmpName := f.nextTmpIdentifier()
	ident := yokast.NewInternalIdentifier(tmpName, tok)
	return append(
		prefix,
		&yokast.Assign{Identifier: ident, Value: fixedExpr},
	), ident
}

// TODO: we should consider looking for asignment expressiosn that already match the literal value so we don't get
// duplicate identifiers that map to the same value
func (f *fixer) simplifyToStringOrCommand(expr yokast.Expr, depth int) ([]yokast.Stmt, yokast.Expr) {
	if lit, ok := expr.(*yokast.String); ok {
		return nil, lit
	}

	if call, ok := expr.(*yokast.Call); ok {
		return nil, &yokast.NestedCall{Depth: depth, Call: call}
	}

	prefix := []yokast.Stmt{}
	ident, ok := expr.(*yokast.Identifier)
	if !ok {
		tok := token.Token{Type: token.Identifier}
		tmpName := f.nextTmpIdentifier()
		ident = yokast.NewInternalIdentifier(tmpName, tok)

		fixedPrefix, fixedExpr := f.fixExpr(expr, depth)
		prefix = append(
			fixedPrefix,
			&yokast.Assign{Identifier: ident, Value: fixedExpr},
		)
	}

	call := &yokast.NestedCall{
		Depth: depth,
		Call: &yokast.Call{
			Identifier: yokast.NewInternalIdentifier("echo", token.Token{Type: token.Identifier}),
			Arguments:  []yokast.Expr{yokast.NewInternalString("-n", token.Token{Type: token.StringLiteral}), ident},
		},
	}

	return prefix, call
}
