package errors

// TODO: we need to build better error types here
// good error handling is going to be absolutely essential for this to be a good tool

// TODO: there are two classes of errors that are showing up in the code so far:
//  1) errors that really should never happen (e.g. missing encoding functionality for a given node type)
//  2) errors that are the fault of the user (e.g. syntax errors)
// We probably need two different methods for handling these errors.
// We could also just panic a lot more often but I don't really like that

type Err struct {
	msg string
}

func (e Err) Error() string {
	return e.msg
}

func New(msg string) Err {
	return Err{msg}
}
