package terror

import (
	"io"

	"go.uber.org/multierr"
)

// CloseAndAppendOnError is a helper function that makes it easier to close Closers in defer
// statements and still capture any errors that arise during closing. Use this function when a
// failure to Close() indicates that something has gone wrong and it should not be ignored. For
// example, when writing to a file, a failure to Close() the file may indicate that not all of the
// bytes were in fact written, even if the call to Write() did not return an error. This should
// always be used with named return variables. See https://play.golang.org/p/ECMc6EfWXxt for
// examples of what can go wrong if a local variable is used instead.
func CloseAndAppendOnError(pErr *error, c io.Closer, format string, args ...interface{}) {
	multierr.AppendInto(pErr, Wrap(c.Close(), format, args...))
}

// CloseAndLogOnError is a helper function that makes it easier to close Closers in defer
// statements and still log any errors that arise during closing. Use this function when closing
// closers where a failure to Close() can happen and the errors cannot be handled.
func CloseAndLogOnError(fn func(template string, args ...interface{}), closers ...io.Closer) {
	var err error
	for _, closer := range closers {
		multierr.AppendInto(&err, closer.Close())
	}
	if err != nil {
		fn("Close operation failed with error: %+v", err)
	}
}
