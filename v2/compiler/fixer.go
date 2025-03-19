package compiler

import (
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
	case *yokast.StmtExpr:
		stmts, expr := f.fixExpr(s.Expression, 0)
		s.Expression = expr
		return append(stmts, s)
	default:
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
			stmts, arg := f.simplifyToIdent(e.Arguments[0], depth+1)
			e.Arguments[0] = arg
			return stmts, e
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

func (f *fixer) simplifyToIdent(expr yokast.Expr, depth int) ([]yokast.Stmt, *yokast.Identifier) {
	if ident, ok := expr.(*yokast.Identifier); ok {
		return nil, ident
	}

	prefix, fixedExpr := f.fixExpr(expr, depth)
	tok := token.Token{Type: token.IRIdentifier, Pos: token.Pos(f.nextInternalIdentifier())}
	ident := &yokast.Identifier{Token: tok}
	return append(
		prefix,
		&yokast.Assign{Identifier: ident.Token, Value: fixedExpr},
	), ident
}

func (f *fixer) nextInternalIdentifier() int {
	f.internalIdentifier += 1
	return f.internalIdentifier
}
