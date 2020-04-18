package concurrent

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/obsidiandynamics/stdlibgo/arity"
)

// AtomicCounter encapsulates an int64 value that may updated atomically.
type AtomicCounter interface {
	fmt.Stringer
	Get() int64
	GetInt() int
	Add(amount int64) int64
	Inc() int64
	Dec() int64
	Set(amount int64)
	CompareAndSwap(expected int64, replacement int64) bool
	Fill(atLeast int64, timeout time.Duration, interval ...time.Duration) int64
	Drain(atMost int64, timeout time.Duration, interval ...time.Duration) int64
	Await(cond I64Condition, timeout time.Duration, interval ...time.Duration) int64
	AwaitCtx(ctx context.Context, cond I64Condition, interval ...time.Duration) int64
}

type atomicCounter struct {
	notify chan int
	value  int64
}

// NewAtomicCounter creates a new counter, optionally assigning its value to the given
// initial value (0 by default)
func NewAtomicCounter(initial ...int64) AtomicCounter {
	c := &atomicCounter{}
	c.value = arity.SoleUntyped(int64(0), initial).(int64)
	c.notify = make(chan int, 1)
	return c
}

// String obtains a string representation of the atomic counter.
func (c atomicCounter) String() string {
	return fmt.Sprint("AtomicCounter[", c.Get(), "]")
}

// Gets the current value of the counter.
func (c *atomicCounter) Get() int64 {
	return atomic.LoadInt64(&c.value)
}

// GetInt obtains the current value as a signed int.
func (c *atomicCounter) GetInt() int {
	return int(c.Get())
}

// Adds a specified amount to the counter, returning the updated value.
func (c *atomicCounter) Add(amount int64) int64 {
	defer c.notifyUpdate()
	return atomic.AddInt64(&c.value, amount)
}

// Increments the counter, returning the updated value.
func (c *atomicCounter) Inc() int64 {
	return c.Add(1)
}

// Decrements the counter, returning the updated value.
func (c *atomicCounter) Dec() int64 {
	return c.Add(-1)
}

// Sets a new value to the counter.
func (c *atomicCounter) Set(amount int64) {
	defer c.notifyUpdate()
	atomic.StoreInt64(&c.value, amount)
}

func (c *atomicCounter) notifyUpdate() {
	select {
	case c.notify <- 0:
	default:
	}
}

// CompareAndSwap conditionally assigns a replacement value if the existing value matched the given
// expected value.
func (c *atomicCounter) CompareAndSwap(expected int64, replacement int64) bool {
	if atomic.CompareAndSwapInt64(&c.value, expected, replacement) {
		c.notifyUpdate()
		return true
	}
	return false
}

// Fill blocks until the counter reaches a value that is at least a given minimum.
func (c *atomicCounter) Fill(atLeast int64, timeout time.Duration, interval ...time.Duration) int64 {
	return c.Await(I64GreaterThanOrEqual(atLeast), timeout, interval...)
}

// Drain blocks until the counter drops to a value that does not exceed a given maximum.
func (c *atomicCounter) Drain(atMost int64, timeout time.Duration, interval ...time.Duration) int64 {
	return c.Await(I64LessThanOrEqual(atMost), timeout, interval...)
}

// DefaultCounterCheckInterval is the default check interval used by Await/AwaitCtx/Drain/Fill.
const DefaultCounterCheckInterval = 10 * time.Millisecond

// Await blocks until a condition is met or expires, returning the last observed counter value. The optional
// interval argument places an upper bound on the check interval (defaults to DefaultCounterCheckInterval).
func (c *atomicCounter) Await(cond I64Condition, timeout time.Duration, interval ...time.Duration) int64 {
	ctx, cancel := Timeout(context.Background(), timeout)
	defer cancel()
	return c.AwaitCtx(ctx, cond, interval...)
}

// Await blocks until a condition is met or the context is cancelled, returning the last observed counter value.
// The optional interval argument places an upper bound on the check interval (defaults to DefaultCounterCheckInterval).
func (c *atomicCounter) AwaitCtx(ctx context.Context, cond I64Condition, interval ...time.Duration) int64 {
	checkInterval := optional(DefaultCounterCheckInterval, interval...)
	var sleepTicker *time.Ticker
	for {
		value := c.Get()
		if cond(value) {
			return value
		}

		if sleepTicker == nil {
			sleepTicker = time.NewTicker(checkInterval)
			defer sleepTicker.Stop()
		}

		select {
		case <-ctx.Done():
			return value
		case <-c.notify:
		case <-sleepTicker.C:
		}
	}
}
