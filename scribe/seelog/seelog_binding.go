// Package seelog provides a Seelog binding for Scribe.
package seelog

import (
	"fmt"

	"github.com/cihub/seelog"
	"github.com/obsidiandynamics/stdlibgo/scribe"
)

// Binding captures the state of the binding, including the underlying logger instance. The
// binding must be closed when the logger is no longer required.
type Binding interface {
	Factories() scribe.LoggerFactories
	Close()
}

type binding struct {
	logger seelog.LoggerInterface
}

// Factories generates the LoggerFactories required to configure Scribe.
func (b *binding) Factories() scribe.LoggerFactories {
	hook := scribe.AppendScene()
	return scribe.LoggerFactories{
		scribe.Trace: func(level scribe.Level, scene scribe.Scene) scribe.Logger {
			return func(format string, args ...interface{}) {
				enrich(b.logger, scene).Trace(fmtMessage(hook, level, scene, format, args...))
			}
		},
		scribe.Debug: func(level scribe.Level, scene scribe.Scene) scribe.Logger {
			return func(format string, args ...interface{}) {
				enrich(b.logger, scene).Debug(fmtMessage(hook, level, scene, format, args...))
			}
		},
		scribe.Info: func(level scribe.Level, scene scribe.Scene) scribe.Logger {
			return func(format string, args ...interface{}) {
				enrich(b.logger, scene).Info(fmtMessage(hook, level, scene, format, args...))
			}
		},
		scribe.Warn: func(level scribe.Level, scene scribe.Scene) scribe.Logger {
			return func(format string, args ...interface{}) {
				enrich(b.logger, scene).Warn(fmtMessage(hook, level, scene, format, args...))
			}
		},
		scribe.Error: func(level scribe.Level, scene scribe.Scene) scribe.Logger {
			return func(format string, args ...interface{}) {
				enrich(b.logger, scene).Error(fmtMessage(hook, level, scene, format, args...))
			}
		},
	}
}

// Closes the underlying logger.
func (b *binding) Close() {
	b.logger.Close()
}

// KeyErr is used to key Scene.Err into the custom context.
const KeyErr = "Err"

func enrich(logger seelog.LoggerInterface, scene scribe.Scene) seelog.LoggerInterface {
	m := map[string]interface{}{}
	for k, v := range scene.Fields {
		m[k] = v
	}
	if scene.Err != nil {
		m[KeyErr] = scene.Err.Error()
	}
	return logger
}

func fmtMessage(hook scribe.Hook, level scribe.Level, scene scribe.Scene, format string, args ...interface{}) string {
	hook(level, &scene, &format, &args)
	msg := fmt.Sprintf(format, args...) + "\n"
	return msg
}

// Constructor is a way of creating a Seelog logger.
type Constructor func() seelog.LoggerInterface

// Bind makes a new Seelog binding using the given constructor to create the underlying Seelog logger. The returned
// binding must be closed after the logger is no longer required.
//
// This implementation uses shimming to realise the binding, having compensated for the call stack depth with the
// underlying logger.
func Bind(ctor Constructor) Binding {
	logger := ctor()
	logger.SetAdditionalStackDepth(1)
	return &binding{logger}
}
