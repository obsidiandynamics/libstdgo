/*
Package check contains assertions to assist with unit testing.
*/
package check

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/stretchr/testify/assert"
)

// Tester is an API-compatible stand-in for *testing.T.
type Tester interface {
	Errorf(format string, args ...interface{})
}

// PrintStack prints the call stack, starting from the given depth.
func PrintStack(depth int) string {
	var str strings.Builder
	for i, s := range assert.CallerInfo()[depth:] {
		str.WriteString("\n")
		if i == 0 {
			str.WriteString("> ")
		} else {
			str.WriteString("  ")
		}
		str.WriteString(s)
	}
	return str.String()
}

// ErrFault is a pre-canned error, useful in simulating faults.
var ErrFault = errors.New("Simulated")

// PanicAssertion checks a given panic cause. It is used by ThatPanicsAsExpected.
type PanicAssertion func(t Tester, cause interface{})

// ThatPanicsAsExpected checks that the given function f panics, where the trapped panic complies
// with the supplied assertion.
func ThatPanicsAsExpected(t Tester, assertion PanicAssertion, f func()) {
	defer func() {
		if cause := recover(); cause != nil {
			assertion(t, cause)
		}
	}()
	f()
	assert.Fail(t, "Did not panic as expected")
}

// ThatDoesNotPanic ensures that the given function f returns without panicking. This is useful in tests that
// must perform multiple assertions without terminating the test. (Otherwise, if we let the panic through, the
// test will not run to completion.)
func ThatDoesNotPanic(t Tester, f func()) {
	defer func() {
		if panic := recover(); panic != nil {
			assert.Fail(t, fmt.Sprintf("Unexpected panic: %v", panic))
		}
	}()
	f()
}

// AnyCause satisfies the panic assertion irrespective of the cause. Use to test that a function panics without
// caring as to the nature of the panic.
func AnyCause() PanicAssertion {
	return func(_ Tester, _ interface{}) {}
}

// ErrorWithValue checks that the panic is of the built-in error type, where the result of calling err.Error()
// matches the given value.
func ErrorWithValue(value string) PanicAssertion {
	return func(t Tester, cause interface{}) {
		err, ok := cause.(error)
		if !ok {
			assert.Fail(t, fmt.Sprintf("Expected error, got %T", cause))
			return
		}
		assert.Equal(t, value, err.Error())
	}
}

// ErrorContaining checks that the panic is of the built-in error type, where the result of calling err.Error()
// contains the given substring.
func ErrorContaining(substr string) PanicAssertion {
	return func(t Tester, cause interface{}) {
		err, ok := cause.(error)
		if !ok {
			assert.Fail(t, fmt.Sprintf("Expected error, got %T", cause))
			return
		}
		assert.Contains(t, err.Error(), substr)
	}
}

// CauseEqual checks that the panic is equal to the given cause.
func CauseEqual(expected interface{}) PanicAssertion {
	return func(t Tester, cause interface{}) {
		assert.Equal(t, expected, cause)
	}
}

// Timesert provides a mechanism for awaiting an assertion or a condition from a test.
type Timesert interface {
	Until(p Predicate) bool
	UntilAsserted(a Assertion) bool
}

type timesert struct {
	t        Tester
	timeout  time.Duration
	interval time.Duration
}

// DefaultWaitCheckInterval is the default value of the optional check interval
// passed to Wait().
const DefaultWaitCheckInterval = 1 * time.Millisecond

// Wait returns a Timesert object will block for up to the given timeout.
// The third argument is optional, specifying the upper bound on the check interval
// (defaults to DefaultWaitCheckInterval).
func Wait(t Tester, timeout time.Duration, interval ...time.Duration) Timesert {
	checkInterval := DefaultWaitCheckInterval
	switch {
	case len(interval) > 1:
		panic(fmt.Errorf("Argument list too long"))
	case len(interval) == 1:
		checkInterval = interval[0]
	}
	return &timesert{t: t, timeout: timeout, interval: checkInterval}
}

// Predicate is a condition that must be satisfied for Timesert.Until to return.
type Predicate func() bool

// Not inverts a given predicate.
func Not(p Predicate) Predicate {
	return func() bool {
		return !p()
	}
}

// Assertion is a check that must pass for Timesert.UntilAsserted to return without raising
// an error.
type Assertion func(t Tester)

// Waits until the given predicate is met, up to the timeout configured in the Timesert. Returns
// the final response of the predicate (true if satisfied).
func (ts *timesert) Until(p Predicate) bool {
	return ts.untilAsserted(func(t Tester) {
		if !p() {
			assert.Fail(t, "Condition not met")
		}
	})
}

// Equal tests that the value returned by the given supplied matches the expected value.
func Equal(supplier func() interface{}, expected interface{}) Predicate {
	return func() bool {
		return supplier() == expected
	}
}

// Waits until the given assertion is satisfied, up to the timeout configured in the Timesert, returning
// the outcome of the assertion (true if passed). Any
// errors reported while the assertion isn't met are captured. If the assertion is satisfied within the
// timeout period, these errors are discarded; otherwise, they are reported back to the Tester.
func (ts *timesert) UntilAsserted(a Assertion) bool {
	return ts.untilAsserted(a)
}

func (ts *timesert) untilAsserted(a Assertion) bool {
	var intervalTicker *time.Ticker
	var timeoutTimer *time.Timer

	c := NewTestCapture()

	for {
		a(c)
		if c.Length() == 0 {
			return true
		}

		if intervalTicker == nil {
			intervalTicker = time.NewTicker(ts.interval)
			timeoutTimer = time.NewTimer(ts.timeout)
			defer intervalTicker.Stop()
			defer timeoutTimer.Stop()
		}

		select {
		case <-timeoutTimer.C:
			for _, cap := range c.Captures() {
				captured := cap.Captured()
				ts.t.Errorf("Assertion not satisfied within %v: %s%s", ts.timeout, *captured, PrintStack(3))
			}
			return false
		case <-intervalTicker.C:
		}
		c.Reset()
	}
}
