package concurrent

import (
	"math"
	"time"

	"github.com/obsidiandynamics/libstdgo/arity"
)

// Indefinitely is a constant that represents the longest allowable duration (approx. 290 years).
// It is used when an arbitrarily long timeout is needed.
const Indefinitely = math.MaxInt64 * time.Nanosecond

func optional(def time.Duration, args ...time.Duration) time.Duration {
	return arity.SoleUntyped(def, args).(time.Duration)
}
