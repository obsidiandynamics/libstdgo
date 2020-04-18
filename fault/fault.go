// Package fault provides facilities for simulated fault injection.
package fault

import (
	"math/rand"

	"github.com/obsidiandynamics/libstdgo/concurrent"
)

// Spec outlines the conditions for a fault, comprising a contingency, as well as an error that is reported when said contingency arises.
//
// Specifications are completely reusable; one can create multiple Fault objects from a single Spec. Fault objects,
// on the other hand, should not be reused as they encompass invocation counters.
type Spec struct {
	Cnt Contingency
	Err error
}

// Fault is an injector of simulated errors. A single fault instance should be spawned for one test.
//
// A fault is thread-safe; it can be invoked from multiple goroutines.
type Fault interface {
	Try() error
	Calls() int
	Faults() int
}

// None is a convenience function for specifying a no-fault.
func None() Spec {
	return Spec{Never(), nil}
}

// Build creates a Fault instance from its Spec.
func (s Spec) Build() Fault {
	if s.Cnt != nil {
		return &fault{
			spec:   s,
			calls:  concurrent.NewAtomicCounter(),
			faults: concurrent.NewAtomicCounter(),
		}
	}

	// If zero value of Spec was provided (where the contingency is nil).
	return None().Build()
}

type fault struct {
	spec   Spec
	calls  concurrent.AtomicCounter
	faults concurrent.AtomicCounter
}

// Try simulates an invocation, returning an error if a contingency occurs. The total number of invocations and
// the number of injected faults are retained within the Fault struct.
func (f *fault) Try() error {
	f.calls.Inc()
	if f.spec.Cnt(f) {
		f.faults.Inc()
		return f.spec.Err
	}
	return nil
}

// Calls returns the total number of invocations, including those that have started but not yet returned.
func (f *fault) Calls() int {
	return f.calls.GetInt()
}

// Faults returns the number of injected faults.
func (f *fault) Faults() int {
	return f.faults.GetInt()
}

// Contingency is a condition under which a fault should be injected. It is effectively a predicate; if it
// evaluates to true, a fault will be injected. Otherwise, if false, no fault will be returned to the application.
type Contingency func(f Fault) bool

// Never is a contingency that never occurs.
func Never() Contingency {
	return func(f Fault) bool {
		return false
	}
}

// Always is a contingency that always occurs.
func Always() Contingency {
	return func(f Fault) bool {
		return true
	}
}

// Random is a contingency that occurs with a probability equal to the given p-value.
func Random(p float32) Contingency {
	return func(f Fault) bool {
		return rand.Float32() < p
	}
}

// First is a contingency that occurs during the first n attempts.
func First(n int) Contingency {
	return func(f Fault) bool {
		return f.Calls() <= n
	}
}

// After is a contingency that occurs after the first n attempts.
func After(n int) Contingency {
	return func(f Fault) bool {
		return f.Calls() > n
	}
}
