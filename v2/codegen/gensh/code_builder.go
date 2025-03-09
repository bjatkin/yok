package gensh

import (
	"fmt"
	"strings"
)

const indentToken = "    "

type codeUnit struct {
	line     string
	children []codeUnit
}

func newCodeUnitf(format string, a ...any) codeUnit {
	line := fmt.Sprintf(format, a...)
	return codeUnit{line: line}
}

func (f *codeUnit) addChildren(units []codeUnit) {
	f.children = append(f.children, units...)
}

func (f codeUnit) render(depth int) []string {
	indent := strings.Repeat(indentToken, depth)
	lines := []string{indent + f.line}
	for _, child := range f.children {
		childLines := child.render(depth + 1)
		lines = append(lines, childLines...)
	}

	return lines
}

type codeBuilder struct {
	units []codeUnit
}

func newCodeBuilder(line string) codeBuilder {
	return codeBuilder{units: []codeUnit{{line: line}}}
}

func (s *codeBuilder) addLine(line string) *codeUnit {
	unit := codeUnit{line: line}
	s.units = append(s.units, unit)
	return &unit
}

func (s *codeBuilder) addLinef(format string, a ...any) *codeUnit {
	line := fmt.Sprintf(format, a...)
	unit := codeUnit{line: line}
	s.units = append(s.units, unit)
	return &unit
}

func (s *codeBuilder) addUnit(unit codeUnit) {
	s.units = append(s.units, unit)
}

func (s *codeBuilder) addUnits(unit []codeUnit) {
	s.units = append(s.units, unit...)
}

func (s codeBuilder) render() string {
	units := []string{}
	for _, fragment := range s.units {
		lines := fragment.render(0)
		unit := strings.Join(lines, "\n")
		units = append(units, unit)
	}

	return strings.Join(units, "\n")
}
