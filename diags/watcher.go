package diags

import (
	"time"

	"github.com/obsidiandynamics/stdlibgo/scribe"
)

// Watcher contains a timer that fires if an operation fails to complete within a deadline.
type Watcher struct {
	operation string
	duration  time.Duration
	timer     *time.Timer
	done      chan int
}

// Trigger is a function that is fired when a deadline is missed.
type Trigger func(watcher *Watcher)

// Print is a trigger function that will emit a message to the given printf-style logger.
func Print(logger scribe.Logger) Trigger {
	return func(watcher *Watcher) {
		logger("Operation '%s' took longer than %v", watcher.operation, watcher.duration)
	}
}

// Watch creates a Watcher that will fire the specified trigger when the deadline specified by the
// duration argument expires, unless End() is called beforehand.
func Watch(operation string, duration time.Duration, trigger Trigger) *Watcher {
	w := &Watcher{
		operation: operation,
		duration:  duration,
		done:      make(chan int),
	}

	go func() {
		timer := time.NewTimer(duration)
		defer timer.Stop()

		select {
		case <-timer.C:
			trigger(w)
		case <-w.done:
		}
	}()

	return w
}

// End completes the watcher, preventing the trigger from firing, unless it has already done so.
func (w *Watcher) End() {
	close(w.done)
}
