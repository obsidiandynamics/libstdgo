package scribe

import (
	"bytes"
	"fmt"
	"log"

	"github.com/obsidiandynamics/libstdgo/arity"
)

/*
Bindings for built-in loggers, as well as support for shimming and hook-based transforms.
*/

// StandardBinding creates a shim-based binding for log.Printf(), appending scene information. This should
// be the default binding used by applications that have not yet adopted a logging framework.
func StandardBinding() LoggerFactories {
	return ShimFacs(BindLogPrintf(), AppendScene())
}

// BindFmt creates a binding for the logger used by fmt. There are several issues with fmt:
//   1. It's Printf has return values, making it incompatible with Scribe.
//   2. It does not add a newline.
//
// As a result of these limitations, this binding is implemented as a shim, rather than a function
// pointer. In practice, this is of no consequence, as fmt does not care about caller site information.
func BindFmt() LoggerFactories {
	return LoggerFactories{
		All: Fac(func(format string, args ...interface{}) {
			msg := fmt.Sprintf(format, args...)
			fmt.Println(msg)
		}),
	}
}

// BindLogPrintf creates a pass-through binding for log.Printf(). An optional Logger instance can be specified;
// if omitted, the standard logger will be used.
func BindLogPrintf(logger ...*log.Logger) LoggerFactories {
	l := arity.SoleUntyped(nil, logger)
	var printf Logger
	if l != nil {
		printf = l.(*log.Logger).Printf
	} else {
		printf = log.Printf
	}
	return LoggerFactories{
		All: Fac(printf),
	}
}

// Hook is a function that can inspect log arguments before they are passed to the underlying logger, and
// potentially modify these arguments.
//
// The supplied log level cannot be modified, as log factories are provided on a per log level basis.
// The other arguments can be modified by assignment through dereferencing:
//  *scene = Scene{...}
//  *format = "new format with %s %s"
//  *args = []interface{}{"new", "args"}
type Hook func(level Level, scene *Scene, format *string, args *[]interface{})

// AppendScene is a hook that appends the contents of the captured scene after the formatted log message.
func AppendScene() Hook {
	return func(level Level, scene *Scene, format *string, args *[]interface{}) {
		buffer := &bytes.Buffer{}
		buffer.WriteString(fmt.Sprint(fmt.Sprintf(*format, *args...)))
		WriteScene(buffer, *scene)
		msg := buffer.String()
		*format = "%s"
		*args = []interface{}{msg}
	}
}

// Space appends a whitespace character to the given buffer if the latter is non-empty. This function
// is used to separate fields.
func Space(buffer *bytes.Buffer) {
	if buffer.Len() > 0 {
		fmt.Fprint(buffer, " ")
	}
}

// WriteScene is a utility for compactly writing scene contents to an output writer.
func WriteScene(buffer *bytes.Buffer, scene Scene) {
	if len(scene.Fields) > 0 {
		Space(buffer)
		buffer.Write([]byte("<"))
		numFields := len(scene.Fields)
		i := 0
		for k, v := range scene.Fields {
			buffer.Write([]byte(k))
			buffer.Write([]byte(":"))
			buffer.Write([]byte(fmt.Sprint(v)))
			if i < numFields-1 {
				buffer.Write([]byte(" "))
			}
			i++
		}
		buffer.Write([]byte(">"))
	}

	if scene.Err != nil {
		Space(buffer)
		buffer.Write([]byte("<"))
		buffer.Write([]byte(scene.Err.Error()))
		buffer.Write([]byte(">"))
	}
}

// ShimFacs applies a shim to all factories in facs, using the given hook, returning an equivalent map of shimmed factories.
func ShimFacs(facs LoggerFactories, hook Hook) LoggerFactories {
	shimmedFacs := LoggerFactories{}
	for k, v := range facs {
		shimmedFacs[k] = ShimFac(v, hook)
	}
	return shimmedFacs
}

// ShimFac applies a shim to fac, using the given hook. The result is a LoggerFactory that applies the hook before invoking
// the shimmed logger.
//
// Note: shimming is an intrusive process that changes the call site from the perspective of the underlying logger. Shimming
// and hooks within Scribe should only be used for debugging, or to enhance loggers that don't natively support features such
// as structured logging. Where possible, use the native hooks provided by your chosen logging framework.
func ShimFac(fac LoggerFactory, hook Hook) LoggerFactory {
	return func(level Level, scene Scene) Logger {
		return func(format string, args ...interface{}) {
			hook(level, &scene, &format, &args)
			fac(level, scene)(format, args...)
		}
	}
}
