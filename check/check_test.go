package check

import (
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPrintStack(t *testing.T) {
	stack := PrintStack(0)
	lines := strings.FieldsFunc(stack, func(r rune) bool {
		return r == '\n'
	})

	assert.Equal(t, 2, len(lines))
	assert.Contains(t, lines[0], "> check.go")
	assert.Contains(t, lines[1], "  check_test.go")
}

func TestThatPanics_withPanic(t *testing.T) {
	c := NewTestCapture()

	ThatPanicsAsExpected(c, AnyCause(), func() {
		panic(fmt.Errorf("Boom"))
	})

	// Test should complete without a reported error
	c.First().AssertNil(t)
}

func TestThatPanics_withoutPanic(t *testing.T) {
	c := NewTestCapture()

	ThatPanicsAsExpected(c, AnyCause(), func() {})

	// Test should complete with a reported error
	c.First().AssertContains(t, "Did not panic as expected")
	t.Log(c.First().CapturedLines())
}

func TestThatPanicsAsExpected_withExpectedPanic_ErrorWithValue(t *testing.T) {
	c := NewTestCapture()

	ThatPanicsAsExpected(c, ErrorWithValue("Boom"), func() {
		panic(fmt.Errorf("Boom"))
	})

	// Test should complete without a reported error
	c.First().AssertNil(t)
}

func TestThatPanicsAsExpected_withExpectedPanic_CauseEqual(t *testing.T) {
	c := NewTestCapture()

	ThatPanicsAsExpected(c, CauseEqual("Boom"), func() {
		panic("Boom")
	})

	// Test should complete without a reported error
	c.First().AssertNil(t)
}

func TestThatPanicsAsExpected_withUnexpectedPanic_ErrorWithValue(t *testing.T) {
	c := NewTestCapture()

	ThatPanicsAsExpected(c, ErrorWithValue("Boom"), func() {
		panic(fmt.Errorf("Blast"))
	})

	// Test should complete with a reported error
	c.First().AssertContains(t, "Not equal")
	t.Log(c.First().CapturedLines())
}

func TestThatPanicsAsExpected_withUnexpectedPanic_ErrorWitHValue_typeMismatch(t *testing.T) {
	c := NewTestCapture()

	ThatPanicsAsExpected(c, ErrorWithValue("Boom"), func() {
		panic(42)
	})

	// Test should complete with a reported error
	c.First().AssertContains(t, "Expected error, got int")
	t.Log(c.First().CapturedLines())
}

func TestThatPanicsAsExpected_withUnexpectedPanic_ErrorContaining(t *testing.T) {
	c := NewTestCapture()

	ThatPanicsAsExpected(c, ErrorContaining("Boom"), func() {
		panic(fmt.Errorf("Blast"))
	})

	// Test should complete with a reported error
	c.First().AssertContains(t, "does not contain")
	t.Log(c.First().CapturedLines())
}

func TestThatPanicsAsExpected_withUnexpectedPanic_ErrorContaining_typeMismatch(t *testing.T) {
	c := NewTestCapture()

	ThatPanicsAsExpected(c, ErrorContaining("Boom"), func() {
		panic(42)
	})

	// Test should complete with a reported error
	c.First().AssertContains(t, "Expected error, got int")
	t.Log(c.First().CapturedLines())
}

func TestThatPanicsAsExpected_withUnexpectedPanic_CauseEqual(t *testing.T) {
	c := NewTestCapture()

	ThatPanicsAsExpected(c, CauseEqual("Boom"), func() {
		panic(42)
	})

	// Test should complete with a reported error
	c.First().AssertContains(t, "Not equal")
	t.Log(c.First().CapturedLines())
}

func TestThatDoesNotPanic_withoutPanic(t *testing.T) {
	c := NewTestCapture()

	ThatDoesNotPanic(c, func() {})

	// Test should complete without a reported error
	c.First().AssertNil(t)
}

func TestThatDoesNotPanic_withPanic(t *testing.T) {
	c := NewTestCapture()

	ThatDoesNotPanic(c, func() {
		panic(fmt.Errorf("Blast"))
	})

	// Test should complete with a reported error
	c.First().AssertContains(t, "Unexpected panic: Blast")
	t.Log(c.First().CapturedLines())
}

func TestWait_optionalArgsTooLong(t *testing.T) {
	ThatPanicsAsExpected(t, ErrorWithValue("argument list too long"), func() {
		Wait(t, time.Microsecond, time.Millisecond, time.Second)
	})
}

func TestWait_conditionWithinDeadline(t *testing.T) {
	c := NewTestCapture()

	counter := int32(3)

	passed := Wait(c, 10*time.Second).Until(func() bool {
		c := atomic.LoadInt32(&counter)
		if c > 0 {
			atomic.StoreInt32(&counter, c-1)
			return false
		}
		return true
	})
	assert.True(t, passed)

	c.First().AssertNil(t)
}

func TestWait_conditionNotWithinDeadline(t *testing.T) {
	c := NewTestCapture()

	passed := Wait(c, 1*time.Millisecond, 1*time.Microsecond).Until(func() bool {
		return false
	})
	assert.False(t, passed)

	c.First().AssertFirstLineContains(t, "Assertion not satisfied within 1ms")
	t.Log(c.First().CapturedLines())
}

func TestWait_equalsCondition(t *testing.T) {
	c := NewTestCapture()

	// Simple incrementor
	v := 0
	f := func() interface{} {
		v++
		return v
	}

	passed := Wait(c, 10*time.Second).Until(Equal(f, 2))
	assert.True(t, passed)

	c.First().AssertNil(t)
}

func TestWait_notEqualsCondition(t *testing.T) {
	c := NewTestCapture()

	// Simple incrementor
	v := 0
	f := func() interface{} {
		v++
		return v
	}

	passed := Wait(c, 10*time.Second).Until(Not(Equal(f, 0)))
	assert.True(t, passed)

	c.First().AssertNil(t)
}

func TestWait_assertionWithinDeadline(t *testing.T) {
	c := NewTestCapture()

	counter := int32(3)

	passed := Wait(c, 10*time.Second).UntilAsserted(func(t Tester) {
		c := atomic.LoadInt32(&counter)
		if c > 0 {
			t.Errorf("c is %d", c)
			atomic.StoreInt32(&counter, c-1)
			return
		}
	})
	assert.True(t, passed)

	c.First().AssertNil(t)
}

func TestWait_assertionNotWithinDeadline(t *testing.T) {
	c := NewTestCapture()

	passed := Wait(c, 1*time.Millisecond, 1*time.Microsecond).UntilAsserted(func(t Tester) {
		t.Errorf("Not happening")
	})
	assert.False(t, passed)

	c.First().AssertFirstLineEqual(t, "Assertion not satisfied within 1ms: Not happening")
	t.Log(c.First().CapturedLines())
	assert.Equal(t, 2, c.First().NumCapturedLines()) // check stack trace elements
}

func TestWait_multipleAssertionsNotWithinDeadline(t *testing.T) {
	c := NewTestCapture()

	passed := Wait(c, 1*time.Millisecond, 1*time.Microsecond).UntilAsserted(func(t Tester) {
		t.Errorf("Not happening")
		t.Errorf("Still not happening")
	})
	assert.False(t, passed)

	first := c.Capture(0)
	first.AssertFirstLineEqual(t, "Assertion not satisfied within 1ms: Not happening")
	t.Log(first.CapturedLines())
	assert.Equal(t, 2, first.NumCapturedLines()) // check stack trace elements

	second := c.Capture(1)
	second.AssertFirstLineEqual(t, "Assertion not satisfied within 1ms: Still not happening")
	t.Log(second.CapturedLines())
	assert.Equal(t, 2, second.NumCapturedLines()) // check stack trace elements
}
