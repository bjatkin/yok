package sym

import (
	"fmt"
	"regexp"
)

var (
	stringTypeReg = regexp.MustCompile(`^".*"$`)
	intTypeReg    = regexp.MustCompile(`^[0-9]+$`)
	boolTypeReg   = regexp.MustCompile(`^(true|false)$`)
	pathTypeReg   = regexp.MustCompile(`(\.|\.\.|~){0,1}\/[^ \(\)\[\]\{\}\n\r]+`)
)

// TODO: again, this should proabbly be an int but debugging that is a pain so
// I'm leaving this as a string for now.
type YokType string

const (
	UnknownType = YokType("")
	StringType  = YokType("string")
	IntType     = YokType("int")
	BoolType    = YokType("bool")
	PathType    = YokType("path")
)

func StrToType(t string) YokType {
	switch t {
	case string(StringType):
		return StringType
	case string(IntType):
		return IntType
	case string(BoolType):
		return BoolType
	case string(PathType):
		return PathType
	default:
		// TODO: this should be an error instead of a panic
		fmt.Println("paicigin:", t)
		panic(1)
	}
}

func TypeFromValue(value string) YokType {
	switch {
	case stringTypeReg.MatchString(value):
		return StringType
	case intTypeReg.MatchString(value):
		return IntType
	case boolTypeReg.MatchString(value):
		return BoolType
	case pathTypeReg.MatchString(value):
		return PathType
	default:
		// TODO: this should be an error instead of a panic
		panic(1)
	}
}

func DefaultValue(yokType YokType) string {
	switch yokType {
	case StringType:
		return `""`
	case IntType:
		return "0"
	case BoolType:
		return "false"
	case PathType:
		return "./"
	default:
		// TODO: this should be an error instead of a panic
		panic(1)
	}
}
