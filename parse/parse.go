package parse

import (
	"github.com/bjatkin/yok/slice"
)

func oneOf(parsers ...parser) parser {
	return func(itter slice.Itter[Node]) parseMatch {
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
	return func(itter slice.Itter[Node]) parseMatch {
		ret := parseMatch{}
		for itter.Continue() {
			match := p(itter)
			if match.count == 0 {
				break
			}
			ret.count += match.count
			itter.Pop(match.count)
			ret.nodes = append(ret.nodes, CloneNodes(match.nodes)...)
		}
		return ret
	}
}

func typeSequence(seq ...NodeType) parser {
	return func(itter slice.Itter[Node]) parseMatch {
		ret := parseMatch{}

		// reset the itterator so the loop works correctly
		itter = slice.NewIttr(itter.All())
		for i := 0; itter.Next() && i < len(seq); i++ {
			if itter.Item().NodeType != seq[i] {
				return parseMatch{}
			}
			ret.count++
			ret.nodes = append(ret.nodes, itter.Item().Clone())
		}

		return ret
	}
}

func sequence(parsers ...parser) parser {
	return func(itter slice.Itter[Node]) parseMatch {
		ret := parseMatch{}
		for _, p := range parsers {
			match := p(itter)
			if match.count == 0 {
				return parseMatch{}
			}

			ret.count += match.count
			ret.nodes = append(ret.nodes, CloneNodes(match.nodes)...)
			itter.Pop(match.count)
		}
		return ret
	}
}

func tree(root Node, p parser) parser {
	return func(itter slice.Itter[Node]) parseMatch {
		match := p(itter)
		if match.count == 0 {
			return parseMatch{}
		}
		clone := root.Clone()
		clone.Nodes = append(clone.Nodes, CloneNodes(match.nodes)...)

		return parseMatch{
			count: match.count,
			nodes: []Node{clone},
		}
	}
}

func typeTree(root NodeType, p parser) parser {
	return func(itter slice.Itter[Node]) parseMatch {
		if itter.Item().NodeType != root {
			return parseMatch{}
		}

		n := itter.Pop(1)
		if len(n) == 0 {
			return parseMatch{}
		}
		root := n[0].Clone()

		match := p(itter)
		if match.count == 0 {
			return parseMatch{}
		}
		root.Nodes = CloneNodes(match.nodes)

		return parseMatch{
			count: 1 + match.count,
			nodes: []Node{root},
		}
	}
}

func typeUntil(t NodeType) parser {
	return func(itter slice.Itter[Node]) parseMatch {
		ret := parseMatch{}

		// reset the itterator so the loop works correctly
		itter = slice.NewIttr(itter.All())
		for itter.Next() {
			if itter.Item().NodeType == t {
				return ret
			}

			ret.count++
			ret.nodes = append(ret.nodes, itter.Item().Clone())
		}
		return ret
	}
}

type parseMatch struct {
	count int
	nodes []Node
}

var parseNewLine = tree(Node{NodeType: NewLineGroup},
	repeat(typeSequence(NewLine)),
)

var parseAssign = tree(Node{NodeType: Assign},
	oneOf(
		typeSequence(Identifyer, SetOp, Identifyer, NewLine),
		typeSequence(Identifyer, SetOp, Value, NewLine),
	),
)

var parseDecl = tree(Node{NodeType: Decl}, typeSequence(LetKeyword, Identifyer, TypeKeyword))

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
			tree(Node{NodeType: ImportExpr},
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

var parseCall = tree(Node{NodeType: Call},
	oneOf(
		sequence(
			parseCallPrefix,

			// get all the normal arguments
			repeat(
				tree(Node{NodeType: Arg},
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
			tree(Node{NodeType: Arg},
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

func parseIfBlock(itter slice.Itter[Node]) parseMatch {
	// this has to be a function so that initialization cycles dont occure
	return typeTree(IfKeyword,
		sequence(
			tree(Node{NodeType: Expr},
				typeUntil(OpenBlock),
			),
			parseBlock,
		),
	)(itter)
}

// TODO: I feel like I should be able to simplify this a little
// especially now that the .parse function is available on the parse client
func parseBlock(itter slice.Itter[Node]) parseMatch {
	prefix := typeSequence(OpenBlock, NewLine)(itter)
	if prefix.count == 0 {
		return parseMatch{}
	}
	itter.Pop(prefix.count)
	count := prefix.count

	var block []Node
	var indent int
	for itter.Continue() {

		match := typeSequence(OpenBlock, NewLine)(itter)
		if match.count > 0 {
			indent++
			block = append(block, CloneNodes(match.nodes)...)
			count += match.count
			itter.Pop(match.count)
			continue
		}

		match = typeSequence(CloseBlock, NewLine)(itter)
		if match.count > 0 {
			if indent == 0 {
				break
			}
			indent--
			block = append(block, CloneNodes(match.nodes)...)
			count += match.count
			itter.Pop(match.count)
			continue
		}

		count++
		block = append(block, itter.Item().Clone())
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
