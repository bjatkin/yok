package parser

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/bjatkin/yok/v2/ast"
	"github.com/bjatkin/yok/v2/ekit"
	"github.com/bjatkin/yok/v2/token"
)

type Parser struct {
	filePath string
	src      string
}

func New() *Parser {
	return &Parser{}
}

func (p *Parser) Parse(filePath string, src []byte) (ast.Program, error) {
	p.filePath = filePath
	p.src = string(src)

	tokens := newStream(newLexer().lex(p.src))

	errs := ekit.NewErrList(8)
	program := ast.Program{}
	for !tokens.isEmpty() {
		stmt, err := p.stmt(tokens)
		if err != nil {
			errs.AddErr(err)

			// panic mode! eat tokens until we're safe again
			p.panicTake(tokens)
		}

		program.Stmts = append(program.Stmts, stmt)
	}

	if errs.HasErrors() {
		return ast.Program{}, errs
	}

	return program, nil
}

func (p *Parser) panicTake(tokens *stream[token.Token]) {
	for !tokens.isEmpty() {
		// look for safe tokens to exit panic mode
		switch tokens.peek().Type {
		case token.If:
			return
		case token.Return:
			return
		case token.Break:
			return
		default:
			tokens.take()
		}
	}
}

func (p *Parser) newErr(start token.Token, title ekit.Title, message string) *ekit.Err {
	return &ekit.Err{
		File:    p.filePath,
		Src:     p.src,
		Start:   start,
		Title:   title,
		Message: message,
	}
}

func (p *Parser) wrapErr(parent error, title ekit.Title, message string) *ekit.Err {
	switch err := parent.(type) {
	case *ekit.Err:
		// if there's a full error from deeper in the call stack, that error should take precidence
		return err
	case *ekit.Condition:
		return p.newErr(err.Start, title, message).AddCondition(err.Conditions...)
	default:
		// TODO: this could get dirty... I wonder if there's a better way to handle this case
		return p.newErr(token.Token{}, title, message).AddCondition(err.Error())

	}
}

func (p *Parser) stmt(tokens *stream[token.Token]) (ast.Stmt, error) {
	switch tokens.peek().Type {
	case token.OpenBrace:
		return p.block(tokens)
	case token.If:
		tokens.take()

		expr, err := p.expr(tokens)
		if err != nil {
			return nil, p.wrapErr(
				err,
				ekit.TitleInvalidStatement,
				"the 'if' keyword must be followed by a valid expression",
			)
		}

		block, err := p.block(tokens)
		if err != nil {
			return nil, p.wrapErr(
				err,
				ekit.TitleInvalidBlock,
				"an 'if' statement must be followed by a valid block",
			)
		}

		if tokens.peek().Type == token.Else {
			tokens.take()
			elseBlock, err := p.block(tokens)
			if err != nil {
				return nil, p.wrapErr(
					err,
					ekit.TitleInvalidBlock,
					"an 'else' statement must be followed by a valid block",
				)
			}

			return ast.If{
				Check:    expr,
				Body:     block,
				ElseBody: elseBlock,
			}, nil
		}

		return &ast.If{
			Check: expr,
			Body:  block,
		}, nil
	case token.Var:
		// take the 'var' keyword
		tokens.take()

		name := tokens.take()
		if name.Type != token.Identifyer {
			return nil, p.newErr(
				name,
				ekit.TitleInvalidStatement,
				"a 'var' keyword must be followed by a valid identifyer",
			).AddCondition(
				"The 'var' keyword declares a new variable and must be followed by an identifyer which provides the variables name",
				"Valid identifyers must start with a letter (a-z or A-Z) and can only contain letters (a-z or A-Z) numbers (0-9) or underscores ('_')",
			)
		}

		// TODO: can we resolve types here? should we?
		var value ast.Expr
		switch tokens.take().Type {
		case token.Assign:
			expr, err := p.expr(tokens)
			if err != nil {
				return nil, p.wrapErr(
					err,
					ekit.TitleInvalidExpression,
					"this is not a valid expression",
				).AddCondition("only valid expressions can be assigned to variables")
			}

			value = expr
		case token.Int:
			value = &ast.IntLiteral{}
		case token.Bool:
			value = &ast.BoolLiteral{}
		case token.String:
			value = &ast.StringLiteral{}
		case token.Error:
			value = &ast.ErrorLiteral{}
		case token.Path:
			value = &ast.PathLiteral{Value: "./"}
		default:
			return nil, p.newErr(
				tokens.peek(),
				ekit.TitleInvalidStatement,
				"this token is not valid as part of a variable declaration",
			).AddCondition(
				"Variable declrations must be followed by an assignemnt ( = {value} ) or a TYPE name",
				fmt.Sprintf("Instead I found '%s'", tokens.peek().Lexeme),
			)
		}

		return &ast.Decl{
			Name:  name,
			Value: value,
		}, nil
	case token.Identifyer:
		// TODO: this can be an expression... should we check that here?
		name := tokens.take()
		if got := tokens.take(); got.Type != token.Assign {
			// TODO: handle error
			return nil, errors.New("TODO: 20394872 " + fmt.Sprintf("%#v %#v", name, got))
		}

		value, err := p.expr(tokens)
		if err != nil {
			return nil, p.wrapErr(
				err,
				ekit.TitleInvalidStatement,
				fmt.Sprintf("this is not a valid value to assign to a '%s'", name.Lexeme),
			)
		}

		return &ast.Assign{
			Name:  name,
			Value: value,
		}, nil
	case token.Return:
		// take the 'return' token
		tokens.take()
		if tokens.peek().Type == token.NewLine {
			return &ast.Return{}, nil
		}

		// TODO: check to make sure we're inside a function context,
		// returning from the root is not valid and show show up as failed here
		ret, err := p.expr(tokens)
		if err != nil {
			return nil, p.wrapErr(
				err,
				ekit.TitleInvalidControllFlow,
				"this is not a valid return value",
			)
		}
		return &ast.Return{Value: ret}, nil
	case token.Break:
		// take the 'break' token
		tokens.take()
		if tokens.peek().Type == token.NewLine {
			return &ast.Break{}, nil
		}

		// TODO: check to make sure we're inside a loop/ switch context.
		// breaking out of a function or loop is not valid and shoulw show up as failed here
		ret, err := p.expr(tokens)
		if err != nil {
			return nil, p.wrapErr(
				err,
				ekit.TitleInvalidControllFlow,
				"this is not a valid break value",
			)
		}

		return &ast.Break{Value: ret}, nil
	case token.NewLine:
		// take the '\n' token
		tokens.take()
		return ast.NewLine{}, nil
	case token.Comment:
		comment := tokens.take()
		return ast.Comment{Value: strings.TrimSpace(comment.Lexeme)}, nil
	default:
		return nil, p.newErr(
			tokens.peek(),
			ekit.TitleUnknownToken,
			fmt.Sprintf("'%s' is not a valid way to start a statement", tokens.peek().Lexeme),
		)
	}
}

func (p *Parser) block(tokens *stream[token.Token]) (*ast.Block, error) {
	if got := tokens.take(); got.Type != token.OpenBrace {
		return nil, ekit.NewCondition(got,
			"I expected to find a block here and blocks must begin wth a '{'.",
			fmt.Sprintf("Instead, I found '%s'", got.Lexeme),
		)
	}

	block := &ast.Block{}
	for {
		if tokens.isEmpty() {
			return nil, ekit.NewCondition(tokens.prev(),
				"I expected to find a closing '}' for the block",
				"Instead, I found the end of the file",
			)
		}
		if tokens.peek().Type == token.CloseBrace {
			tokens.take()
			break
		}

		stmt, err := p.stmt(tokens)
		if err != nil {
			// no context to add here
			return nil, err
		}

		block.Stmts = append(block.Stmts, stmt)
	}

	return block, nil
}

func (p *Parser) expr(tokens *stream[token.Token]) (ast.Expr, error) {
	// TODO: do we actually need this expr function?
	// right now it just makes expanding expression parsing easy
	return p.equality(tokens)
}

func (p *Parser) equality(tokens *stream[token.Token]) (ast.Expr, error) {
	left, err := p.comparison(tokens)
	if err != nil {
		return nil, err
	}

	next := tokens.peek().Type
	if next != token.Equal && next != token.NotEqual {
		return left, nil
	}

	op := tokens.take()
	right, err := p.expr(tokens)
	if err != nil {
		return nil, ekit.NewCondition(tokens.peek(),
			fmt.Sprintf("I expected an expression on the right side of the '%s' symbol. Instead I found '%s'", op.Lexeme, tokens.peek().Lexeme),
		)
	}

	return &ast.Binary{
		Left:  left,
		Op:    op,
		Right: right,
	}, nil
}

func (p *Parser) comparison(tokens *stream[token.Token]) (ast.Expr, error) {
	left, err := p.addSubExpr(tokens)
	if err != nil {
		return nil, err
	}

	next := tokens.peek().Type
	if next != token.Greater && next != token.GreaterEqual && next != token.Less && next != token.LessEqual {
		return left, nil
	}

	op := tokens.take()
	right, err := p.expr(tokens)
	if err != nil {
		return nil, ekit.NewCondition(tokens.peek(),
			fmt.Sprintf("I expected an expression on the right side of the '%s' symbol. Instead I found '%s'", op.Lexeme, tokens.peek().Lexeme),
		)
	}

	return &ast.Binary{
		Left:  left,
		Op:    op,
		Right: right,
	}, nil
}

func (p *Parser) addSubExpr(tokens *stream[token.Token]) (ast.Expr, error) {
	left, err := p.multiplyDivideExpr(tokens)
	if err != nil {
		return nil, err
	}

	next := tokens.peek().Type
	if next != token.Add && next != token.Minus {
		return left, nil
	}

	op := tokens.take()
	right, err := p.expr(tokens)
	if err != nil {
		return nil, ekit.NewCondition(tokens.peek(),
			fmt.Sprintf("I expected an expression on the right side of the '%s' symbol. Instead I found '%s'", op.Lexeme, tokens.peek().Lexeme),
		)
	}

	return &ast.Binary{
		Left:  left,
		Op:    op,
		Right: right,
	}, nil
}

func (p *Parser) multiplyDivideExpr(tokens *stream[token.Token]) (ast.Expr, error) {
	left, err := p.baseExpr(tokens)
	if err != nil {
		return nil, err
	}

	next := tokens.peek().Type
	if next != token.Star && next != token.Divide {
		return left, nil
	}

	op := tokens.take()
	right, err := p.expr(tokens)
	if err != nil {
		return nil, ekit.NewCondition(tokens.peek(),
			fmt.Sprintf("I expected an expression on the right side of the '%s' symbol. Instead I found '%s'", op.Lexeme, tokens.peek().Lexeme),
		)
	}

	return &ast.Binary{
		Left:  left,
		Op:    op,
		Right: right,
	}, nil
}

func (p *Parser) baseExpr(tokens *stream[token.Token]) (ast.Expr, error) {
	switch tokens.peek().Type {
	case token.IntLiteral:
		valueToken := tokens.take()
		value, err := strconv.ParseInt(valueToken.Lexeme, 10, 64)
		if err != nil {
			return nil, ekit.NewCondition(
				valueToken,
				fmt.Sprintf("invalid int literal '%s'", valueToken.Lexeme),
			)
		}

		return &ast.IntLiteral{Value: value}, nil
	case token.BoolLiteral:
		valueToken := tokens.take()
		if valueToken.Lexeme == "true" {
			return &ast.BoolLiteral{Value: true}, nil
		}
		if valueToken.Lexeme == "false" {
			return &ast.BoolLiteral{Value: false}, nil
		}
		return nil, ekit.NewCondition(
			valueToken,
			fmt.Sprintf("invalid bool literal '%s'", valueToken.Lexeme),
		)
	case token.PathLiteral:
		valueToken := tokens.take()
		// TODO: validate the path
		return &ast.PathLiteral{Value: valueToken.Lexeme}, nil
	case token.Identifyer:
		nameToken := tokens.take()
		return &ast.Identifyer{Name: nameToken.Lexeme}, nil
	case token.OpenParen:
		// take the open paren
		tokens.take()
		group, err := p.expr(tokens)
		if err != nil {
			return nil, err
		}

		if got := tokens.take(); got.Type != token.CloseParen {
			return nil, ekit.NewCondition(got,
				fmt.Sprintf("I expected the opening paren to be matched by a closing paren, instead I found '%s'", got.Lexeme),
			)
		}

		return group, nil
	default:
		return nil, ekit.NewCondition(
			tokens.peek(),
			fmt.Sprintf("'%s' is a not valid way to start an expression", tokens.peek().Lexeme),
			fmt.Sprintf("I expected there to be an expression here"),
		)
	}
}
