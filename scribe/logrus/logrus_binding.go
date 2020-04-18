// Package logrus provides a Logrus binding for Scribe.
package logrus

import (
	"context"

	"github.com/obsidiandynamics/stdlibgo/arity"
	"github.com/obsidiandynamics/stdlibgo/scribe"
	"github.com/sirupsen/logrus"
	lr "github.com/sirupsen/logrus"
)

// The standard Logrus API. Used by both Logger and Entry.
type logAPI interface {
	WithError(err error) *lr.Entry
	WithFields(fields lr.Fields) *lr.Entry
	WithContext(ctx context.Context) *lr.Entry
	Tracef(format string, args ...interface{})
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

func enrich(api logAPI, scene scribe.Scene) logAPI {
	if len(scene.Fields) > 0 {
		api = api.WithFields((lr.Fields)(scene.Fields))
	}
	if scene.Ctx != nil {
		api = api.WithContext(scene.Ctx)
	}
	if scene.Err != nil {
		api = api.WithError(scene.Err)
	}
	return api
}

// Bind creates a Logrus binding for an optional logger. If omitted, the logger defaults to
// StandardLogger.
func Bind(logger ...*lr.Logger) scribe.LoggerFactories {
	l := arity.SoleUntyped(lr.StandardLogger(), logger).(*lr.Logger)
	return scribe.LoggerFactories{
		scribe.Trace: func(level scribe.Level, scene scribe.Scene) scribe.Logger {
			if l.IsLevelEnabled(logrus.TraceLevel) {
				return enrich(l, scene).Tracef
			} else {
				return scribe.Nop
			}
		},
		scribe.Debug: func(level scribe.Level, scene scribe.Scene) scribe.Logger {
			if l.IsLevelEnabled(logrus.DebugLevel) {
				return enrich(l, scene).Debugf
			} else {
				return scribe.Nop
			}
		},
		scribe.Info: func(level scribe.Level, scene scribe.Scene) scribe.Logger {
			if l.IsLevelEnabled(logrus.InfoLevel) {
				return enrich(l, scene).Infof
			} else {
				return scribe.Nop
			}
		},
		scribe.Warn: func(level scribe.Level, scene scribe.Scene) scribe.Logger {
			if l.IsLevelEnabled(logrus.WarnLevel) {
				return enrich(l, scene).Warnf
			} else {
				return scribe.Nop
			}
		},
		scribe.Error: func(level scribe.Level, scene scribe.Scene) scribe.Logger {
			if l.IsLevelEnabled(logrus.ErrorLevel) {
				return enrich(l, scene).Errorf
			} else {
				return scribe.Nop
			}
		},
	}
}
