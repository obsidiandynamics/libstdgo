package fault

import (
	"testing"
	"time"

	"github.com/obsidiandynamics/libstdgo/check"
	"github.com/stretchr/testify/assert"
)

func TestNone(t *testing.T) {
	f := None().Build()
	assert.Equal(t, 0, f.Calls())
	assert.Equal(t, 0, f.Faults())

	assert.Nil(t, f.Try())
	assert.Equal(t, 1, f.Calls())
	assert.Equal(t, 0, f.Faults())
}

func TestZeroValue(t *testing.T) {
	var s Spec
	f := s.Build()
	assert.Equal(t, 0, f.Calls())
	assert.Equal(t, 0, f.Faults())

	assert.Nil(t, f.Try())
	assert.Equal(t, 1, f.Calls())
	assert.Equal(t, 0, f.Faults())
}

func TestAlways(t *testing.T) {
	f := Spec{Always(), check.ErrSimulated}.Build()
	assert.Equal(t, f.Try(), check.ErrSimulated)
	assert.Equal(t, 1, f.Calls())
	assert.Equal(t, 1, f.Faults())
}

func TestRandom_always(t *testing.T) {
	f := Spec{Random(1), check.ErrSimulated}.Build()
	assert.Equal(t, f.Try(), check.ErrSimulated)
	assert.Equal(t, 1, f.Calls())
	assert.Equal(t, 1, f.Faults())
}

func TestRandom_sometimes(t *testing.T) {
	f := Spec{Random(.1), check.ErrSimulated}.Build()
	check.Wait(t, time.Second, time.Nanosecond).UntilAsserted(func(t check.Tester) {
		assert.Equal(t, f.Try(), check.ErrSimulated)
	})
	calls := f.Calls()
	assert.GreaterOrEqual(t, calls, 1)
	assert.Equal(t, 1, f.Faults())

	check.Wait(t, time.Second, time.Nanosecond).UntilAsserted(func(t check.Tester) {
		assert.Nil(t, f.Try())
	})
	assert.Greater(t, f.Calls(), calls)
	assert.GreaterOrEqual(t, f.Faults(), 1)
}

func TestFirst(t *testing.T) {
	f := Spec{First(2), check.ErrSimulated}.Build()

	assert.Equal(t, f.Try(), check.ErrSimulated)
	assert.Equal(t, 1, f.Calls())
	assert.Equal(t, 1, f.Faults())

	assert.Equal(t, f.Try(), check.ErrSimulated)
	assert.Equal(t, 2, f.Calls())
	assert.Equal(t, 2, f.Faults())

	assert.Nil(t, f.Try())
	assert.Equal(t, 3, f.Calls())
	assert.Equal(t, 2, f.Faults())
}

func TestAfter(t *testing.T) {
	f := Spec{After(1), check.ErrSimulated}.Build()

	assert.Nil(t, f.Try())
	assert.Equal(t, 1, f.Calls())
	assert.Equal(t, 0, f.Faults())

	assert.Equal(t, f.Try(), check.ErrSimulated)
	assert.Equal(t, 2, f.Calls())
	assert.Equal(t, 1, f.Faults())

	assert.Equal(t, f.Try(), check.ErrSimulated)
	assert.Equal(t, 3, f.Calls())
	assert.Equal(t, 2, f.Faults())
}
