package check

import (
	"fmt"
	"strings"
	"sync"
)

const mockTesterStackDepth = 2

// MockTester aids in the mocking of the Tester interface.
type MockTester struct {
	ErrorFunc func(format string, args ...interface{})
}

// Errorf feeds a formatted error message to the mocked tester.
func (m *MockTester) Errorf(format string, args ...interface{}) {
	m.ErrorFunc(format, args...)
}

// TestCapture provides a mechanism for capturing the results of tests. This is useful when testing assertion
// libraries.
// TestCapture is thread-safe; it may be invoked concurrently from multiple goroutines to capture test results.
type TestCapture interface {
	Tester
	First() SingleCapture
	Capture(index int) SingleCapture
	Captures() []SingleCapture
	Length() int
	Reset()
}

type testCapture struct {
	MockTester
	lock     sync.Mutex
	captured []string
}

// SingleCapture represents one instance of the invocation of TestCapture.Errorf.
type SingleCapture interface {
	Captured() *string
	CapturedLines() []string
	NumCapturedLines() int
	AssertNil(t Tester)
	AssertNotNil(t Tester)
	AssertFirstLineEqual(t Tester, expected string)
	AssertFirstLineContains(t Tester, substr string)
	AssertContains(t Tester, substr string)
}

type singleCapture struct {
	captured *string
}

// NewTestCapture creates a new TestCapture object.
func NewTestCapture() TestCapture {
	c := &testCapture{}
	c.ErrorFunc = func(format string, args ...interface{}) {
		msg := fmt.Sprintf(format, args...)
		c.lock.Lock()
		defer c.lock.Unlock()
		c.captured = append(c.captured, msg)
	}
	return c
}

// First is a convenience for Capture(0). It's used often in testing. If no invocations occurred,
// this method still returns a SingleCapture object, albeit an empty one (containing a nil string).
func (c *testCapture) First() SingleCapture {
	return c.Capture(0)
}

// Capture returns a SingleCapture instance in the invocation order. E.g. Capture(2) returns the third
// call to TestCapture.Errorf. If no invocation took place for the given index, an empty SingleCapture
// is returned (containing a nil capture string).
func (c *testCapture) Capture(index int) SingleCapture {
	captures := c.Captures()
	if length := len(captures); index < length {
		return captures[index]
	}
	return &singleCapture{nil}
}

// Captures returns a copy of all captured invocations.
func (c *testCapture) Captures() []SingleCapture {
	c.lock.Lock()
	defer c.lock.Unlock()
	copy := make([]SingleCapture, len(c.captured))
	for i, cap := range c.captured {
		cap := cap
		copy[i] = &singleCapture{&cap}
	}
	return copy
}

// Length obtains the number of captured invocations.
func (c *testCapture) Length() int {
	c.lock.Lock()
	defer c.lock.Unlock()
	return len(c.captured)
}

// Resets TestCapture to its initial (blank) state.
func (c *testCapture) Reset() {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.captured = []string{}
}

// Captured return the contents of the captured invocation as a pre-formatted string. If the
// invocation did not actually take place, nil is returned.
func (s *singleCapture) Captured() *string {
	return s.captured
}

// CapturedLines splits the captured content into lines, returning the resulting slice. If the
// invocation did not actually take place, nil is returned.
func (s *singleCapture) CapturedLines() []string {
	if s.captured == nil {
		return nil
	}
	return strings.FieldsFunc(*s.captured, func(r rune) bool {
		return r == '\n'
	})
}

// NumCapturedLines returns the number of captured lines, which is always one or greater if an invocation
// took place. Otherwise, if no invocation occurred, 0 is returned.
func (s *singleCapture) NumCapturedLines() int {
	if lines := s.CapturedLines(); lines != nil {
		return len(lines)
	}
	return 0
}

// AssertNil checks that nothing was captured. In other words, it verifies that the invocation did not
// take place.
func (s *singleCapture) AssertNil(t Tester) {
	if s.captured != nil {
		t.Errorf("Expected nil; got '%v'%s", *s.captured, PrintStack(mockTesterStackDepth))
	}
}

// AssetNotNil checks that something was captured, without being concerned with the contents of the capture.
func (s *singleCapture) AssertNotNil(t Tester) {
	if s.captured == nil {
		t.Errorf("Expected not nil%s", PrintStack(mockTesterStackDepth))
	}
}

// AssertFirstLineEqual checks that the first line of the capture matches the expected string. This method is
// useful because assertion frameworks often dump additional information (e.g. stack traces) after the first
// line, and it's often not practical to test these lines.
func (s *singleCapture) AssertFirstLineEqual(t Tester, expected string) {
	if s.captured == nil {
		t.Errorf("Expected '%s'; got nil%s", expected, PrintStack(mockTesterStackDepth))
		return
	}

	if lines := s.CapturedLines(); lines[0] != expected {
		t.Errorf("Expected '%s'; got '%s'%s", expected, lines[0], PrintStack(mockTesterStackDepth))
	}
}

// AssertFirstLineContains checks that the first line of the capture contains the given substring. Like
// AssertFirstLineEqual, this is useful for verifying multi-line assertion messages.
func (s *singleCapture) AssertFirstLineContains(t Tester, substr string) {
	if s.captured == nil {
		t.Errorf("Expected string containing '%s'; got nil%s", substr, PrintStack(mockTesterStackDepth))
		return
	}

	if lines := s.CapturedLines(); !strings.Contains(lines[0], substr) {
		t.Errorf("Expected string containing '%s'; got '%s'%s", substr, lines[0], PrintStack(mockTesterStackDepth))
	}
}

// AssertContains checks that the captured message contains the given substring.
func (s *singleCapture) AssertContains(t Tester, substr string) {
	if s.captured == nil {
		t.Errorf("Expected string containing '%s'; got nil%s", substr, PrintStack(mockTesterStackDepth))
		return
	}

	if !strings.Contains(*s.captured, substr) {
		t.Errorf("Expected string containing '%s'; got '%s'%s", substr, *s.captured, PrintStack(mockTesterStackDepth))
	}
}
