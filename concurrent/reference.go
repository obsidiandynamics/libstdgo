package concurrent

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/obsidiandynamics/libstdgo/arity"
)

// AtomicReference encapsulates a pointer that may updated atomically. Unlike its sync/atomic.Value counterpart,
// this implementation permits nil pointers.
type AtomicReference interface {
	fmt.Stringer
	Set(value interface{})
	Get() interface{}
	Await(cond RefCondition, timeout time.Duration, interval ...time.Duration) interface{}
	AwaitCtx(ctx context.Context, cond RefCondition, interval ...time.Duration) interface{}
}

type pointer struct {
	referent interface{}
}

type atomicReference struct {
	notify chan int
	value  atomic.Value
}

// NewAtomicReference creates a new reference, optionally assigning its contents to the given
// initial referent (nil by default)
func NewAtomicReference(initial ...interface{}) AtomicReference {
	v := atomicReference{
		notify: make(chan int, 1),
		value:  atomic.Value{},
	}
	initVal := arity.SoleUntyped(nil, initial)
	v.value.Store(pointer{initVal})
	return &v
}

// String obtains a string representation of the atomic reference, printing the underlying referent.
func (v atomicReference) String() string {
	return fmt.Sprint(v.Get())
}

// Sets a new referent.
func (v *atomicReference) Set(referent interface{}) {
	v.value.Store(pointer{referent})
	select {
	case v.notify <- 0:
	default:
	}
}

// Gets the current referent of the reference.
func (v *atomicReference) Get() interface{} {
	return v.value.Load().(pointer).referent
}

// DefaultReferenceCheckInterval is the default check interval used by Await/AwaitCtx.
const DefaultReferenceCheckInterval = 10 * time.Millisecond

// RefCondition is a predicate that checks whether the current (supplied) referent meets some condition, returning
// true if the condition is met.
type RefCondition func(referent interface{}) bool

// RefNot produces a logical inverse of the given condition.
func RefNot(cond RefCondition) RefCondition {
	return func(referent interface{}) bool { return !cond(referent) }
}

// RefNil checks that the current referent is nil.
func RefNil() RefCondition {
	return func(referent interface{}) bool { return referent == nil }
}

// RefEqual tests that the encapsulated referent equals a target referent.
func RefEqual(target interface{}) RefCondition {
	return func(referent interface{}) bool { return referent == target }
}

// Await blocks until a condition is met or expires, returning the last observed referent. The optional
// interval argument places an upper bound on the check interval (defaults to DefaultReferenceCheckInterval).
func (v *atomicReference) Await(cond RefCondition, timeout time.Duration, interval ...time.Duration) interface{} {
	ctx, cancel := Timeout(context.Background(), timeout)
	defer cancel()
	return v.AwaitCtx(ctx, cond, interval...)
}

// Await blocks until a condition is met or the context is cancelled, returning the last observed referent.
// The optional interval argument places an upper bound on the check interval (defaults to DefaultReferenceCheckInterval).
func (v *atomicReference) AwaitCtx(ctx context.Context, cond RefCondition, interval ...time.Duration) interface{} {
	checkInterval := optional(DefaultReferenceCheckInterval, interval...)
	var sleepTicker *time.Ticker
	for {
		referent := v.Get()
		if cond(referent) {
			return referent
		}

		if sleepTicker == nil {
			sleepTicker = time.NewTicker(checkInterval)
			defer sleepTicker.Stop()
		}

		select {
		case <-ctx.Done():
			return referent
		case <-v.notify:
		case <-sleepTicker.C:
		}
	}
}
