package terror

import "errors"

// This file contains a variety of functions that exercise the wrapping
// functionality. Care should be taken when modifying this file (including
// adding/removing imports) because many of the tests assume specific line
// numbers in this file.

var errSentinel = errors.New("some error")

func failed() error                  { return errSentinel }
func tryButFail() error              { return Wrap(failed(), "trying something") }
func tryButFailf(param string) error { return Wrap(failed(), "trying %s", param) }

func wrapNoMessage(err error) error { return Wrap(err, "") }
func wrapMessage(err error, msg string, args ...interface{}) error {
	return Wrap(err, msg, args...)
}

type SomeType struct{}

func (t SomeType) wrap(err error, msg string, args ...interface{}) error {
	return Wrap(err, msg, args...)
}
func (t *SomeType) wrapPtr(err error, msg string, args ...interface{}) error {
	return Wrap(err, msg, args...)
}

func newErr(format string, args ...interface{}) error { return New(format, args...) }

func panicError() (err error) {
	defer func() {
		err = Wrap(recover().(error), "panic")
	}()
	panic(New("error"))
}
