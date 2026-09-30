package panicutil_test

import (
	"fmt"
	"io"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/go-pnp/go-pnp/pkg/panicutil"
)

func TestNewErrorHidesErrorPanicValuesFromErrorChecks(t *testing.T) {
	err := panicutil.NewError(io.EOF)

	require.NotErrorIs(t, err, io.EOF)
	assert.Equal(t, "panic: EOF", err.Error())
}

func TestNewErrorDescribesNonErrorPanicValues(t *testing.T) {
	err := panicutil.NewError(42)

	assert.Equal(t, "panic: 42", err.Error())
}

func TestCallReturnsPanicAsErrorWithStackOfPanickingFunction(t *testing.T) {
	err := panicutil.Call(panickingFunction)

	require.EqualError(t, err, "panic: boom")
	assert.Contains(t, fmt.Sprintf("%+v", err), "panicutil_test.panickingFunction")
}

func TestCallReturnsRuntimeErrorPanicWithStackOfPanickingFunction(t *testing.T) {
	err := panicutil.Call(outOfRangeIndexingFunction)

	require.ErrorContains(t, err, "panic: runtime error: index out of range")
	assert.Contains(t, fmt.Sprintf("%+v", err), "panicutil_test.outOfRangeIndexingFunction")
}

func TestCallReturnsErrorOfFunction(t *testing.T) {
	err := panicutil.Call(func() error { return io.EOF })

	require.ErrorIs(t, err, io.EOF)
}

func TestCallReturnsNilWhenFunctionSucceeds(t *testing.T) {
	err := panicutil.Call(func() error { return nil })

	require.NoError(t, err)
}

func TestRecoverHandsPanicOfDeferringGoroutineToHandler(t *testing.T) {
	var (
		handledError error
		waitGroup    sync.WaitGroup
	)

	waitGroup.Add(1)

	go func() {
		defer waitGroup.Done()
		defer panicutil.Recover(func(err error) { handledError = err })

		panic("boom")
	}()

	waitGroup.Wait()

	require.EqualError(t, handledError, "panic: boom")
}

func TestRecoverDoesNotCallHandlerWithoutPanic(t *testing.T) {
	handlerCalled := false

	func() {
		defer panicutil.Recover(func(error) { handlerCalled = true })
	}()

	assert.False(t, handlerCalled)
}

func panickingFunction() error {
	panic("boom")
}

func outOfRangeIndexingFunction() error {
	var values []int

	index := len(values)
	values[index]++

	return nil
}
