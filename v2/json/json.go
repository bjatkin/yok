package json

import (
	"fmt"
	"strings"
)

const indentToken = "    "

type Value interface {
	Render(depth int) string
}

type Array struct {
	values []Value
}

func (a *Array) AddItem(value Value) {
	a.values = append(a.values, value)
}

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

type Field struct {
	key   string
	value Value
}

func NewField(key string, value Value) Field {
	return Field{key: key, value: value}
}

func (f Field) render(depth int) string {
	return fmt.Sprintf(`"%s": %s`, f.key, f.value.Render(depth))
}

type Object struct {
	fields []Field
}

func NewObject(fields ...Field) Object {
	return Object{fields: fields}
}

func (o *Object) AddFields(fields ...Field) {
	o.fields = append(o.fields, fields...)
}

func (o Object) Render(depth int) string {
	if len(o.fields) == 0 {
		return "{}"
	}

	indent := strings.Repeat(indentToken, depth+1)
	fields := []string{}
	for _, field := range o.fields {
		f := field.render(depth + 1)
		f = indent + f
		fields = append(fields, f)
	}

	indent = strings.Repeat(indentToken, depth)
	stackedFields := strings.Join(fields, ",\n") + "\n"
	return "{\n" + stackedFields + indent + "}"
}

type Null struct{}

func (n Null) Render(depth int) string {
	return "null"
}

type String string

func (s String) Render(depth int) string {
	return fmt.Sprintf(`"%s"`, s)
}

type Int int

func (i Int) Render(depth int) string {
	return fmt.Sprintf("%d", i)
}
