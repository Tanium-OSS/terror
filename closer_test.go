package terror

import (
	"errors"
	"fmt"
	"testing"

	"go.uber.org/multierr"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ExampleCloser returns an error on use and on close.
type ExampleCloser struct{}

// Close always returns an error.
func (c *ExampleCloser) Close() error {
	return errors.New("could not close")
}

// Use always returns an error.
func (c *ExampleCloser) Use() error {
	return errors.New("could not use")
}

func allocateExampleResource() *ExampleCloser {
	return &ExampleCloser{}
}

func ExampleCloseAndAppendOnError() {
	errExample := func() (err error) {
		file := allocateExampleResource()
		defer CloseAndAppendOnError(&err, file, "close file")

		if err := file.Use(); err != nil {
			return Wrap(err, "first use")
		}
		if err := file.Use(); err != nil {
			return Wrap(err, "second use")
		}
		return nil
	}()

	fmt.Println(errExample)
	// Output: first use: could not use; close file: could not close
}

// TestCloser will return an error on closing if we want it to.
type TestCloser struct {
	errorOnClose error
	closeCount   int
	t            *testing.T
}

// Close returns an error if we want it to.
func (c *TestCloser) Close() error {
	c.closeCount++
	c.t.Log("TestCloser.Close", "closeCount", c.closeCount)
	return c.errorOnClose
}

// CloseCount returns the number of times Close() has been called.
func (c TestCloser) CloseCount() int {
	c.t.Log("TestCloser.CloseCount", "closeCount", c.closeCount)
	return c.closeCount
}

func TestCloseAndAppendOnError_NoError_GoodCloser(t *testing.T) {
	t.Parallel()
	c := &TestCloser{t: t}
	var err error
	CloseAndAppendOnError(&err, c, "")
	assert.Equal(t, 1, c.closeCount)
	require.Nil(t, err)
}

func TestCloseAndAppendOnError_NoError_BadCloser(t *testing.T) {
	t.Parallel()
	c := &TestCloser{t: t, errorOnClose: errors.New("badness")}
	var err error
	CloseAndAppendOnError(&err, c, "")
	assert.Equal(t, 1, c.closeCount)
	require.Error(t, err)
	require.Contains(t, err.Error(), "badness")
}

func TestCloseAndAppendOnError_Error_GoodCloser(t *testing.T) {
	t.Parallel()
	c := &TestCloser{t: t}
	err := errors.New("sadness")
	CloseAndAppendOnError(&err, c, "")
	assert.Equal(t, 1, c.closeCount)
	require.Error(t, err)
	require.Contains(t, err.Error(), "sadness")
}

func TestCloseAndAppendOnError_Error_BadCloser(t *testing.T) {
	t.Parallel()
	c := &TestCloser{t: t, errorOnClose: errors.New("badness")}
	err := errors.New("sadness")
	CloseAndAppendOnError(&err, c, "")
	assert.Equal(t, 1, c.closeCount)
	require.Error(t, err)
	require.Contains(t, err.Error(), "badness")
	require.Contains(t, err.Error(), "sadness")
}

func TestCloseAndLogOnError(t *testing.T) {
	errMessage1 := "badness"
	errMessage2 := "badness2"
	cBad := &TestCloser{t: t, errorOnClose: errors.New(errMessage1)}
	cBad2 := &TestCloser{t: t, errorOnClose: errors.New(errMessage2)}
	cGood := &TestCloser{t: t}

	var err error
	logFn := func(template string, args ...interface{}) {
		err = args[0].(error)
	}

	CloseAndLogOnError(logFn, cBad, cBad2, cGood)
	assert.EqualError(t, err, multierr.Combine(New(errMessage1), New(errMessage2)).Error())
}
