package gensh

import (
	"fmt"
	"strings"
)

// indentToken is used as the indent string for the rendered code
const indentToken = "    "

// codeUnit is a unit of code. It can contain a single line and
// any number of children units
type codeUnit struct {
	line     string
	children []codeUnit
}

// newCodeUnitf creates a new codeUnit from a format string and arguments
func newCodeUnitf(format string, a ...any) codeUnit {
	line := fmt.Sprintf(format, a...)
	return codeUnit{line: line}
}

// addChildren adds the given code units as children of the parent unit
func (f *codeUnit) addChildren(units []codeUnit) {
	f.children = append(f.children, units...)
}

// render the code unit into a slice of strings
func (f codeUnit) render(depth int) []string {
	indent := strings.Repeat(indentToken, depth)
	lines := []string{indent + f.line}
	for _, child := range f.children {
		childLines := child.render(depth + 1)
		lines = append(lines, childLines...)
	}

	return lines
}

// codeBuilder is used to take a slice of code units and convert them into
// well formated text
type codeBuilder struct {
	units []codeUnit
}

// newCodeBuilder creates a new codeBuilder where the first unit contains
// the given line
func newCodeBuilder(lines ...string) codeBuilder {
	units := []codeUnit{}
	for _, line := range lines {
		units = append(units, codeUnit{line: line})
	}

	return codeBuilder{units: units}
}

// addLine adds a new unit to the codeBuilder with the given line
func (s *codeBuilder) addLine(line string) *codeUnit {
	unit := codeUnit{line: line}
	s.units = append(s.units, unit)
	return &unit
}

// addUnit adds a new codeUnit to the codeBuilder
func (s *codeBuilder) addUnit(unit codeUnit) {
	s.units = append(s.units, unit)
}

// addUnits adds a slice of codeUnits to the code builder
func (s *codeBuilder) addUnits(unit []codeUnit) {
	s.units = append(s.units, unit...)
}

// render the codeUnits into well formated text
func (s codeBuilder) render() string {
	units := []string{}
	for _, fragment := range s.units {
		lines := fragment.render(0)
		unit := strings.Join(lines, "\n")
		units = append(units, unit)
	}

	return strings.Join(units, "\n")
}
