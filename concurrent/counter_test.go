package concurrent

import (
	"context"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewAtomicCounterWithInitialValue(t *testing.T) {
	c := NewAtomicCounter(42)
	assert.Equal(t, 42, c.GetInt())
}

func TestAtomicCounterDrainInDeepSleep(t *testing.T) {
	c := NewAtomicCounter(1)
	go func() {
		time.Sleep(1 * time.Millisecond)
		res := c.Add(-1)
		assert.Equal(t, int64(0), res)
	}()

	res := c.Drain(0, Indefinitely, 1*time.Hour)
	assert.Equal(t, int64(0), res)
}

func TestAtomicCounterAwaitWithTwoWaiters(t *testing.T) {
	c := NewAtomicCounter(1)
	wg := sync.WaitGroup{}
	const waiters = 2
	wg.Add(waiters)

	waitedAtLeastOnce := sync.WaitGroup{}
	waitedAtLeastOnce.Add(waiters)
	for i := 0; i < waiters; i++ {
		go func() {
			waited := false
			res := c.Await(func(value int64) bool {
				if !waited {
					waited = true
					waitedAtLeastOnce.Done()
				}
				return value == 0
			}, Indefinitely, 1*time.Nanosecond)
			assert.Equal(t, int64(0), res)
			wg.Done()
		}()
	}

	waitedAtLeastOnce.Wait()
	res := c.Dec()
	assert.Equal(t, int64(0), res)
	wg.Wait()

	// Tests that we can repeatedly set the value without anyone awaiting it.
	c.Set(0)
	c.Set(0)
}

func TestAtomicCounterAwaitCtxInDeepSleep(t *testing.T) {
	c := NewAtomicCounter(1)
	go func() {
		time.Sleep(1 * time.Millisecond)
		res := c.Dec()
		assert.Equal(t, int64(0), res)
	}()

	ctx, cancel := Forever(context.Background())
	defer cancel()
	res := c.AwaitCtx(ctx, I64Equal(0), 1*time.Hour)
	assert.Equal(t, int64(0), res)
}

func TestAtomicCounterAwaitCtxCancel(t *testing.T) {
	c := NewAtomicCounter(1)
	ctx, cancel := Forever(context.Background())
	go func() {
		time.Sleep(1 * time.Millisecond)
		cancel()
	}()

	defer cancel()
	res := c.AwaitCtx(ctx, I64Equal(0), 1*time.Hour)
	assert.Equal(t, int64(1), res)
}

func TestAtomicCounterDrainWithTimeout(t *testing.T) {
	c := NewAtomicCounter(1)
	res := c.Drain(0, 1*time.Microsecond)
	assert.Equal(t, int64(1), res)
}

func TestAtomicCounterFillWithTimeout(t *testing.T) {
	c := NewAtomicCounter()
	res := c.Fill(1, 1*time.Microsecond)
	assert.Equal(t, int64(0), res)
}

func TestAtomicCounterIncrement(t *testing.T) {
	c := NewAtomicCounter()
	res := c.Inc()
	assert.Equal(t, int64(1), res)
	assert.Equal(t, 1, c.GetInt())
}

func TestAtomicCounterThreadedIncrement(t *testing.T) {
	c := NewAtomicCounter()

	const routines = 10
	const perRoutine = 100

	wg := sync.WaitGroup{}
	wg.Add(routines)
	for r := 0; r < routines; r++ {
		go func() {
			defer wg.Done()
			for j := 0; j < perRoutine; j++ {
				c.Inc()
				runtime.Gosched()
			}
		}()
	}
	wg.Wait()

	assert.Equal(t, routines*perRoutine, c.GetInt())
}

func TestAtomicCounterSet(t *testing.T) {
	c := NewAtomicCounter(3)
	c.Set(7)
	assert.Equal(t, 7, c.GetInt())
}

func TestAtomicCounterCompareAndSwap(t *testing.T) {
	c := NewAtomicCounter(3)
	assert.False(t, c.CompareAndSwap(2, 3))
	assert.Equal(t, 3, c.GetInt())
	assert.True(t, c.CompareAndSwap(3, 2))
	assert.Equal(t, 2, c.GetInt())
}

func TestAtomicCounterStringer(t *testing.T) {
	c := NewAtomicCounter(1)
	assert.Equal(t, "AtomicCounter[1]", c.String())
}
