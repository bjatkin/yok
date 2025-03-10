package repr

import (
	"fmt"
	"strings"
)

const (
	// indentToken is used as the indent string for the rendered repr
	indentToken = "    "
	maxLineLen  = 100
)

// Value is any repr value that can be rendered
type Value interface {
	Render(depth int) string
}

// Array is an array of repr values
type Array struct {
	values []Value
}

// AddValue adds a new json value to the array
func (a *Array) AddValue(value Value) {
	a.values = append(a.values, value)
}

// Render the array with the proper indent
func (a Array) Render(depth int) string {
	if len(a.values) == 0 {
		return "[]"
	}

	indent := strings.Repeat(indentToken, depth+1)
	// initial line leng is the indent plus array charater (i.e. [])
	lineLen := len(indent) + 2

	indentedValues := []string{}
	values := []string{}
	for _, value := range a.values {
		valueString := value.Render(depth + 1)
		// +2 to account for the ", " between items
		lineLen += len(valueString) + 2
		values = append(values, valueString)

		valueString = indent + valueString
		indentedValues = append(indentedValues, valueString)
	}

	if lineLen < maxLineLen {
		return "[ " + strings.Join(values, ", ") + " ]"
	}

	indent = strings.Repeat(indentToken, depth)
	stackedItems := strings.Join(indentedValues, ",\n") + "\n"
	return "[\n" + stackedItems + indent + "]"
}

// Field is a repr key value field
type Field struct {
	key   string
	value Value
}

// NewField creates a new Field struct with the given key and value
func NewField(key string, value Value) Field {
	return Field{key: key, value: value}
}

// Render the field with the proper indent
func (f Field) Render(depth int) string {
	return fmt.Sprintf(`%s=%s`, f.key, f.value.Render(depth))
}

// Object is a repr object
type Object struct {
	name   string
	fields []Field
}

// NewObject creates a new object with the given name and fields
func NewObject(name string, fields ...Field) Object {
	return Object{name: name, fields: fields}
}

// AddFields adds one or more new fields to the repr object
func (o *Object) AddFields(fields ...Field) {
	o.fields = append(o.fields, fields...)
}

// Render the json object with the correct indent
func (o Object) Render(depth int) string {
	if len(o.fields) == 0 {
		return fmt.Sprintf("%s()", o.name)
	}

	// initial line leng is the indent plus the name, pluse the parens (i.e. ())
	lineLen := len(indentToken)*depth + 2
	fields := []string{}
	for _, field := range o.fields {
		fieldString := field.Render(depth + 1)
		lineLen += len(fieldString) + 2
		fields = append(fields, fieldString)
	}

	if lineLen < maxLineLen {
		return o.name + "(" + strings.Join(fields, ", ") + ")"
	}

	indent := strings.Repeat(indentToken, depth+1)
	stackedFields := indent + strings.Join(fields, ",\n"+indent) + ",\n"

	indent = strings.Repeat(indentToken, depth)
	return o.name + "(\n" + stackedFields + indent + ")"
}

// nil is the repr empty value
type Nil struct{}

// Render the nil value
func (n Nil) Render(depth int) string {
	return "nil"
}

// String is a repr string value
type String string

// Render the string as a valid repr value
func (s String) Render(depth int) string {
	return fmt.Sprintf(`"%s"`, s)
}

// Bool is a repr bool value
type Bool bool

// Render the bool as a valid repr value
func (b Bool) Render(depth int) string {
	return fmt.Sprintf("%v", b)
}

// Int is a repr integer value
type Int int

// Render the integer as a valid repr value
func (i Int) Render(depth int) string {
	return fmt.Sprintf("%d", i)
}
