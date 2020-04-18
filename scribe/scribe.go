/*
Package scribe represents a functional abstraction for unifying loggers. It is not a façade (à la SLF4J), in
that it does not transform from one API to another. Instead, Scribe uses function pointers to expose a standard
printf-style API to the application. In that sense, Scribe is more of a mechanism for standardising and organising
logger implementations. It works with any logger that exposes a printf-style contract.

Scribe supports log enrichment, allowing the user to provide additional contextual
metadata. (This is called a Scene.) It achieves the same goal as a traditional façade, without creating a layer
of indirection and increasing the depth of the call stack. Therefore, the original call site information is
preserved. (By contrast, a naive façade/adapter/shim injects itself into the call stack, reporting the wrong
file name and line number to the logger.)

Scribe is thread-safe; multiple goroutines may use the same instance.
*/
package scribe

import (
	"context"
	"fmt"
)

// Level of logging. The lowest ordinal corresponds to the most fine-grained level. By convention, a level
// subsumes all levels above it, meaning that if a level L is enabled, then any level M, where M > L, is also
// enabled.
type Level uint8

// We allocate ordinals by hand rather than use a iota to ensure that we can add more levels in the future without
// breaking compatibility. In addition, the user may specify their own log level, provided it conforms to the
// Level data type.
const (
	// All is a symbolic value for the lowest possible level. It does not actually get logged, but is useful for
	// addressing all levels above it (for example, to enable all logging).
	All Level = 0

	// Trace is the most fine-grained level that actually gets logged.
	Trace Level = 10

	// Debug level.
	Debug Level = 20

	// Info level.
	Info Level = 30

	// Warn level
	Warn Level = 40

	// Error is the most coarse-grained level that actually gets logged.
	Error Level = 50

	// Off is a symbolic value for the highest possible level. It does not actually get logged, but is useful for
	// addressing all levels below it (for example, to disable all logging).
	Off Level = 200

	// ...
	// Reserving symbolic values 201-255 in case they are needed later.
)

// DefaultEnabledLevel specifies the level of logging which is enabled by default. This includes all levels
// having higher ordinals.
const DefaultEnabledLevel = Trace

// LevelSpec describes a log level.
type LevelSpec struct {
	Level       Level
	Name        string
	Abbreviated string
}

// String obtains a textual representation of a LevelSpec.
func (ls LevelSpec) String() string {
	return fmt.Sprint("LevelSpec[Level=", ls.Level, ", Name=", ls.Name, ", Abbreviated=", ls.Abbreviated, "]")
}

// Levels lists built-in levels (including the two symbolic ones, All and Off), mapping them to their descriptions.
//
// Custom levels can be defined (provided they conform to the Level data type); the knowledge of such levels will
// remain within the confines of the user application.
var Levels = map[Level]LevelSpec{
	All:   {All, "All", "ALL"},
	Trace: {Trace, "Trace", "TRC"},
	Debug: {Debug, "Debug", "DBG"},
	Info:  {Info, "Info", "INF"},
	Warn:  {Warn, "Warn", "WRN"},
	Error: {Error, "Error", "ERR"},
	Off:   {Off, "Off", "WRN"},
}

// String obtains a textual depiction of the log level.
func (l Level) String() string {
	name, _ := LevelName(l)
	return name
}

func noLevelForOrdinal(level Level) (string, error) {
	return fmt.Sprintf("<ordinal %d>", level), fmt.Errorf("no level for ordinal %d", level)
}

// LevelName gets the name of the given level, if one is known. An error is returned if the level is not among the
// known Levels map. In the error case, the name will contain its ordinal.
func LevelName(level Level) (string, error) {
	if spec, ok := Levels[level]; ok {
		return spec.Name, nil
	}
	return noLevelForOrdinal(level)
}

// LevelNameAbbreviated gives the abbreviated name for a given level. An error is returned if the level is not among the
// known Levels map. In the error case, the name will contain its ordinal.
func LevelNameAbbreviated(level Level) (string, error) {
	if spec, ok := Levels[level]; ok {
		return spec.Abbreviated, nil
	}
	return noLevelForOrdinal(level)
}

// ParseLevelName locates a LevelSpec for a given name string, returning an error if none could be matched.
func ParseLevelName(name string) (LevelSpec, error) {
	for _, spec := range Levels {
		if name == spec.Name {
			return spec, nil
		}
	}
	return LevelSpec{}, fmt.Errorf("no level specification for name '%s'", name)
}

// Fields is a free-form set of attributes that can be captured as part of a Scene, supporting
// log enrichment and structured logging.
type Fields map[string]interface{}

// Scene captures additional metadata from the call site that is forwarded to the logger. This is used to
// facilitate structured logging, pass contexts onto loggers, communicate application errors, and so forth.
type Scene struct {
	Fields Fields
	Ctx    context.Context
	Err    error
}

// String obtains a textual representation of a Scene.
func (s Scene) String() string {
	return fmt.Sprint("Scene[Fields=", s.Fields, ", Ctx=", s.Ctx, ", Err=", s.Err, "]")
}

// IsSet returns true if the scene, meaning it has at least one field specified, a context set or carries an error.
func (s Scene) IsSet() bool {
	return len(s.Fields) > 0 || s.Ctx != nil || s.Err != nil
}

// LoggerFactory specifies the behaviour for constructing a logger instance. The log factory is called upon each time
// a logger is requested — every time an application needs to log something.
type LoggerFactory func(level Level, scene Scene) Logger

// Logger is a single-use function for logging output. It is meant to be used at the point where the application is ready
// to submit the log message. This is not a constraint as such; it allows for the capture of contextual scene metadata.
//
// Implementations may reuse a single instance of a logger function if they don't care about scene capture or deal with
// race-prone state.
type Logger func(format string, args ...interface{})

// LoggerFactories is used to configure Scribe, specifying a LogFactory for each supported level.
type LoggerFactories map[Level]LoggerFactory

type sceneStub struct {
	s     *scribe
	scene Scene
}

// StdLogAPI represents the standard way of interacting with Scribe.
type StdLogAPI interface {
	L(level Level) Logger
	T() Logger
	D() Logger
	I() Logger
	W() Logger
	E() Logger
}

// Scribe is the starting point for invoking a logger. There is no concept of a default Scribe logger; one
// one must be constructed explicitly using NewScribe() and handed to the application. (Or the application
// may instantiate a singleton logger and use the same Scribe instance throughout.)
type Scribe interface {
	StdLogAPI
	Enabled() Level
	SetEnabled(level Level)
	Capture(scene Scene) StdLogAPI
}

type scribe struct {
	facs    LoggerFactories
	enabled Level
}

var nopFac = Fac(Nop)

// Fac wraps a given reusable logger function in a factory. Useful for simple loggers that don't care about scene
// metadata, and willing to recycle the same logging function.
func Fac(logger Logger) LoggerFactory {
	return func(_ Level, _ Scene) Logger {
		return logger
	}
}

// Nop is a no-op logger function.
func Nop(_ string, _ ...interface{}) {}

// New constructs a Scribe instance from the given facs configuration.
//
// The supplied facs maps a supported log level to a corresponding LoggerFactory. Factories may be supplied individually
// for each supported log level. The special All level can be used to configure a default factory that will be applied
// to all built-in log levels that have not been explicitly configured in facs. If one of the built-in levels
// is not configured, and no default LogFactory is specified for All, this function will panic.
//
// Custom log levels are supported by supplying a mapping for a custom Level. However, the default LogFactory specified
// for the All level does not apply to custom levels. In other words, each custom level requires an explicit LogFactory.
func New(facs LoggerFactories) Scribe {
	var defFac = facs[All]

	expandedFacs := LoggerFactories{}
	for k, v := range facs {
		expandedFacs[k] = v
	}

	if _, ok := expandedFacs[Off]; !ok {
		expandedFacs[Off] = nopFac
	}

	for _, l := range Levels {
		if l.Level == Off || l.Level == All {
			continue
		}
		if _, ok := expandedFacs[l.Level]; !ok {
			if defFac == nil {
				panic(fmt.Errorf("missing logger factory for level %s; no default has been provided", l.Name))
			}
			expandedFacs[l.Level] = defFac
		}
	}

	return &scribe{expandedFacs, DefaultEnabledLevel}
}

// Capture contextual scene metadata for passing onto the underlying logger, in preparation for a
// subsequent logging call.
func (s *scribe) Capture(scene Scene) StdLogAPI {
	return &sceneStub{s, scene}
}

// Enabled returns the most fine-grained log level that is enabled. By implication, all levels that are coarser
// than the returned level are also enabled.
func (s *scribe) Enabled() Level {
	return s.enabled
}

// SetEnabled enables logging at the given level. By implication, all levels that are coarser
// than the supplied level are also enabled.
func (s *scribe) SetEnabled(level Level) {
	s.enabled = level
}

// L obtains a logger function for the supplied level. This method is the long form of calling T(), D(), I(), etc.,
// and is useful when the level is selected dynamically (as opposed to being embedded in code).
//
// L also allows for custom log levels that don't have a corresponding short-form method.
func (s *scribe) L(level Level) Logger {
	return s.fac(level)(level, Scene{})
}

// T is the short form of L(Trace), returning a logger for the Trace level.
func (s *scribe) T() Logger { return s.L(Trace) }

// D is the short form of L(Debug), returning a logger for the Debug level.
func (s *scribe) D() Logger { return s.L(Debug) }

// I is the short form of L(Info), returning a logger for the Info level.
func (s *scribe) I() Logger { return s.L(Info) }

// W is the short form of L(Warn), returning a logger for the Warn level.
func (s *scribe) W() Logger { return s.L(Warn) }

// E is the short form of L(Error), returning a logger for the Error level.
func (s *scribe) E() Logger { return s.L(Error) }

// Retrieves a LoggerFactory for the specified level.
func (s *scribe) fac(level Level) LoggerFactory {
	if level < s.enabled {
		return nopFac
	}
	if loggerFac, ok := s.facs[level]; ok {
		return loggerFac
	}

	// An invalid level was supplied
	panic(fmt.Errorf("missing logger factory for level %s", level.String()))
}

func (ss *sceneStub) L(level Level) Logger {
	return ss.s.fac(level)(level, ss.scene)
}

// T is the short form of L(Trace), returning a logger for the Trace level.
func (ss *sceneStub) T() Logger { return ss.L(Trace) }

// D is the short form of L(Debug), returning a logger for the Debug level.
func (ss *sceneStub) D() Logger { return ss.L(Debug) }

// I is the short form of L(Info), returning a logger for the Info level.
func (ss *sceneStub) I() Logger { return ss.L(Info) }

// W is the short form of L(Warn), returning a logger for the Warn level.
func (ss *sceneStub) W() Logger { return ss.L(Warn) }

// E is the short form of L(Error), returning a logger for the Error level.
func (ss *sceneStub) E() Logger { return ss.L(Error) }
