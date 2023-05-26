package source

import (
	"fmt"
	"strings"
)

const indent = "  "

type NewLine struct{}

func (y NewLine) String() string {
	return ""
}

type Line string

func Linef(format string, a ...any) Line {
	return Line(fmt.Sprintf(format, a...))
}

func (y Line) String() string {
	return string(y)
}

type Import struct {
	MaxNameLen int
	Name       string
	Alias      string
}

func (y Import) String() string {
	if y.Alias == "" {
		return y.Name
	}

	padd := y.MaxNameLen - len(y.Name)
	return y.Name + strings.Repeat(" ", padd) + " as " + y.Alias
}

type Block struct {
	indent int
	Lines  []fmt.Stringer
}

func (y Block) String() string {
	var yok []string
	blockIndent := strings.Repeat(indent, y.indent)

	for _, line := range y.Lines {
		switch v := line.(type) {
		case Block:
			v.indent = y.indent + 1
			yok = append(yok, v.String())
		case PrefixBlock:
			v.indent = y.indent
			yok = append(yok, v.String())
		case NewLine:
			yok = append(yok, v.String())
		default:
			yok = append(yok, blockIndent+v.String())
		}
	}

	return strings.Join(yok, "\n")
}

type PrefixBlock struct {
	indent int
	Prefix fmt.Stringer
	Block  Block
	Suffix fmt.Stringer
}

func (y PrefixBlock) String() string {
	var yok []string
	prefixIndent := strings.Repeat(indent, y.indent)

	if y.Prefix != nil {
		yok = append(yok, prefixIndent+y.Prefix.String())
	}

	y.Block.indent = y.indent + 1
	yok = append(yok, y.Block.String())

	if y.Suffix != nil {
		yok = append(yok, prefixIndent+y.Suffix.String())
	}

	return strings.Join(yok, "\n")
}
