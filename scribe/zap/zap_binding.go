// Package zap provides a Zap binding for Scribe.
package zap

import (
	"fmt"

	"github.com/obsidiandynamics/stdlibgo/scribe"
	"go.uber.org/zap"
)

// KeyErr is used to key Scene.Err into the custom logging context.
const KeyErr = "Err"

func enrich(sug *zap.SugaredLogger, scene scribe.Scene) *zap.SugaredLogger {
	for k, v := range scene.Fields {
		sug = sug.With(k, fmt.Sprint(v))
	}
	if scene.Err != nil {
		sug = sug.With(KeyErr, scene.Err.Error())
	}
	return sug
}

// Bind creates a Zap binding for a given sugared logger.
func Bind(logger *zap.SugaredLogger) scribe.LoggerFactories {
	return scribe.LoggerFactories{
		scribe.Trace: func(level scribe.Level, scene scribe.Scene) scribe.Logger {
			return enrich(logger, scene).Debugf
		},
		scribe.Debug: func(level scribe.Level, scene scribe.Scene) scribe.Logger {
			return enrich(logger, scene).Debugf
		},
		scribe.Info: func(level scribe.Level, scene scribe.Scene) scribe.Logger {
			return enrich(logger, scene).Infof
		},
		scribe.Warn: func(level scribe.Level, scene scribe.Scene) scribe.Logger {
			return enrich(logger, scene).Warnf
		},
		scribe.Error: func(level scribe.Level, scene scribe.Scene) scribe.Logger {
			return enrich(logger, scene).Errorf
		},
	}
}
