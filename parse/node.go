package parse

import (
	"fmt"
	"strings"

	"github.com/bjatkin/yok/sym"
)

// TODO: this should probably be an int but it makes debugging a pain.
// maybe if I can find a better way to work with int constants I'll swap this.
type Type string

const (
	Uknown       = Type("")
	Root         = Type("root")
	Value        = Type("value")
	Comment      = Type("comment")
	NewLineGroup = Type("newlinegroup")
	NewLine      = Type("newline")
	WhiteSpace   = Type("whitespace")
	BinaryOp     = Type("binaryop")
	CompOp       = Type("compop")
	BoolOp       = Type("boolop")
	SetOp        = Type("setop")
	Expr         = Type("expr")
	IfKeyword    = Type("ifkeyword")
	EnvKeyword   = Type("envkeyword")
	LetKeyword   = Type("letkeyword")
	TypeKeyword  = Type("typekeyword")
	UseKeyword   = Type("usekeyword")
	ImportExpr   = Type("importexpr")
	AsKeyword    = Type("askeyword")
	OpenBlock    = Type("openblock")
	CloseBlock   = Type("closeblock")
	OpenCall     = Type("opencall")
	CloseCall    = Type("closecall")
	OpenIndex    = Type("openindex")
	CloseIndex   = Type("closeindex")
	Identifyer   = Type("identifyer")
	Assign       = Type("assign")
	Decl         = Type("decl")
	Call         = Type("call")
	Comma        = Type("comma")
	Dot          = Type("dot")
	Arg          = Type("arg")
	Body         = Type("body")
)

type Node struct {
	ID    sym.ID
	Value string
	Type  Type
	Nodes []Node
}

func (n Node) String() string {
	var sub []string
	for _, node := range n.Nodes {
		sub = append(sub, node.String())
	}

	if len(sub) > 0 {
		return fmt.Sprintf("%s(%s) [ %s ]", n.Type, n.Value, strings.Join(sub, ", "))
	}

	return fmt.Sprintf("%s(%s)", n.Type, n.Value)
}
