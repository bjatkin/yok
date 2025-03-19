package yokast

// This file contains yokast.Nodes that are not part of the yok spec directly but are in fact
// used to help map yok more cleanly to sh. We could turn this into a full IR package but it's
// pretty small so I'm keeping it a part of this package for now.

type NestedCall struct {
	Expr
	Depth int
	Call  *Call
}
