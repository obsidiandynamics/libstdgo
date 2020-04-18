package concurrent

import (
	"context"
	"time"
)

// Forever returns a context that never times out.
func Forever(parent context.Context) (context.Context, context.CancelFunc) {
	return Timeout(parent, Indefinitely)
}

// Timeout returns a context that will expire within the given timeout.
func Timeout(parent context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithDeadline(context.Background(), time.Now().Add(timeout))
}
