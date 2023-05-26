package ast

import "fmt"

type validator interface {
	visitor
	errors() []string
}

// TODO: check for unused imports
type validateUse struct {
	useBlocks        int
	visited          int
	useIsNotFirst    bool
	duplicateImports []string
	unknownCommand   []string
	imported         map[string]int
}

func newValidateuse() *validateUse {
	return &validateUse{
		imported: map[string]int{
			"echo": 1, // echo is a built in
		},
	}
}

func (v *validateUse) visit(node Node) visitor {
	switch t := node.(type) {
	case Comment:
		return nil
	case NewLine:
		return nil
	case Root:
		return v
	case Use:
		if v.visited > 0 {
			v.useIsNotFirst = true
		}
		v.useBlocks++
		return v
	case Import:
		name := t.Alias
		if name == "" {
			name = t.CmdName
		}
		if name == "" {
			name = t.Path
		}
		if _, ok := v.imported[name]; ok {
			v.duplicateImports = append(v.duplicateImports, name)
		}
		v.imported[name] = 0
		return nil
	case Command:
		v.visited++
		if _, ok := v.imported[t.Identifyer]; !ok {
			v.unknownCommand = append(v.unknownCommand, t.Identifyer)
		}
		v.imported[t.Identifyer]++
		return v
	default:
		v.visited++
		return v
	}
}

func (v *validateUse) errors() []string {
	var errs []string
	if v.useIsNotFirst {
		errs = append(errs, "use import must be the first command in a yok script")
	}

	for _, imports := range v.duplicateImports {
		errs = append(errs, fmt.Sprintf("duplicate import %s", imports))
	}

	for _, cmd := range v.unknownCommand {
		errs = append(errs, fmt.Sprintf("unknown command %s", cmd))
	}

	for name, count := range v.imported {
		if count == 0 {
			errs = append(errs, fmt.Sprintf("unused command %s", name))
		}
	}

	return errs
}
