package parse

import (
	"github.com/bjatkin/yok/slice"
)

func oneOf(parsers ...parser) parser {
	return func(itter slice.Itter[Token]) parseMatch {
		for _, p := range parsers {
			match := p(itter)
			if match.count > 0 {
				return match
			}
		}
		return parseMatch{}
	}
}

func repeat(p parser) parser {
	return func(itter slice.Itter[Token]) parseMatch {
		ret := parseMatch{}
		for itter.Continue() {
			match := p(itter)
			if match.count == 0 {
				break
			}
			ret.count += match.count
			itter.Pop(match.count)
			ret.nodes = append(ret.nodes, match.nodes...)
		}
		return ret
	}
}

func nest(p parser) parser {
	return func(itter slice.Itter[Token]) parseMatch {
		ret := parseMatch{}
		match := p(itter)
		if match.count == 0 {
			return ret
		}
		ret.count += match.count
		itter.Pop(match.count)
		ret.nodes = append(ret.nodes, match.nodes...)

		current := &ret.nodes[len(ret.nodes)-1]
		for itter.Continue() {
			match := p(itter)
			if match.count == 0 {
				break
			}
			ret.count += match.count
			itter.Pop(match.count)
			current.Nodes = append(current.Nodes, match.nodes...)
			current = &current.Nodes[len(current.Nodes)-1]
		}
		return ret
	}
}

func typeSequence(seq ...Type) parser {
	return func(itter slice.Itter[Token]) parseMatch {
		ret := parseMatch{}

		// reset the itterator so the loop works correctly
		itter = slice.NewIttr(itter.All())
		for i := 0; itter.Next() && i < len(seq); i++ {
			if itter.Item().Type != seq[i] {
				return parseMatch{}
			}
			ret.count++
			token := itter.Item()
			ret.nodes = append(ret.nodes, Node{ID: token.ID, Value: token.Value, Type: token.Type})
		}

		return ret
	}
}

func sequence(parsers ...parser) parser {
	return func(itter slice.Itter[Token]) parseMatch {
		ret := parseMatch{}
		for _, p := range parsers {
			match := p(itter)
			if match.count == 0 {
				return parseMatch{}
			}

			ret.count += match.count
			ret.nodes = append(ret.nodes, match.nodes...)
			itter.Pop(match.count)
		}
		return ret
	}
}

func tree(root Type, p parser) parser {
	return func(itter slice.Itter[Token]) parseMatch {
		match := p(itter)
		if match.count == 0 {
			return parseMatch{}
		}
		clone := Node{Type: root}
		clone.Nodes = append(clone.Nodes, match.nodes...)

		return parseMatch{
			count: match.count,
			nodes: []Node{clone},
		}
	}
}

func typeTree(root Type, p parser) parser {
	return func(itter slice.Itter[Token]) parseMatch {
		if itter.Item().Type != root {
			return parseMatch{}
		}

		n := itter.Pop(1)
		if len(n) == 0 {
			return parseMatch{}
		}
		token := n[0]
		root := Node{ID: token.ID, Value: token.Value, Type: token.Type}

		match := p(itter)
		if match.count == 0 {
			return parseMatch{}
		}
		root.Nodes = match.nodes

		return parseMatch{
			count: 1 + match.count,
			nodes: []Node{root},
		}
	}
}

type parseMatch struct {
	count int
	nodes []Node
}

var parseNewLine = tree(NewLineGroup,
	repeat(typeSequence(NewLine)),
)

var parseAssign = tree(Assign,
	sequence(
		typeSequence(Identifyer, SetOp),
		nest(parseExpr),
		typeSequence(NewLine),
	),
)

var parseDecl = tree(Decl, typeSequence(LetKeyword, Identifyer, TypeKeyword))

var parseEnv = typeTree(EnvKeyword, typeSequence(OpenIndex, Value, CloseIndex, NewLine))

var parseComment = typeSequence(Comment)

var parseUseBlock = typeTree(UseKeyword,
	sequence(
		// start of the block
		oneOf(
			typeSequence(OpenBlock, NewLine),
			typeSequence(OpenBlock),
		),

		// actual imports
		repeat(
			tree(ImportExpr,
				oneOf(
					typeSequence(Identifyer, NewLine),
					typeSequence(Identifyer, AsKeyword, Identifyer, NewLine),
					typeSequence(Value, AsKeyword, Identifyer, NewLine),
					typeSequence(Identifyer),
				),
			),
		),

		// end of the block
		typeSequence(CloseBlock, NewLine),
	),
)

var parseCall = tree(Call,
	oneOf(
		sequence(
			parseCallPrefix,

			// get all the normal arguments
			repeat(
				tree(Arg,
					oneOf(
						typeSequence(Value, Comma, NewLine),
						typeSequence(Identifyer, Comma, NewLine),
						typeSequence(Value, Comma),
						typeSequence(Identifyer, Comma),
						typeSequence(Value),
						typeSequence(Identifyer),
					),
				),
			),

			// end of the call
			typeSequence(CloseCall, NewLine),
		),

		// special case when the call has no arguments
		sequence(
			parseCallPrefix,

			// end of the call
			typeSequence(CloseCall, NewLine),
		),
	),
)

var parseCallPrefix = oneOf(
	// basic call with no sub commands
	oneOf(
		typeSequence(Identifyer, OpenCall, NewLine),
		typeSequence(Identifyer, OpenCall),
	),

	// command call with sub commands
	sequence(
		typeSequence(Identifyer),
		repeat(
			tree(Arg,
				oneOf(
					typeSequence(Dot, NewLine, Identifyer),
					typeSequence(Dot, Identifyer),
				),
			),
		),
		oneOf(
			typeSequence(OpenCall, NewLine),
			typeSequence(OpenCall),
		),
	),
)

var parseExpr = oneOf(
	tree(Expr, oneOf(
		typeSequence(Identifyer, BinaryOp),
		typeSequence(Value, BinaryOp),
	)),
	typeSequence(Identifyer),
	typeSequence(Value),
)

func parseIfBlock(itter slice.Itter[Token]) parseMatch {
	return typeTree(IfKeyword,
		sequence(
			parseExpr,
			parseBlock,
		),
	)(itter)
}

// TODO: I feel like I should be able to simplify this a little
// especially now that the .parse function is available on the parse client
func parseBlock(itter slice.Itter[Token]) parseMatch {
	prefix := typeSequence(OpenBlock, NewLine)(itter)
	if prefix.count == 0 {
		return parseMatch{}
	}
	itter.Pop(prefix.count)
	count := prefix.count

	var block []Token
	var indent int
	for itter.Continue() {

		match := typeSequence(OpenBlock, NewLine)(itter)
		if match.count > 0 {
			indent++
			count += match.count
			block = append(block, itter.Pop(match.count)...)
			continue
		}

		match = typeSequence(CloseBlock, NewLine)(itter)
		if match.count > 0 {
			if indent == 0 {
				break
			}
			indent--
			count += match.count
			block = append(block, itter.Pop(match.count)...)
			continue
		}

		count++
		block = append(block, itter.Item())
		itter.Pop(1)
	}
	suffix := typeSequence(CloseBlock, NewLine)(itter)
	if suffix.count == 0 {
		return parseMatch{}
	}
	count += suffix.count

	// TODO: should probably pass a symbol table in here
	client := NewClient(nil)
	root, err := client.Parse(block)
	if err != nil {
		return parseMatch{}
	}

	return parseMatch{
		count: count,
		nodes: append(append(prefix.nodes, root), suffix.nodes...),
	}
}
