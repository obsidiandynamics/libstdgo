package overlog

import "github.com/obsidiandynamics/libstdgo/scribe"

// Bind creates a direct binding for the given logger.
func Bind(logger Overlog) scribe.LoggerFactories {
	return scribe.LoggerFactories{
		scribe.All: logger.With,
	}
}
