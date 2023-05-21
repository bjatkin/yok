package parse

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/bjatkin/yok/slice"
	"github.com/bjatkin/yok/sym"
)

// TODO: we should consider lexing to an token rather than directly into a stream of nodes.
// Doing this means we have to do a lot of node cloning to prevent possible pointer cycles in the
// final tree structure we build.
// Converting from a token into a node would be a simple way to ensure those cycles don't occure
// and prevent situations were a node is cloned multiple times
func (c *Client) Lex(file string, code []byte) ([]Node, error) {
	var tokens []Node
	itter := slice.NewIttr([]rune(string(code)))
	var line, col int

	for itter.Continue() {
		var match lexMatch
		for _, pattern := range c.patterns {
			if match = pattern.lex(itter); match.ok {
				value := itter.Pop(match.count)
				col += match.count
				if string(value) == "\n" {
					col = 0
					line++
				}

				if match.nodeType == WhiteSpace {
					break
				}

				id := c.table.AddSymbol(&sym.Symbol{
					Value: string(value),
					File:  file,
					Col:   col,
					Line:  line,
				})

				tokens = append(tokens, Node{
					ID:       id,
					NodeType: match.nodeType,
					Value:    string(value),
				})
				break
			}
		}
		if !match.ok {
			return nil, fmt.Errorf("unknown token %s", string(itter.All()))
		}
	}
	return tokens, nil
}

type lexMatch struct {
	ok       bool
	nodeType NodeType
	count    int
}

type pat interface {
	lex(itter slice.Itter[rune]) lexMatch
}

type sPat struct {
	pat      string
	nodeType NodeType
}

func newSPat(pat string, nodeType NodeType) pat {
	return &sPat{
		pat:      pat,
		nodeType: nodeType,
	}
}

func (s *sPat) lex(itter slice.Itter[rune]) lexMatch {
	check := string(itter.Pop(len(s.pat)))
	if check == s.pat {
		return lexMatch{
			ok:       true,
			count:    len(s.pat),
			nodeType: s.nodeType,
		}
	}
	return lexMatch{}
}

type regPat struct {
	reg      *regexp.Regexp
	nodeType NodeType
}

func newRegPat(pat string, nodeType NodeType) pat {
	if !strings.HasPrefix("^", pat) {
		pat = "^" + pat
	}

	return &regPat{
		reg:      regexp.MustCompile(pat),
		nodeType: nodeType,
	}
}

func (m regPat) lex(itter slice.Itter[rune]) lexMatch {
	matchCount := len(m.reg.FindString(string(itter.All())))
	return lexMatch{
		ok:       matchCount > 0,
		count:    matchCount,
		nodeType: m.nodeType,
	}
}

type stringValuePat struct{}

func (s stringValuePat) lex(i slice.Itter[rune]) lexMatch {
	if i.Item() != '"' {
		return lexMatch{}
	}
	i.Next()

	var escape bool
	// this starts at 2 because the first loop is for the second rune
	// the first rune is already known to be "
	for matchCount := 2; i.Next(); matchCount++ {
		if i.Item() == '\\' {
			escape = true
			continue
		}
		if i.Item() == '\n' || i.Item() == '\r' {
			return lexMatch{}
		}
		if i.Item() == '"' && !escape {
			return lexMatch{
				ok:       true,
				count:    matchCount,
				nodeType: Value,
			}
		}
		escape = false
	}

	return lexMatch{}
}
