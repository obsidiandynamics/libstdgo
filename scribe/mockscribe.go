package scribe

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/obsidiandynamics/libstdgo/check"
)

// MockScribe provides a facility for mocking a Scribe, capturing log calls for subsequent inspection, filtering and assertions.
// This implementation is thread-safe.
type MockScribe interface {
	Loggers() LoggerFactories
	Reset()
	Entries() Entries
	ContainsEntries() DynamicAssertion
}

// Entry is a single, captured log entry.
type Entry struct {
	Timestamp time.Time
	Level     Level
	Format    string
	Args      []interface{}
	Scene     Scene
}

// FormattedMessage returns the application of fmt.Sprintf to Entry.Format and Entry.Args.
func (e Entry) FormattedMessage() string {
	return fmt.Sprintf(e.Format, e.Args...)
}

// String obtains a textual representation of an entry.
func (e Entry) String() string {
	return fmt.Sprint("Entry[Timestamp=", e.Timestamp,
		", Level=", e.Level,
		", Format=", e.Format,
		", Args=", e.Args,
		", Scene=", e.Scene, "]")
}

// Predicate is a condition evaluated against a given entry, returning true if the underlying condition has been satisfied.
type Predicate func(e Entry) bool

// Assertion is a verification of Entries that returns a nil string if the assertion passes, or a string describing the nature
// of the failure otherwise. MockScribe will append the stack trace behind the scenes.
type Assertion func(e Entries) *string

// Entries is an immutable snapshot of captured log calls.
type Entries interface {
	Having(p Predicate) Entries
	List() []Entry
	Length() int
	Assert(t check.Tester, a Assertion) Entries
}

type entries []Entry

type mockScribe struct {
	lock    sync.Mutex
	entries entries
}

// NewMock creates a new MockScribe. The returning instance cannot be used to log directly — only to inspect and assert captures.
// To configure a Scribe to use the mocks for subsequent logging:
//  mock := scribe.NewMock()
//	scribe := scribe.New(mock.Loggers())
func NewMock() MockScribe {
	return &mockScribe{}
}

/*
Implemented methods.
*/

// Loggers obtains the necessary LoggerFactories to configure Scribe.
func (s *mockScribe) Loggers() LoggerFactories {
	facs := LoggerFactories{}

	for _, l := range Levels {
		if l.Level == Off {
			continue
		}

		level := l.Level
		facs[level] = func(level Level, scene Scene) Logger {
			return func(format string, args ...interface{}) {
				s.append(Entry{
					Timestamp: time.Now(),
					Level:     level,
					Format:    format,
					Args:      args,
					Scene:     scene,
				})
			}
		}
	}

	return facs
}

// Resets the mock, clearing any calls that may have been previously captured.
func (s *mockScribe) Reset() {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.entries = []Entry{}
}

// Obtains a snapshot of captured entries. Any subsequent captures will not effect the contents of the
// returned snapshot.
func (s *mockScribe) Entries() Entries {
	s.lock.Lock()
	defer s.lock.Unlock()
	// Use the Anything predicate to take a copy of the accumulated entries
	return s.entries.Having(Anything())
}

// Having is a filtering operation that takes a copy of the snapshot, eliminating entries that do not
// satisfy the given predicate. The original Entries structure remains unchanged.
func (e entries) Having(p Predicate) Entries {
	filtered := make(entries, 0, len(e))
	for _, entry := range e {
		if p(entry) {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}

// Asserts that the contents of the Entries snapshot satisfy the given assertion. Because the assertion
// operates on the snapshot, any additional log calls that are captured by the mock are not reflected
// in the snapshot and, therefore, do not impact the assertion. Use a combination of ContainsEntries()
// with DynamicAssertion to test the current state of the mock.
func (e entries) Assert(t check.Tester, a Assertion) Entries {
	msg := a(e)
	if msg != nil {
		t.Errorf("%s%s", *msg, check.PrintStack(2))
	}
	return e
}

// Length returns the number of captured log calls.
func (e entries) Length() int {
	return len(e)
}

// List returns a slice of the underlying entries.
func (e entries) List() []Entry {
	return e
}

/*
 * Private methods.
 */

func (s *mockScribe) append(e Entry) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.entries = append(s.entries, e)
}

/*
Predicates.
*/

// Anything is a predicate that matches any entry. It is useful for taking a copy of Entries:
//  exactCopy := original.Having(scribe.Anything())
func Anything() Predicate {
	return func(_ Entry) bool {
		return true
	}
}

// LogLevel matches all entries that are logged at the given level.
func LogLevel(level Level) Predicate {
	return func(e Entry) bool {
		return e.Level == level
	}
}

// MessageContaining matches entries where the formatted message contains the given substr.
func MessageContaining(substr string) Predicate {
	return func(e Entry) bool {
		return strings.Contains(e.FormattedMessage(), substr)
	}
}

// MessageEqual matches entries where the formatted message exactly matches the expected string.
func MessageEqual(expected string) Predicate {
	return func(e Entry) bool {
		return e.FormattedMessage() == expected
	}
}

// Not produces a logical inverse of a predicate.
func Not(p Predicate) Predicate {
	return func(e Entry) bool {
		return !p(e)
	}
}

// ScenePredicate is a refinement of the predicate concept, applying to the Scene field of an Entry
// (as opposed to the entire Entry struct).
type ScenePredicate func(scene Scene) bool

// ASceneWith returns a conventional (Entry) predicate that satisfies the given ScenePredicate.
func ASceneWith(p ScenePredicate) Predicate {
	return func(e Entry) bool {
		return p(e.Scene)
	}
}

// AField is satisfied if the scene contains a field with the given name-value pair.
func AField(name string, value interface{}) ScenePredicate {
	return func(scene Scene) bool {
		existing, ok := scene.Fields[name]
		return ok && existing == value
	}
}

// AFieldNamed is satisfied if the scene contains a field with the given name.
func AFieldNamed(name string) ScenePredicate {
	return func(scene Scene) bool {
		_, ok := scene.Fields[name]
		return ok
	}
}

// AnError is satisfied if the scene holds an error.
func AnError() ScenePredicate {
	return func(scene Scene) bool {
		return scene.Err != nil
	}
}

// Invert a scene predicate.
func (p ScenePredicate) Invert() ScenePredicate {
	return func(scene Scene) bool { return !p(scene) }
}

// Content is satisfied if the scene has any of its fields set.
func Content() ScenePredicate {
	return func(scene Scene) bool { return scene.IsSet() }
}

/*
Assertions.
*/

// Count ensures that the number of entries equals the given expected number.
func Count(expected int) Assertion {
	return func(e Entries) *string {
		actual := len(e.List())
		if actual == expected {
			return nil
		}
		msg := fmt.Sprintf("Expected %d entries; got %d", expected, actual)
		return &msg
	}
}

// CountAtLeast ensures that there is a minimum number of entries.
func CountAtLeast(minimum int) Assertion {
	return func(e Entries) *string {
		actual := len(e.List())
		if actual >= minimum {
			return nil
		}
		msg := fmt.Sprintf("Expected at least %d entries; got %d", minimum, actual)
		return &msg
	}
}

// CountAtMost ensures that there is a maximum number of entries.
func CountAtMost(maximum int) Assertion {
	return func(e Entries) *string {
		actual := len(e.List())
		if actual <= maximum {
			return nil
		}
		msg := fmt.Sprintf("Expected at most %d entries; got %d", maximum, actual)
		return &msg
	}
}

/*
Dynamic assertions.
*/

// DynamicAssertion permits the testing of captures housed by the ScribeMock, rather than the Entries snapshot.
// Owing to this, assertions created from a DynamicAssertion are evaluated against a refreshed Entries snapshot. This
// is ideal for performing time-based assertions, where a condition might be satisfied after some time (when the
// application eventually logs the missing entry).
type DynamicAssertion struct {
	m          MockScribe
	predicates []Predicate
}

// ContainsEntries creates a new DynamicAssertion stub.
func (s *mockScribe) ContainsEntries() DynamicAssertion {
	return DynamicAssertion{m: s}
}

// Having applies a predicate to the DynamicAssertion.
func (s DynamicAssertion) Having(p Predicate) DynamicAssertion {
	s.predicates = append(s.predicates, p)
	return s
}

// Passes returns an Assertion that is evaluated against a refreshed Entries snapshot.
func (s DynamicAssertion) Passes(a Assertion) check.Assertion {
	return func(t check.Tester) {
		entries := s.m.Entries()
		for _, p := range s.predicates {
			entries = entries.Having(p)
		}
		entries.Assert(t, a)
	}
}
