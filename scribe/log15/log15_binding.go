// Package log15 provides a Log15 binding for Scribe.
package log15

import (
	"bytes"
	"fmt"
	"runtime"
	"strings"

	"github.com/go-stack/stack"
	"github.com/inconshreveable/log15"
	"github.com/obsidiandynamics/stdlibgo/arity"
	"github.com/obsidiandynamics/stdlibgo/scribe"
)

// Binding captures the state of the binding, including the underlying logger instance. The
// binding must be closed when the logger is no longer required.
type Binding interface {
	Factories() scribe.LoggerFactories
	Close() error
}

type binding struct {
	dtor   Destructor
	logger log15.Logger
}

// Factories generates the LoggerFactories required to configure Scribe.
func (b *binding) Factories() scribe.LoggerFactories {
	return scribe.LoggerFactories{
		scribe.Trace: func(level scribe.Level, scene scribe.Scene) scribe.Logger {
			return func(format string, args ...interface{}) {
				b.logger.Debug(fmt.Sprintf(format, args...), buildContext(scene)...)
			}
		},
		scribe.Debug: func(level scribe.Level, scene scribe.Scene) scribe.Logger {
			return func(format string, args ...interface{}) {
				b.logger.Debug(fmt.Sprintf(format, args...), buildContext(scene)...)
			}
		},
		scribe.Info: func(level scribe.Level, scene scribe.Scene) scribe.Logger {
			return func(format string, args ...interface{}) {
				ctx := buildContext(scene)
				b.logger.Info(fmt.Sprintf(format, args...), ctx...)
			}
		},
		scribe.Warn: func(level scribe.Level, scene scribe.Scene) scribe.Logger {
			return func(format string, args ...interface{}) {
				b.logger.Warn(fmt.Sprintf(format, args...), buildContext(scene)...)
			}
		},
		scribe.Error: func(level scribe.Level, scene scribe.Scene) scribe.Logger {
			return func(format string, args ...interface{}) {
				b.logger.Error(fmt.Sprintf(format, args...), buildContext(scene)...)
			}
		},
	}
}

// KeyErr is used to key Scene.Err into the custom context.
const KeyErr = "Err"

func buildContext(scene scribe.Scene) []interface{} {
	errCount := 0
	if scene.Err != nil {
		errCount = 1
	}
	length := (len(scene.Fields) + errCount) * 2
	if length == 0 {
		return nil
	}

	ctx := make([]interface{}, length)
	i := 0
	for k, v := range scene.Fields {
		ctx[i] = k
		ctx[i+1] = v
		i += 2
	}
	if scene.Err != nil {
		ctx[i] = KeyErr
		ctx[i+1] = scene.Err
	}
	return ctx
}

// Closes the underlying logger.
func (b *binding) Close() error {
	return b.dtor(b.logger)
}

// Constructor is a way of creating a Log15 logger.
type Constructor func() log15.Logger

// Destructor provides a way of cleaning up
type Destructor func(logger log15.Logger) error

// WithContext returns a constructor for creating a Log15 logger from the given parent
// logger and additional context.
func WithContext(parent log15.Logger, ctx ...interface{}) Constructor {
	return func() log15.Logger {
		return parent.New(ctx...)
	}
}

// WithHandler wraps a given constructor to inject the given handler to a newly created Log15 logger
// before returning it to the caller.
func WithHandler(ctor Constructor, handler log15.Handler) Constructor {
	return func() log15.Logger {
		logger := ctor()
		logger.SetHandler(handler)
		return logger
	}
}

type stackElement struct {
	file string
	line int
}

// Required for stack calibration.
const thisGoFilename = "log15_binding.go"

func dumpStack() []stackElement {
	elements := make([]stackElement, 0, 10)

	i := 1
	for ; ; i++ {
		_, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		fileParts := strings.Split(file, "/")
		fileName := fileParts[len(fileParts)-1]
		elements = append(elements, stackElement{fileName, line})
	}

	return elements
}

// Calibrates the logger stack depth by walking the stack until it reaches an external caller (a site outside of
// this .go file). The result of the calibration is subsequently used to pinpoint the exact call site.
//
// Calibration is required because the stack is polated from the innards of the handler implementation, which is
// called from Log15. We avoid using some constant stack depth that has been derived through trial and error, as
// this would make it brittle to further changes in the internal Log15 implementation.
func calibrate(logger log15.Logger) int {
	orig := logger.GetHandler()
	defer logger.SetHandler(orig)

	depth := -1
	logger.SetHandler(log15.FuncHandler(func(r *log15.Record) error {
		elements := dumpStack()
		for i := len(elements) - 1; i >= 0; i-- {
			if elements[i].file == thisGoFilename {
				depth = i
				break
			}
		}
		return nil
	}))
	logger.Error("irrelevant") // logging something kicks off calibration in the handler
	return depth
}

// NoDestructor is a no-op destructor.
func NoDestructor() Destructor {
	return func(logger log15.Logger) error {
		return nil
	}
}

// Bind makes a new Log15 binding using the given constructor to create the underlying Log15 logger. The returned
// binding must be closed after the logger is no longer required. The closing of the logger is delegated to an
// optional destructor. The destructor will typically close any handlers that require explit disposal.
//
// This implementation uses shimming to realise the binding, having compensated for the call stack depth with the
// underlying logger.
func Bind(ctor Constructor, dtor ...Destructor) Binding {
	logger := ctor()
	depth := calibrate(logger)
	handler := logger.GetHandler()
	logger.SetHandler(log15.FuncHandler(func(r *log15.Record) error {
		r.Call = stack.Caller(depth)
		return handler.Log(r)
	}))

	dtorArg := arity.SoleUntyped(NoDestructor(), dtor).(Destructor)
	return &binding{dtorArg, logger}
}

// FullFormat prints all fields in a log record. Useful for debugging.
type FullFormat struct{}

const timeFormat = "2006-01-02T15:04:05.000"

// Format obtains a formatted representation of the given log record.
func (FullFormat) Format(r *log15.Record) []byte {
	buffer := &bytes.Buffer{}
	time := r.Time.Format(timeFormat)
	kv := formatKV(r.Ctx)
	fmt.Fprint(buffer, time, " ", r.Lvl, " ", r.Call, ": ", r.Msg, " ", kv, "\n")
	return buffer.Bytes()
}

func formatKV(ctx []interface{}) string {
	builder := strings.Builder{}
	length := len(ctx)
	for i := 0; i < length-1; i += 2 {
		builder.WriteString(fmt.Sprint(ctx[i], "=", ctx[i+1]))
		if i != length-1 {
			builder.WriteString(" ")
		}
	}
	return builder.String()
}
