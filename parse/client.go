package parse

import (
	"fmt"

	"github.com/bjatkin/yok/slice"
	"github.com/bjatkin/yok/sym"
)

type Client struct {
	table    *sym.Table
	patterns []pat
	parsers  []parser
}

func NewClient(table *sym.Table) *Client {
	return &Client{
		table: table,
		patterns: []pat{
			newRegPat(`#[^\n\r]*\n`, Comment),
			newSPat("\n", NewLine),
			newRegPat(`[\t ]+`, WhiteSpace),
			newSPat("=", SetOp),
			newSPat("if", IfKeyword),
			newSPat("let", LetKeyword),
			newSPat("string", TypeKeyword),
			newSPat("int", TypeKeyword),
			newSPat("bool", TypeKeyword),
			newSPat("path", TypeKeyword),
			newSPat("true", Value),
			newSPat("false", Value),
			newSPat("use", UseKeyword),
			newSPat("as", AsKeyword),
			newSPat("env", EnvKeyword),
			newSPat("{", OpenBlock),
			newSPat("}", CloseBlock),
			newSPat("(", OpenCall),
			newSPat(")", CloseCall),
			newSPat("[", OpenIndex),
			newSPat("]", CloseIndex),
			newSPat(",", Comma),
			newRegPat(`[0-9]+`, Value),
			newRegPat(`(\.|\.\.|~){0,1}\/[^ \(\)\[\]\{\}\n]+`, Value),
			newRegPat(`[a-zA-Z][a-zA-Z0-9_]*`, Identifyer),
			newSPat("==", CompOp),
			newSPat("!=", CompOp),
			newSPat(">", CompOp),
			newSPat(">=", CompOp),
			newSPat("<", CompOp),
			newSPat("<=", CompOp),
			newSPat("&&", BoolOp),
			newSPat("||", BoolOp),
			newRegPat(`[><\-\+\-\*\/]`, BinaryOp),
			newSPat(".", Dot),
			stringValuePat{},
		},
		parsers: []parser{
			parseDecl,
			parseAssign,
			parseEnv,
			parseComment,
			parseNewLine,
			parseUseBlock,
			parseCall,
			parseIfBlock,
		},
	}
}

type parser func(slice.Itter[Token]) parseMatch

func (c *Client) Parse(tokens []Token) (Node, error) {
	itter := slice.NewIttr(tokens)
	match := c.parse(itter)
	if len(match.nodes) > 0 {
		return match.nodes[0], nil
	}

	return Node{}, fmt.Errorf("failed to build parse tree")
}

func (c *Client) parse(itter slice.Itter[Token]) parseMatch {
	root := Node{Type: Root}

	for itter.Continue() {
		var match parseMatch
		for _, parse := range c.parsers {
			if match = parse(itter); match.count > 0 {
				itter.Pop(match.count)
				root.Nodes = append(root.Nodes, match.nodes...)
				break
			}
		}
		if match.count == 0 {
			fmt.Println("unknown sequence: ", itter.All())
			return parseMatch{}
		}
	}

	return parseMatch{
		count: 1,
		nodes: []Node{root},
	}
}
