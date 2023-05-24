package parse

import (
	"fmt"
	"strings"
)

// TODO: this should probably be an int but it makes debugging a pain.
// maybe if I can find a better way to work with int constants I'll swap this.
type NodeType string

const (
	Uknown       = NodeType("")
	Root         = NodeType("root")
	Value        = NodeType("value")
	Comment      = NodeType("comment")
	NewLineGroup = NodeType("newlinegroup")
	NewLine      = NodeType("newline")
	WhiteSpace   = NodeType("whitespace")
	BinaryOp     = NodeType("binaryop")
	CompOp       = NodeType("compop")
	BoolOp       = NodeType("boolop")
	SetOp        = NodeType("setop")
	Expr         = NodeType("expr")
	IfKeyword    = NodeType("ifkeyword")
	EnvKeyword   = NodeType("envkeyword")
	LetKeyword   = NodeType("letkeyword")
	TypeKeyword  = NodeType("typekeyword")
	UseKeyword   = NodeType("usekeyword")
	ImportExpr   = NodeType("importexpr")
	AsKeyword    = NodeType("askeyword")
	OpenBlock    = NodeType("openblock")
	CloseBlock   = NodeType("closeblock")
	OpenCall     = NodeType("opencall")
	CloseCall    = NodeType("closecall")
	OpenIndex    = NodeType("openindex")
	CloseIndex   = NodeType("closeindex")
	Identifyer   = NodeType("identifyer")
	Assign       = NodeType("assign")
	Decl         = NodeType("decl")
	Call         = NodeType("call")
	Comma        = NodeType("comma")
	Dot          = NodeType("dot")
	Arg          = NodeType("arg")
	Body         = NodeType("body")
)

type Node struct {
	Token    Token
	NodeType NodeType
	Nodes    []Node
}

// func (n Node) Clone() Node {
// 	var nodes []Node
// 	for _, n := range n.Nodes {
// 		nodes = append(nodes, n.Clone())
// 	}

// 	return Node{
// 		ID:       n.ID,
// 		NodeType: n.NodeType,
// 		Value:    n.Value,
// 		Nodes:    nodes,
// 	}
// }

func (n Node) String() string {
	var sub []string
	for _, node := range n.Nodes {
		sub = append(sub, node.String())
	}

	if len(sub) > 0 {
		return fmt.Sprintf("%s(%s) [ %s ]", n.NodeType, n.Token.Value, strings.Join(sub, ", "))
	}

	return fmt.Sprintf("%s(%s)", n.NodeType, n.Token.Value)
}

// func CloneNodes(nodes []Node) []Node {
// 	var ret []Node
// 	for _, node := range nodes {
// 		ret = append(ret, node.Clone())
// 	}
// 	return ret
// }
