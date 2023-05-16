package sym

import "fmt"

type ID int64

type Symbol struct {
	Value string
	Alias string
	Type  YokType
	File  string
	Line  int
	Col   int
}

type Table struct {
	symbols []*Symbol
}

func NewTable() *Table {
	return &Table{}
}

func (t *Table) AddSymbol(s *Symbol) ID {
	t.symbols = append(t.symbols, s)
	return ID(len(t.symbols) - 1)
}

func (t *Table) GetSymbol(id ID) (*Symbol, error) {
	if id < 0 || int(id) >= len(t.symbols) {
		return nil, fmt.Errorf("invalid id %d", id)
	}
	return t.symbols[id], nil
}

func (t *Table) MustGetSymbol(id ID) *Symbol {
	if id < 0 || int(id) >= len(t.symbols) {
		panic(fmt.Errorf("invalid id %d", id))
	}
	return t.symbols[id]
}
