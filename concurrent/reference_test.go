package concurrent

import (
	"context"
	"testing"
	"time"

	"github.com/obsidiandynamics/libstdgo/check"
	"github.com/stretchr/testify/assert"
)

func TestNewAtomicReference_withInitial(t *testing.T) {
	v := NewAtomicReference(42)
	assert.Equal(t, 42, v.Get())
}

func TestAtomicReferenceAwait_inDeepSleep(t *testing.T) {
	v := NewAtomicReference(1)
	ctx, cancel := Forever(context.Background())
	go func() {
		time.Sleep(1 * time.Millisecond)
		cancel()
	}()

	res := v.AwaitCtx(ctx, RefEqual(0), 1*time.Hour)
	assert.Equal(t, 1, res)
}

func TestAtomicReferenceAwait_ctxCancel(t *testing.T) {
	v := NewAtomicReference(1)
	go func() {
		time.Sleep(1 * time.Millisecond)
		v.Set(0)
	}()

	res := v.Await(RefEqual(0), Indefinitely, 1*time.Hour)
	assert.Equal(t, 0, res)
}

func TestAtomicReferenceAwait_RefNotEqual_withTimeout(t *testing.T) {
	v := NewAtomicReference(1)
	res := v.Await(RefEqual(0), 1*time.Microsecond)
	assert.Equal(t, 1, res)
}

func TestAtomicReferenceAwait_RefEqual_withTimeout(t *testing.T) {
	v := NewAtomicReference(1)
	res := v.Await(RefNot(RefEqual(1)), 1*time.Microsecond)
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
