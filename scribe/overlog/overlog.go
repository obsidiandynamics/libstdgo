// Package overlog provides a reference logging implementation for Scribe. Besides being a minimal
// logger, Overlog provides support for capturing site-specific metadata and logging from concurrent
// applications, preventing the interleaving of logs across goroutine calls. Overlog also
// supports logging of raw strings, bypassing the formatter.
package overlog

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/obsidiandynamics/stdlibgo/arity"
	"github.com/obsidiandynamics/stdlibgo/scribe"
)

// Overlog is a synchronized logger backed by an io.Writer, suitable for use in concurrent applications
// where an unsynchronized logger would result in interleaved log entries.
type Overlog interface {
	With(level scribe.Level, scene scribe.Scene) scribe.Logger
	Raw(str string)
	Tracef(format string, args ...interface{})
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

type overlog struct {
	lock      sync.Mutex
	writer    io.Writer
	formatter Formatter
	last      byte
}

// Event captures attributes of a single log record.
type Event struct {
	Timestamp time.Time
	Message   string
	Level     scribe.Level
	Scene     scribe.Scene
}

// Formatter specifies how a logging event should be rendered in a given output buffer.
type Formatter func(buffer *bytes.Buffer, event Event)

// Append writes a string into the buffer. If the buffer is non-empty, a leading space is inserted before
// the string is written.
func Append(buffer *bytes.Buffer, str string) {
	scribe.Space(buffer)
	fmt.Fprint(buffer, str)
}

// Format composes multiple formatters into one.
func Format(formatters ...Formatter) Formatter {
	return func(buffer *bytes.Buffer, event Event) {
		for _, f := range formatters {
			f(buffer, event)
		}
	}
}

// StandardFormat produces a formatter that includes all conventional elements — the timestamp, log level, message,
// and the scene contents.
func StandardFormat() Formatter {
	return Format(Timestamp(), Level(), Message(), Scene())
}

const (
	// TimestampLayoutDateTime is a full layout containing both a date and a time portion.
	TimestampLayoutDateTime = "2006-01-02 15:04:05.000"

	// TimestampLayoutTimeOnly contains only the time.
	TimestampLayoutTimeOnly = "15:04:05.000"

	// TimestampLayoutDefault is the default layout applied in the formatter returned by Timestamp().
	TimestampLayoutDefault = TimestampLayoutTimeOnly
)

// Timestamp is a formatter that prints the timestamp of the log event using the layout supplied. If
// no layout is supplied, the TimestampLayoutDefault is used.
func Timestamp(layout ...string) Formatter {
	l := arity.SoleUntyped(TimestampLayoutDefault, layout).(string)
	return func(buffer *bytes.Buffer, event Event) {
		Append(buffer, event.Timestamp.Format(l))
	}
}

// Level is a formatter that prints the level of the log event.
func Level() Formatter {
	return func(buffer *bytes.Buffer, event Event) {
		nameAbbr, _ := scribe.LevelNameAbbreviated(event.Level)
		Append(buffer, nameAbbr)
	}
}

// Message is a formatter that prints the formatted message.
func Message() Formatter {
	return func(buffer *bytes.Buffer, event Event) {
		Append(buffer, event.Message)
	}
}

// Scene is a formatter that prints the elements of the scene.
func Scene() Formatter {
	return func(buffer *bytes.Buffer, event Event) {
		scribe.WriteScene(buffer, event.Scene)
	}
}

// New creates a synchronized logger backed by a given writer. If unspecified, os.Stdout will
// be used.
func New(formatter Formatter, writer ...io.Writer) Overlog {
	w := arity.SoleUntyped(os.Stdout, writer).(io.Writer)
	return &overlog{sync.Mutex{}, w, formatter, '\n'}
}

// State returns a printf-style logger that pipes entries to the underlying writer, followed by a newline. If an
// unterminated line exists from a previous write, it will be closed off with a newline before the new entry
// is written.
func (o *overlog) With(level scribe.Level, scene scribe.Scene) scribe.Logger {
	return func(format string, args ...interface{}) {
		msg := fmt.Sprintf(format, args...)
		buffer := &bytes.Buffer{}
		o.formatter(buffer, Event{time.Now(), msg, level, scene})
		fmt.Fprintln(buffer)

		o.lock.Lock()
		defer o.lock.Unlock()
		if o.last != '\n' {
			fmt.Fprintln(o.writer)
		}
		io.Copy(o.writer, buffer)
		o.last = '\n'
	}
}

// Raw writes a raw string to the logger without invoking the formatter and without appending a newline.
func (o *overlog) Raw(str string) {
	o.lock.Lock()
	defer o.lock.Unlock()
	fmt.Fprint(o.writer, str)
	length := len(str)
	if length != 0 {
		o.last = str[length-1]
	}
}

// Tracef is a convenience for With(scribe.Trace, scribe.Scene{}).
func (o *overlog) Tracef(format string, args ...interface{}) {
	o.With(scribe.Trace, scribe.Scene{})(format, args...)
}

// Debugf is a convenience for With(scribe.Debug, scribe.Scene{}).
func (o *overlog) Debugf(format string, args ...interface{}) {
	o.With(scribe.Debug, scribe.Scene{})(format, args...)
}

// Infof is a convenience for With(scribe.Info, scribe.Scene{}).
func (o *overlog) Infof(format string, args ...interface{}) {
	o.With(scribe.Info, scribe.Scene{})(format, args...)
}

// Warnf is a convenience for With(scribe.Warn, scribe.Scene{}).
func (o *overlog) Warnf(format string, args ...interface{}) {
	o.With(scribe.Warn, scribe.Scene{})(format, args...)
}

// Errorf is a convenience for With(scribe.Error, scribe.Scene{}).
func (o *overlog) Errorf(format string, args ...interface{}) {
	o.With(scribe.Error, scribe.Scene{})(format, args...)
}
