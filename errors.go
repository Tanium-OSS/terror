package terror

import (
	"errors"
	"fmt"
	"io"
	"runtime"
	"strings"
)

// CleanFileName is a process global hook that enables sanitizing filenames in
// stack traces.
var CleanFileName = func(filename string) string {
	// Pass filename through by default.
	return filename
}

// Wrap annotates the provided error with the file and line of the call along
// with the provided message. The format and args are formatted printf style.
// If err is nil, Wrap returns nil.
func Wrap(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	return TError{
		base: err,
		msg:  fmt.Sprintf(format, args...),
		loc:  capture(1),
	}
}

// Annotate annotates the provided error with the file and line of the call. If err is nil, Annotate returns nil.
func Annotate(err error) error {
	if err == nil {
		return nil
	}
	return TError{
		base: err,
		loc:  capture(1),
	}
}

// WrapWithCode annotates the provided error with the file and line of the call
// along with the provided message and a specified error code. The error code
// can be retrieved using GetCode(). The format and args are formatted printf
// style. If err is nil, Wrap returns nil.
func WrapWithCode(err error, code int, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	return codeError{TError{
		base: err,
		msg:  fmt.Sprintf(format, args...),
		loc:  capture(1),
	}, code}
}

// New creates an error with the specified message as well as the location of
// this call. This is a drop-in replacement for fmt.Errorf.
func New(format string, args ...interface{}) error {
	return TError{
		base: nil,
		msg:  fmt.Sprintf(format, args...),
		loc:  capture(1),
	}
}

// NewWithCode creates an error with the specified message as well as the
// location of this call and a specified error code. The error code can be
// retrieved using GetCode().
func NewWithCode(code int, format string, args ...interface{}) error {
	return codeError{TError{
		base: nil,
		msg:  fmt.Sprintf(format, args...),
		loc:  capture(1),
	}, code}
}

// GetCode returns the error code most recently added via WrapWithCode or
// NewWithCode. If no error code is present or the error is not created by this
// package, 0 is returned.
func GetCode(err error) int {
	var wrapper codeError
	if !errors.As(err, &wrapper) {
		return 0
	}
	return wrapper.code
}

// codeError adds an error code to an error. base may not be nil.
type codeError struct {
	base error
	code int
}

func (e codeError) Unwrap() error { return e.base }
func (e codeError) Error() string { return e.base.Error() }
func (e codeError) Format(f fmt.State, c rune) {
	if base, ok := e.base.(fmt.Formatter); ok { //nolint:errorlint
		base.Format(f, c)
		return
	}
	fmt.Fprintf(f, origFormatString(f, c), e.base)
}

// TError is the wrapped error implementation.
type TError struct {
	base error
	msg  string
	loc  Location
}

// Unwrap returns the base error, implementing the go1.13 error unwrapping to
// support errors.Is and errors.As.
func (e TError) Unwrap() error { return e.base }

// Error returns the simple, compact, one-line error format.
func (e TError) Error() string {
	// There are three interesting cases here:
	// 1: root error, created by New(): have message but no base error
	//    --> just print the message.
	if e.base == nil {
		return e.msg
	}
	// 2: only a stack marker: have a base error but no message
	//    --> just print the base error
	if e.msg == "" {
		return e.base.Error()
	}
	// 3: wrapped with message: have a base error AND a message
	//    --> print `${msg}: ${base}`
	return e.msg + ": " + e.base.Error()
}

// Location returns a string representation of the tError after cleaning file and function names.
func (e TError) Location() Location {
	return e.loc
}

// detailedError returns the multiline stacktrace-annotated error. This will
// format the wrapped error recursively.
func (e TError) detailedError(w io.Writer) {
	sep := ""
	if e.msg != "" {
		fmt.Fprintf(w, "%s", e.msg)
		sep = "\n"
	}
	if e.loc != (Location{}) {
		io.WriteString(w, sep)
		fmt.Fprintf(w, " --- at %s ---", e.Location().String())
		sep = "\n"
	}
	if ourError, ok := e.base.(TError); ok && ourError.msg == "" { //nolint:errorlint
		io.WriteString(w, sep)
		fmt.Fprintf(w, "%+v", e.base)
	} else if e.base != nil {
		io.WriteString(w, sep)
		if sep != "" {
			fmt.Fprint(w, "caused by ")
		}
		fmt.Fprintf(w, "%+v", e.base)
	}
}

// Format implements fmt.Formatter so that we know when we're being formatted by
// a Printf-style func. This detects when we're being printed specifically by
// "%v" in which case we output the detailed stack trace.
func (e TError) Format(f fmt.State, c rune) {
	if c == 'v' {
		e.detailedError(f)
	} else {
		io.WriteString(f, e.Error())
	}
}

// Error represents the terror-wrapped error which includes additional information such as the stack
// trace.
type Error interface {
	error
	Location() Location
}

var _ Error = TError{}

// RootError returns the innermost terror-wrapped Error. If this error does not contain any
// terror-wrapped error, nil will be returned. This is contrasted with calling errors.As(...) which
// returns the outermost layer.
func RootError(e error) Error {
	// This will be nil if e doesn't implement Error.
	deepestError, _ := e.(Error) //nolint:errorlint
	for err := errors.Unwrap(e); err != nil; err = errors.Unwrap(err) {
		withLoc, ok := err.(Error) //nolint:errorlint
		if ok {
			deepestError = withLoc
		}
	}
	return deepestError
}

// Adapted from github.com/palantir/stacktrace, this removes the package name
// from the function name:
//
//	github.com/foo/bar/package.FuncName                  --> FuncName
//	github.com/foo/bar/package.Receiver.MethodName       --> Receiver.MethodName
//	github.com/foo/bar/package.(*PtrReceiver).MethodName --> PtrReceiver.MethodName
func cleanFuncName(longName string) string {
	withoutPath := longName[strings.LastIndex(longName, "/")+1:]
	withoutPackage := withoutPath[strings.Index(withoutPath, ".")+1:]
	shortName := withoutPackage
	if shortName[0] == '(' {
		pos := strings.IndexByte(shortName, ')')
		return shortName[2:pos] + shortName[pos+1:]
	}
	return shortName
}

// Location describes a single stack frame position.
type Location struct {
	File     string
	Line     int
	Function string
}

// String returns a string representation of the Location.
func (l Location) String() string {
	return fmt.Sprintf("%s:%d (%s)", l.File, l.Line, l.Function)
}

// capture reads the current stack pos.
func capture(skip int) Location {
	// For reading a single stack frame at a time, the paired usage of
	// runtime.Caller + runtime.FuncForPC is fine to use and reasonably efficient.
	// If you want more than one frame (e.g. pulling a stack), it's better to use
	// runtime.Callers + runtime.CallersFrames.
	//
	// For efficiency in the future: it's probably better to pull entire stacks
	// via runtime.Callers and defer resolving the file/line/function information
	// until we're printing the error out. However, runtime.Callers seems a little
	// wonky when reading only a single stack frame.
	pc, file, line, ok := runtime.Caller(skip + 1)
	if !ok {
		return Location{}
	}
	function := runtime.FuncForPC(pc).Name()
	return Location{CleanFileName(file), line, cleanFuncName(function)}
}

// Adapted from github.com/palantir/stacktrace, this reconstructs the format
// string from the Format() args.
func origFormatString(f fmt.State, c rune) string {
	formatString := "%"
	// keep the flags recognized by fmt package
	for _, flag := range "-+# 0" {
		if f.Flag(int(flag)) {
			formatString += string(flag)
		}
	}
	if width, has := f.Width(); has {
		formatString += fmt.Sprint(width)
	}
	if precision, has := f.Precision(); has {
		formatString += "."
		formatString += fmt.Sprint(precision)
	}
	formatString += string(c)
	return formatString
}
