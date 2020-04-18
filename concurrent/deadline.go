package concurrent

import (
	"time"
)

// Allows for atomic compare-and-swap operations on non-monotonic timestamps.
type timeCas struct {
	time AtomicCounter
}

func (t *timeCas) get() time.Time {
	return time.Unix(0, t.time.Get())
}

func (t *timeCas) compareAndSwap(expected time.Time, replacement time.Time) bool {
	return t.time.CompareAndSwap(expected.UnixNano(), replacement.UnixNano())
}

func (t *timeCas) ifSwapped(expected time.Time, replacement time.Time, f func()) bool {
	if t.compareAndSwap(expected, replacement) {
		f()
		return true
	}
	return false
}

func (t *timeCas) set(time time.Time) {
	t.time.Set(time.UnixNano())
}

// Deadline tracks the time a task was last run and conditionally runs a task if the deadline has lapsed.
//
// Deadline is thread-safe.
type Deadline interface {
	TryRun(f func()) bool
	Elapsed() time.Duration
	Lapsed() bool
	Move(new time.Time)
	Last() time.Time
}

type deadline struct {
	lastRun  timeCas
	interval time.Duration
}

// NewDeadline creates a new Deadline with the specified interval.
func NewDeadline(interval time.Duration) Deadline {
	return &deadline{
		lastRun: timeCas{
			time: NewAtomicCounter(0),
		},
		interval: interval,
	}
}

// TryRun conditionally runs the given function if the deadline object has not been exercised
// for a period that exceeds its set interval. Returns true if the function was executed.
func (d *deadline) TryRun(f func()) bool {
	if now, last := time.Now(), d.Last(); now.Sub(last) > d.interval {
		return d.lastRun.ifSwapped(last, now, f)
	}
	return false
}

// Last returns the timestamp of the last run. If no prior run occurred, the Unix epoch timestamp
// given by time.Unix(0, 0) is returned.
func (d *deadline) Last() time.Time {
	return d.lastRun.get()
}

// Elapsed returns the duration since the last time the deadline was exercised.
func (d *deadline) Elapsed() time.Duration {
	return time.Now().Sub(d.Last())
}

// Lapsed returns true if the deadline has lapsed.
func (d *deadline) Lapsed() bool {
	return time.Now().Sub(d.Last()) > d.interval
}

// Move the timestamp of the last run to the new time.
func (d *deadline) Move(new time.Time) {
	d.lastRun.set(new)
}
