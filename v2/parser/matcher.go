package parser

import (
	"regexp"

	"github.com/bjatkin/yok/v2/token"
)

type matcher interface {
	match(src *stream[rune]) (token.Token, bool)
}

type stringMatch struct {
	check     string
	tokenType token.Type
}

func newStringMatch(tokenType token.Type, check string) stringMatch {
	return stringMatch{
		check:     check,
		tokenType: tokenType,
	}
}

func (m stringMatch) match(src *stream[rune]) (token.Token, bool) {
	checkLen := len(m.check)
	if string(src.peekN(checkLen)) != m.check {
		return token.Empty, false
	}

	start := src.current
	return token.Token{
		Start: start,
		End:   start + checkLen,

		Type:   m.tokenType,
		Lexeme: m.check,
	}, true
}

type regexMatch struct {
	pattern   *regexp.Regexp
	tokenType token.Type
}

func newRegexMatch(tokenType token.Type, pattern string) regexMatch {
	regex := regexp.MustCompile("^" + pattern)
	return regexMatch{
		tokenType: tokenType,
		pattern:   regex,
	}
}

func (m regexMatch) match(src *stream[rune]) (token.Token, bool) {
	line := string(src.peekUntil(func(t rune) bool { return t == '\n' }))
	match := m.pattern.FindString(line)
	if match == "" {
		return token.Empty, false
	}

	start := src.current
	matchLen := len(match)
	return token.Token{
		Start: start,
		End:   start + matchLen,

		Type:   m.tokenType,
		Lexeme: match,
	}, true
}

type funcMatch struct {
	tokenType token.Type
	matchFn   func(*stream[rune]) (string, bool)
}

func newFuncMatch(tokenType token.Type, matchFunc func(*stream[rune]) (string, bool)) funcMatch {
	return funcMatch{
		tokenType: tokenType,
		matchFn:   matchFunc,
	}
}

func (m funcMatch) match(src *stream[rune]) (token.Token, bool) {
	match, ok := m.matchFn(src)
	if !ok {
		return token.Empty, ok
	}

	start := src.current
	matchLen := len(match)
	return token.Token{
		Start: start,
		End:   start + matchLen,

		Type:   m.tokenType,
		Lexeme: match,
	}, true
}
