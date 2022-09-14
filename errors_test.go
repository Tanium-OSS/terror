package terror

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/multierr"
)

// When running tests, we'll probably not have the -trimpath arg specified (for
// example, when clicking the "run test" in vscode). We want to ensure that the
// tests produce deterministic output, though! So we'll be sneaky: Before any of
// the tests run, we'll capture a local stack trace and use that to determine
// the root of the repo and trim that off of our stack traces.
func init() {
	// path/to/repo/terror/errors_test.go or pkg.com/repo/terror/errors_test.go
	thisFile := capture(0).File
	// path/to/repo/terror or pkg.com/repo/terror
	thisDir := filepath.Dir(thisFile)
	// path/to/repo/ or pkg.com/repo/ (with trailing / or \)
	repoRoot := strings.TrimSuffix(thisDir, "terror")
	// Now configure our filename cleaning to remove the repo root. This works
	// since we're only ever dealing with files in this dir for this test.
	CleanFileName = func(filename string) string {
		return strings.TrimPrefix(filename, repoRoot)
	}
}

func TestWrapNil(t *testing.T) {
	assert.Nil(t, Wrap(nil, ""))
	assert.Nil(t, Wrap(nil, "something"))
	assert.Nil(t, Wrap(nil, "something %s", "hi"))
}

func TestWrap(t *testing.T) {
	err := tryButFail()
	assert.Equal(t, "trying something: some error", err.Error())
	assert.Equal(t,
		""+
			"trying something\n"+
			" --- at terror/testdata_test.go:13 (tryButFail) ---\n"+
			"caused by some error",
		fmt.Sprintf("%+v", err),
	)

	err = wrapNoMessage(err)
	err = wrapMessage(err, "wrap with %d message(s)", 2)
	assert.Equal(t, "wrap with 2 message(s): trying something: some error", err.Error())
	assert.Equal(t,
		""+
			"wrap with 2 message(s)\n"+
			" --- at terror/testdata_test.go:18 (wrapMessage) ---\n"+
			" --- at terror/testdata_test.go:16 (wrapNoMessage) ---\n"+
			"caused by trying something\n"+
			" --- at terror/testdata_test.go:13 (tryButFail) ---\n"+
			"caused by some error",
		fmt.Sprintf("%+v", err),
	)

	err = wrapNoMessage(err)
	assert.Equal(t, "wrap with 2 message(s): trying something: some error", err.Error())
	assert.Equal(t,
		""+
			" --- at terror/testdata_test.go:16 (wrapNoMessage) ---\n"+
			"caused by wrap with 2 message(s)\n"+
			" --- at terror/testdata_test.go:18 (wrapMessage) ---\n"+
			" --- at terror/testdata_test.go:16 (wrapNoMessage) ---\n"+
			"caused by trying something\n"+
			" --- at terror/testdata_test.go:13 (tryButFail) ---\n"+
			"caused by some error",
		fmt.Sprintf("%+v", err),
	)

	assert.Equal(t,
		"wrap with 2 message(s): trying something: some error",
		err.Error(),
	)
	assert.Equal(t,
		""+
			" --- at terror/testdata_test.go:16 (wrapNoMessage) ---\n"+
			"caused by wrap with 2 message(s)\n"+
			" --- at terror/testdata_test.go:18 (wrapMessage) ---\n"+
			" --- at terror/testdata_test.go:16 (wrapNoMessage) ---\n"+
			"caused by trying something\n"+
			" --- at terror/testdata_test.go:13 (tryButFail) ---\n"+
			"caused by some error",
		fmt.Sprint(err),
	)
	assert.Equal(t,
		"wrap with 2 message(s): trying something: some error",
		fmt.Sprintf("%s", err),
	)
	assert.Equal(t,
		""+
			" --- at terror/testdata_test.go:16 (wrapNoMessage) ---\n"+
			"caused by wrap with 2 message(s)\n"+
			" --- at terror/testdata_test.go:18 (wrapMessage) ---\n"+
			" --- at terror/testdata_test.go:16 (wrapNoMessage) ---\n"+
			"caused by trying something\n"+
			" --- at terror/testdata_test.go:13 (tryButFail) ---\n"+
			"caused by some error",
		fmt.Sprintf("%v", err),
	)
	assert.Equal(t,
		""+
			" --- at terror/testdata_test.go:16 (wrapNoMessage) ---\n"+
			"caused by wrap with 2 message(s)\n"+
			" --- at terror/testdata_test.go:18 (wrapMessage) ---\n"+
			" --- at terror/testdata_test.go:16 (wrapNoMessage) ---\n"+
			"caused by trying something\n"+
			" --- at terror/testdata_test.go:13 (tryButFail) ---\n"+
			"caused by some error",
		fmt.Sprintf("%#v", err),
	)
	assert.Equal(t,
		"wrap with 2 message(s): trying something: some error",
		fmt.Sprintf("%+s", err),
	)
}

func TestWrapWithFormat(t *testing.T) {
	err := tryButFailf("something")
	assert.Equal(t, "trying something: some error", err.Error())
	assert.Equal(t,
		""+
			"trying something\n"+
			" --- at terror/testdata_test.go:14 (tryButFailf) ---\n"+
			"caused by some error",
		fmt.Sprintf("%+v", err),
	)
}

func TestWrap_MethodNames(t *testing.T) {
	var sometype SomeType
	err := sometype.wrap(errSentinel, "checking function name")
	err = sometype.wrapPtr(err, "checking pointer receiver")
	assert.Equal(t,
		""+
			"checking pointer receiver\n"+
			" --- at terror/testdata_test.go:27 (SomeType.wrapPtr) ---\n"+
			"caused by checking function name\n"+
			" --- at terror/testdata_test.go:24 (SomeType.wrap) ---\n"+
			"caused by some error",
		fmt.Sprintf("%+v", err),
	)
}

func TestNew(t *testing.T) {
	err := newErr("adjusting %d things", 12)
	assert.EqualError(t, err, "adjusting 12 things")
	assert.Equal(t,
		""+
			"adjusting 12 things\n"+
			" --- at terror/testdata_test.go:30 (newErr) ---",
		fmt.Sprintf("%+v", err),
	)
}

func TestGo113(t *testing.T) {
	err := wrapMessage(SomeCustomErrorWrapper{errSentinel}, "wrapped in terror")
	assert.True(t, errors.Is(err, errSentinel))
	assert.NotSame(t, err, errSentinel)

	assert.Equal(t, "wrapped in terror: some error", err.Error())

	var unwrapped SomeCustomErrorWrapper
	assert.True(t, errors.As(err, &unwrapped))

	// If you use fmt.Errorf, you commit to the detailed version of the error
	// because fmt.Errorf immediately formats the error with the %v. Subsequent
	// formatting directives have no effect on the wrapped terror error.
	err = fmt.Errorf("destroyed formatting: %w", err)
	const preFormattedErrorString = "" +
		"destroyed formatting: wrapped in terror\n" +
		" --- at terror/testdata_test.go:18 (wrapMessage) ---\n" +
		"caused by some error"
	assert.Equal(t, preFormattedErrorString, err.Error())
	assert.Equal(t, preFormattedErrorString, fmt.Sprintf("%s", err))
	assert.Equal(t, preFormattedErrorString, fmt.Sprintf("%v", err))
	assert.Equal(t, preFormattedErrorString, fmt.Sprintf("%+v", err))

	// ...but at least the following still work!
	assert.True(t, errors.Is(err, errSentinel))
	assert.True(t, errors.As(err, &unwrapped))
	assert.NotSame(t, err, errSentinel)

	// including this:
	var ourError TError
	assert.True(t, errors.As(err, &ourError))
	assert.Equal(t,
		""+
			"wrapped in terror\n"+
			" --- at terror/testdata_test.go:18 (wrapMessage) ---\n"+
			"caused by some error",
		fmt.Sprintf("%+v", ourError),
	)
}

type SomeCustomErrorWrapper struct{ error }

func (e SomeCustomErrorWrapper) Unwrap() error { return e.error }
func (e SomeCustomErrorWrapper) Error() string { return e.error.Error() }

func TestCode(t *testing.T) {
	// No code
	assert.Equal(t, 0, GetCode(nil))
	assert.Equal(t, 0, GetCode(New("foo")))
	assert.Equal(t, 0, GetCode(Wrap(errSentinel, "foo")))

	// single code, one layer
	assert.Equal(t, 1, GetCode(NewWithCode(1, "foo")))
	assert.Equal(t, 2, GetCode(WrapWithCode(errSentinel, 2, "foo")))

	// two layers
	assert.Equal(t, 3, GetCode(Wrap(NewWithCode(3, "foo"), "bar")))
	assert.Equal(t, 4, GetCode(Wrap(WrapWithCode(errSentinel, 4, "foo"), "bar")))

	// two layers but with an externally-wrapped error.
	assert.Equal(t, 5, GetCode(fmt.Errorf("blah %w", NewWithCode(5, "foo"))))
	assert.Equal(t, 6, GetCode(fmt.Errorf("blah %w", WrapWithCode(errSentinel, 6, "foo"))))

	// second layer updates the code
	assert.Equal(t, 7, GetCode(WrapWithCode(NewWithCode(1, "foo"), 7, "bar")))
	assert.Equal(t, 8, GetCode(WrapWithCode(WrapWithCode(errSentinel, 2, "foo"), 8, "bar")))

	// it's possible to update the code to be zero.
	assert.Equal(t, 0, GetCode(WrapWithCode(NewWithCode(1, "foo"), 0, "bar")))
	assert.Equal(t, 0, GetCode(WrapWithCode(WrapWithCode(errSentinel, 2, "foo"), 0, "bar")))

	// multierr picks the first error in the chain
	assert.Equal(t, 11, GetCode(multierr.Combine(NewWithCode(11, "foo"), NewWithCode(12, "foo"))))
}

func TestPanicStacks(t *testing.T) {
	err := panicError()
	assert.EqualError(t, err, "panic: error")
	assert.Equal(t,
		""+
			"panic\n"+
			" --- at terror/testdata_test.go:34 (panicError.func1) ---\n"+
			"caused by error\n"+
			" --- at terror/testdata_test.go:36 (panicError) ---",
		fmt.Sprintf("%+v", err),
	)
}

func TestRootError(t *testing.T) {
	base := fmt.Errorf("some error")
	firstWrap := Wrap(base, "the first wrap")
	wrapped := firstWrap
	for i := 0; i < 3; i++ {
		wrapped = Wrap(wrapped, "oh no an error")
	}

	assert.Equal(t, nil, RootError(base))
	assert.ErrorIs(t, firstWrap, RootError(firstWrap))

	rl := RootError(wrapped)
	assert.NotErrorIs(t, firstWrap, wrapped)
	assert.ErrorIs(t, firstWrap, rl)

	r, err := regexp.Compile(`errors_test\.go:\d+`)
	assert.NoError(t, err)
	assert.True(t, r.MatchString(rl.Location().String()))
}
