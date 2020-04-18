package concurrent

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/obsidiandynamics/libstdgo/check"
	"github.com/stretchr/testify/assert"
)

func TestNewAtomicReference_withInitial(t *testing.T) {
	r := NewAtomicReference(42)
	assert.Equal(t, 42, r.Get())
}

func TestAtomicReferenceAwait_inDeepSleep(t *testing.T) {
	r := NewAtomicReference(1)
	ctx, cancel := Forever(context.Background())
	go func() {
		time.Sleep(1 * time.Millisecond)
		cancel()
	}()

	res := r.AwaitCtx(ctx, RefEqual(0), 1*time.Hour)
	assert.Equal(t, 1, res)
}

func TestAtomicReferenceAwaitWithTwoWaiters(t *testing.T) {
	r := NewAtomicReference(1)
	wg := sync.WaitGroup{}
	const waiters = 2
	wg.Add(waiters)

	waitedAtLeastOnce := sync.WaitGroup{}
	waitedAtLeastOnce.Add(waiters)
	for i := 0; i < waiters; i++ {
		go func() {
			waited := false
			res := r.Await(func(value interface{}) bool {
				if !waited {
					waited = true
					waitedAtLeastOnce.Done()
				}
				return value == 0
			}, Indefinitely, 1*time.Nanosecond)
			assert.Equal(t, 0, res)
			wg.Done()
		}()
	}

	waitedAtLeastOnce.Wait()
	r.Set(0)
	wg.Wait()

	// Tests that we can repeatedly set the value without anyone awaiting it.
	r.Set(0)
	r.Set(0)
}

func TestAtomicReferenceAwait_ctxCancel(t *testing.T) {
	r := NewAtomicReference(1)
	go func() {
		time.Sleep(1 * time.Millisecond)
		r.Set(0)
	}()

	res := r.Await(RefEqual(0), Indefinitely, 1*time.Hour)
	assert.Equal(t, 0, res)
}

func TestAtomicReferenceAwait_RefNotEqual_withTimeout(t *testing.T) {
	r := NewAtomicReference(1)
	res := r.Await(RefEqual(0), 1*time.Microsecond)
	assert.Equal(t, 1, res)
}

func TestAtomicReferenceAwait_RefEqual_withTimeout(t *testing.T) {
	r := NewAtomicReference(1)
	res := r.Await(RefNot(RefEqual(1)), 1*time.Microsecond)
	assert.Equal(t, 1, res)
}

func TestAtomicReference_RefNil(t *testing.T) {
	assert.True(t, RefNil()(nil))
	assert.False(t, RefNil()("test"))
}

func TestAtomicReference_Stringer(t *testing.T) {
	cases := []struct {
		initial      []interface{}
		expectString string
	}{
		{[]interface{}{1}, "1"},
		{[]interface{}{check.ErrSimulated}, "simulated"},
		{[]interface{}{}, "<nil>"},
		{[]interface{}{nil}, "<nil>"},
	}

	for _, c := range cases {
		t := check.Intercept(t).Mutate(check.Appendf("\n%v", c))
		v := NewAtomicReference(c.initial...)
		assert.Equal(t, c.expectString, v.String())
		if len(c.initial) == 1 {
			assert.Equal(t, c.initial[0], v.Get())
		} else {
			assert.Nil(t, v.Get())
		}
	}
}
