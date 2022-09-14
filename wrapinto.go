package terror

import (
	"fmt"
)

// WrapInto annotates the provided error with the file and line of the call along with the provided
// message. The format and args are formatted printf style. If * pErr is nil, WrapInto is a no-op.
//
// The intended use is to use defer with a named return variable to provide additional context. If
// deferred, it will capture the line of the return statement, not the line of the defer statement.
func WrapInto(pErr *error, format string, args ...interface{}) {
	if *pErr == nil {
		return
	}
	*pErr = TError{
		base: *pErr,
		msg:  fmt.Sprintf(format, args...),
		loc:  capture(1),
	}
}
