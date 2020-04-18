package concurrent

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var zeroTime = time.Unix(0, 0)

func truncate(t time.Time) time.Time {
	return t.Truncate(time.Nanosecond)
}

func TestCas(t *testing.T) {
	cas := timeCas{NewAtomicCounter(0)}
	assert.Equal(t, cas.get(), zeroTime)

	replacement := time.Now()
	assert.False(t, cas.compareAndSwap(replacement, replacement))
	assert.Equal(t, cas.get(), zeroTime)

	assert.True(t, cas.compareAndSwap(zeroTime, replacement))
	assert.Equal(t, cas.get(), truncate(replacement))

	another := replacement.Add(time.Hour)
	called := false
	setter := func() {
		called = true
	}
	cas.ifSwapped(another, another, setter)
	assert.False(t, called)

	cas.ifSwapped(replacement, another, setter)
	assert.True(t, called)
}

func TestDeadline(t *testing.T) {
	d := NewDeadline(1 * time.Hour)
	called := false
	setter := func() {
		called = true
	}
	assert.Equal(t, zeroTime, d.Last())
	assert.True(t, d.Lapsed())

	assert.True(t, d.TryRun(setter))
	assert.True(t, called)
	assert.NotEqual(t, zeroTime, d.Last())
	assert.False(t, d.Lapsed())

	called = false
	assert.False(t, d.TryRun(setter))
	assert.False(t, called)

	const grace = int64(5 * time.Second)
	assert.Greater(t, int64(d.Elapsed()), -grace)
}

func TestDeadlineMove(t *testing.T) {
	d := NewDeadline(1 * time.Hour)
	assert.True(t, d.Lapsed())
	d.Move(time.Now())
	assert.False(t, d.Lapsed())

	called := false
	setter := func() {
		called = true
	}
	assert.False(t, d.TryRun(setter))
	assert.False(t, called)

	d.Move(zeroTime)
	assert.Equal(t, zeroTime, d.Last())
	assert.True(t, d.TryRun(setter))
	assert.True(t, called)
}
