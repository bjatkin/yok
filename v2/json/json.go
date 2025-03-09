package json

import (
	"fmt"
	"strings"
)

// indentToken is used as the indent string for the rendered json
const indentToken = "    "

// Value is any json value that can be rendered
type Value interface {
	Render(depth int) string
}

// Array is a json array of values
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
	items := []string{}
	for _, value := range a.values {
		valueString := value.Render(depth + 1)
		valueString = indent + valueString
		items = append(items, valueString)
	}

	indent = strings.Repeat(indentToken, depth)
	stackedItems := strings.Join(items, ",\n") + "\n"
	return "[\n" + stackedItems + indent + "]"
}

// Field is a json key value field
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
	return fmt.Sprintf(`"%s": %s`, f.key, f.value.Render(depth))
}

// Object is a json object
type Object struct {
	fields []Field
}

// NewObject creates a new object with the given fields
func NewObject(fields ...Field) Object {
	return Object{fields: fields}
}

// AddFields adds one or more new fields to the json object
func (o *Object) AddFields(fields ...Field) {
	o.fields = append(o.fields, fields...)
}

// Render the json object with the correct indent
func (o Object) Render(depth int) string {
	if len(o.fields) == 0 {
		return "{}"
	}

	indent := strings.Repeat(indentToken, depth+1)
	fields := []string{}
	for _, field := range o.fields {
		f := field.Render(depth + 1)
		f = indent + f
		fields = append(fields, f)
	}

	indent = strings.Repeat(indentToken, depth)
	stackedFields := strings.Join(fields, ",\n") + "\n"
	return "{\n" + stackedFields + indent + "}"
}

// Null is the json null value
type Null struct{}

// Render the null value
func (n Null) Render(depth int) string {
	return "null"
}

// String is a json string value
type String string

// Render the string as a value json value
func (s String) Render(depth int) string {
	return fmt.Sprintf(`"%s"`, s)
}

// Int is a json integer value
type Int int

// Render the integer as a valid json value
func (i Int) Render(depth int) string {
	return fmt.Sprintf("%d", i)
}
