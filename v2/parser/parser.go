package parser

import (
	"github.com/bjatkin/yok/ast/yokast"
	"github.com/bjatkin/yok/errors"
	"github.com/bjatkin/yok/token"
)

// precedence is the precedence of yok operations
type precedence int

const (
	Unknown = precedence(iota)
	Lowest
	Equals
	LessOrGreater
	Sum
	Product
	Prefix
	Call
)

// precedenceMap maps yok tokens to their precedence
var precedenceMap = map[token.Type]precedence{
	token.EqualEqual:   Equals,
	token.NotEqual:     Equals,
	token.LessThan:     LessOrGreater,
	token.LessEqual:    LessOrGreater,
	token.GreaterThan:  LessOrGreater,
	token.GreaterEqual: LessOrGreater,
	token.Plus:         Sum,
	token.Minus:        Sum,
	token.Divide:       Product,
	token.Multiply:     Product,
	token.OpenParen:    Call,
}

// these are the prat parser function types
type (
	prefixParseFn func() yokast.Expr
	infixParseFn  func(yokast.Expr) yokast.Expr
)

// Parser is a parser for yok source code
type Parser struct {
	lexer  lexer
	Errors []error

	prefixParseFn map[token.Type]prefixParseFn
	infixParseFn  map[token.Type]infixParseFn
}

// New creates a new parser for the given yok source code
func New(source []byte) *Parser {
	if source[len(source)-1] != '\n' {
		// TODO: this is silly, we really don't need to support windows line endings for a language
		// that transpiles to POSIX shell. We only need to do this because I'm currently developing
		// primarily on windows. At some point I need to move over to
		// Linux and get rid of all this silly-ness
		source = append(source, '\r', '\n')
	}

	p := &Parser{
		lexer: newLexer(source),
	}

	p.prefixParseFn = map[token.Type]prefixParseFn{
		token.StringLiteral: p.parseStringLiteral,
		token.Atom:          p.parseAtom,
		token.Identifier:    p.parseIdentifier,
		token.Minus:         p.parsePrefixExpr,
		token.OpenParen:     p.parseGroupExpr,
	}

	p.infixParseFn = map[token.Type]infixParseFn{
		token.OpenParen:    p.parseCall,
		token.Plus:         p.parseInfix,
		token.Minus:        p.parseInfix,
		token.Multiply:     p.parseInfix,
		token.Divide:       p.parseInfix,
		token.Mod:          p.parseInfix,
		token.GreaterThan:  p.parseInfix,
		token.GreaterEqual: p.parseInfix,
		token.LessThan:     p.parseInfix,
		token.LessEqual:    p.parseInfix,
		token.EqualEqual:   p.parseInfix,
		token.NotEqual:     p.parseInfix,
	}

	return p
}

// Parse parses the source code that was given to the parser.
// If an error is returned the Parser.Errors field will contain all the encountered parsing errors
func (p *Parser) Parse() (*yokast.Script, error) {
	script := &yokast.Script{}

	for {
		t := p.peek()
		if t.Type == token.EOF {
			break
		}

		stmt := p.parseStmt()
		if stmt == nil {
			// we should continue to consume tokens here to get us back on track
			_ = p.take()
			continue
		}

		script.Statements = append(script.Statements, stmt)
	}

	if len(p.Errors) > 0 {
		return script, errors.New("there were errors while parsing the script")
	}

	return script, nil
}

// peek calls peek on the lexer
func (p *Parser) peek() token.Token {
	return p.lexer.peek()
}

// take calls take on the lexer
func (p *Parser) take() token.Token {
	return p.lexer.take()
}

// getValue gets the string value of the token from the lexer source
func (p *Parser) getValue(t token.Token) string {
	return t.Value(p.lexer.source)
}

func (p *Parser) parseStmt() yokast.Stmt {
	switch p.peek().Type {
	case token.Comment:
		// we treat comments as statements because they need to show up in the generated code
		comment := p.getValue(p.take())

		if p.peek().Type != token.NewLine {
			p.Errors = append(p.Errors, errors.New("comment did not end with a new line: "+p.getValue(p.peek())))
			return nil
		}

		_ = p.take()
		return &yokast.Comment{
			Value: comment,
		}
	case token.NewLine:
		// we treat empty new lines as statements because they need to show up in the generated code
		_ = p.take()
		return &yokast.NewLine{}
	case token.LetKeyword:
		return p.parseAssignStmt()
	case token.IfKeyword:
		return p.parseIfStmt()
	default:
		expr := p.parseExpr(Lowest)

		// All statements must end with a new line
		if p.peek().Type != token.NewLine {
			p.Errors = append(p.Errors, errors.New("statement did not end with a new line: "+p.getValue(p.peek())))
			return nil
		}

		_ = p.take()
		return &yokast.StmtExpr{
			Expression: expr,
		}
	}
}

// parseAssignStmt parses a yok let statement
// Examples:
//
//	let a = 10
//	let b = myFunc()
//	let c = 10 * 20
func (p *Parser) parseAssignStmt() *yokast.Assign {
	// discard the 'let' token
	_ = p.take()

	ident := p.take()

	if p.peek().Type != token.Assign {
		p.Errors = append(p.Errors, errors.New("let statement must include an '=' after the identifier"))
		return nil
	}
	// discard the '=' token
	_ = p.take()

	value := p.parseExpr(Lowest)

	if p.peek().Type != token.NewLine {
		p.Errors = append(p.Errors, errors.New("let statement must end with a new line"))
		return nil
	}
	// discard the new line
	_ = p.take()

	return &yokast.Assign{
		Identifier: ident,
		Value:      value,
	}
}

// parseIfStmt parses a yok if statement
// Examples:
//
// if a > 10 { ... }
func (p *Parser) parseIfStmt() *yokast.If {
	// discard the 'if' token
	_ = p.take()

	test := p.parseExpr(Lowest)
	body := p.parseBlock()

	// check for the 'else' token
	var elseBody *yokast.Block
	if p.peek().Type == token.ElseKeyword {
		_ = p.take()

		elseBody = p.parseBlock()
	}

	// ensure the final token is a new line or we have some random syntax to deal with...
	if p.peek().Type != token.NewLine {
		p.Errors = append(p.Errors, errors.New("if body must end with '}' on it's own line"))
		return nil
	}

	// take the final '\n'
	_ = p.take()

	return &yokast.If{
		Test:     test,
		Body:     body,
		ElseBody: elseBody,
	}
}

// parseBlock parses a yok body
func (p *Parser) parseBlock() *yokast.Block {
	if p.peek().Type != token.OpenBrace {
		p.Errors = append(p.Errors, errors.New("must start with a '{'"))
		return nil
	}

	// discard the '{' token
	_ = p.take()

	// the opening '{' may or may not be followed by a new line
	if p.peek().Type == token.NewLine {
		p.take()
	}

	stmts := []yokast.Stmt{}
	for {
		if p.peek().Type == token.EOF {
			p.Errors = append(p.Errors, errors.New("the block was not closed."))
			return nil
		}

		if p.peek().Type == token.CloseBrace {
			break
		}

		stmt := p.parseStmt()
		if stmt == nil {
			continue
		}

		stmts = append(stmts, stmt)
	}

	// discard the final '}' token
	_ = p.take()

	return &yokast.Block{
		Statements: stmts,
	}
}

// parseExpr parses a yok expression
func (p *Parser) parseExpr(leftPrecedence precedence) yokast.Expr {
	prefix, ok := p.prefixParseFn[p.lexer.peek().Type]
	if !ok {
		p.Errors = append(p.Errors, errors.New("missing prefix function for token: "+p.getValue(p.peek())))
		return nil
	}

	left := prefix()

	for p.peek().Type != token.EOF {
		if p.peek().Type == token.NewLine {
			break
		}

		if leftPrecedence > tokenPrecedence(p.peek()) {
			break
		}

		infix, ok := p.infixParseFn[p.peek().Type]
		if !ok {
			return left
		}

		// // take the infix operator since that's implied by the infix parse function
		// _ = p.lexer.take()

		left = infix(left)
	}

	return left
}

// tokenPrecedence returns the precedence of the given token
func tokenPrecedence(t token.Token) precedence {
	if p, ok := precedenceMap[t.Type]; ok {
		return p
	}

	return Lowest
}

// parseStringLiteral parses a string literal in yok
//
// Example:
//
//	"hello world"
func (p *Parser) parseStringLiteral() yokast.Expr {
	if p.peek().Type != token.StringLiteral {
		panic("token is not a string literal: " + p.getValue(p.peek()))
	}

	return &yokast.String{
		Value: p.getValue(p.take()),
	}
}

// parseAtom parses an atom in yok
//
// Example:
//
//	:hello
//	:world
func (p *Parser) parseAtom() yokast.Expr {
	if p.peek().Type != token.Atom {
		panic("token is not an atom: " + p.getValue(p.peek()))
	}

	return &yokast.Atom{
		Value: p.getValue(p.take()),
	}
}

// parsePrefixExpr parses prefix yok expressions
func (p *Parser) parsePrefixExpr() yokast.Expr {
	return &yokast.PrefixExpr{
		Token: p.take(),
		Expr:  p.parseExpr(Prefix),
	}
}

// parseGroupExpr parses grouped yok expressions
func (p *Parser) parseGroupExpr() yokast.Expr {
	// take the initial '('
	_ = p.take()

	expr := p.parseExpr(Lowest)

	if p.peek().Type != token.CloseParen {
		// TODO: add an error here?
		return nil
	}

	// take the final ')'
	_ = p.take()

	return &yokast.GroupExpr{
		Expression: expr,
	}
}

// parseIdentifier parses the next token as a yok identifier
func (p *Parser) parseIdentifier() yokast.Expr {
	t := p.take()
	return &yokast.Identifier{
		Token: t,
	}
}

// parseCall parses a function call in yok
//
// Example:
//
// print("hello world")
func (p *Parser) parseCall(ident yokast.Expr) yokast.Expr {
	// need to take the initial '('
	_ = p.take()

	identifier, ok := ident.(*yokast.Identifier)
	if !ok {
		p.Errors = append(p.Errors, errors.New("left side of the call was not an identifier"))
		return nil
	}

	args := []yokast.Expr{}
	for {
		// just skip new lines when parsing arguments
		if p.peek().Type == token.NewLine {
			continue
		}

		expr := p.parseExpr(Lowest)
		if expr == nil {
			// TODO, should this be an error?
			return nil
		}

		args = append(args, expr)

		if p.peek().Type == token.Comma {
			_ = p.take()
			continue
		}

		if p.peek().Type != token.CloseParen {
			p.Errors = append(p.Errors, errors.New("unclosed argument list: "+p.getValue(p.peek())))
			return nil
		}

		// take the closing paren ')'
		_ = p.take()
		break
	}

	return &yokast.Call{
		Identifier: identifier,
		Arguments:  args,
	}
}

func (p *Parser) parseInfix(left yokast.Expr) yokast.Expr {
	operator := p.take()
	precedence := tokenPrecedence(operator)
	right := p.parseExpr(precedence)

	return &yokast.InfixExpr{
		Left:     left,
		Operator: operator,
		Right:    right,
	}
}
