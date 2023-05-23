package errors

import (
	"fmt"

	"github.com/bjatkin/yok/parse"
)

type Code struct {
	Num   int
	Title string
}

type Anno struct {
	// TODO: swap this out for a the token type once it's introduced
	Token parse.Node
	Node  string
}

type User struct {
	Warn      bool
	ErrorCode Code
	// This can be an abitrary structure to help support more structured summaries
	Summary fmt.Stringer
	Annos   []Anno
}

// Example User Errors
//
// ERR: 1020 - Incompatable Types
// [ src/test.yk:103:32 ]
// > assignment of in to string variable 'a' is not allowed
//     |
//  90 | let a string
//     |* a is declared here as a string
//     |
// 103 | a = 14
//     |* a is assigned the value 14 here which is an int
//

// WARN: 1203 - Unused Identifyer
// [ src/again.yk:23:2 ]
// > variable 'b' is assigned but the value is never read
//   |
// 1 | b = 25
//   |* b is declared here
//   |
//   |* the value of b is never read once it's declared

type Internal struct {
}

// Example Internal Errors
