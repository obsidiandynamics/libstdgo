package logrus

import (
	"bytes"
	"context"
	"testing"

	"github.com/obsidiandynamics/stdlibgo/check"
	"github.com/obsidiandynamics/stdlibgo/scribe"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestLogLevels_traceEnabled(t *testing.T) {
	buffer := &bytes.Buffer{}
	lr := logrus.New()
	lr.SetOutput(buffer)
	lr.SetLevel(logrus.TraceLevel)
	s := scribe.New(Bind(lr))

	s.T()("Alpha %d", 1)
	assert.Contains(t, buffer.String(), "level=trace")
	assert.Contains(t, buffer.String(), "Alpha 1")
	buffer.Reset()

	s.D()("Bravo %d", 2)
	assert.Contains(t, buffer.String(), "level=debug")
	assert.Contains(t, buffer.String(), "Bravo 2")
	buffer.Reset()

	s.I()("Charlie %d", 3)
	assert.Contains(t, buffer.String(), "level=info")
	assert.Contains(t, buffer.String(), "Charlie 3")
	buffer.Reset()

	s.W()("Delta %d", 4)
	assert.Contains(t, buffer.String(), "level=warn")
	assert.Contains(t, buffer.String(), "Delta 4")
	buffer.Reset()

	s.E()("Echo %d", 5)
	assert.Contains(t, buffer.String(), "level=error")
	assert.Contains(t, buffer.String(), "Echo 5")
	buffer.Reset()
}

func TestLogLevels_panicEnabled(t *testing.T) {
	buffer := &bytes.Buffer{}
	lr := logrus.New()
	lr.SetOutput(buffer)
	lr.SetLevel(logrus.PanicLevel)
	s := scribe.New(Bind(lr))

	s.T()("Alpha %d", 1)
	assert.Empty(t, buffer.String())

	s.D()("Bravo %d", 2)
	assert.Empty(t, buffer.String())

	s.I()("Charlie %d", 3)
	assert.Empty(t, buffer.String())

	s.W()("Delta %d", 4)
	assert.Empty(t, buffer.String())

	s.E()("Echo %d", 5)
	assert.Empty(t, buffer.String())
}

func TestWithScene_fieldsAndError(t *testing.T) {
	buffer := &bytes.Buffer{}
	lr := logrus.New()
	lr.SetOutput(buffer)
	s := scribe.New(Bind(lr))

	s.Capture(scribe.Scene{}).
		I()("Charlie %d", 3)
	assert.Contains(t, buffer.String(), "level=info")
	assert.Contains(t, buffer.String(), "Charlie 3")
	assert.NotContains(t, buffer.String(), "Fields")
	assert.NotContains(t, buffer.String(), "Err")
	buffer.Reset()

	s.Capture(scribe.Scene{Fields: scribe.Fields{"x": "y"}}).
		I()("Charlie %d", 3)
	assert.Contains(t, buffer.String(), "level=info")
	assert.Contains(t, buffer.String(), "Charlie 3")
	assert.Contains(t, buffer.String(), "x=y")
	assert.NotContains(t, buffer.String(), "Err")
	buffer.Reset()

	s.Capture(scribe.Scene{Fields: scribe.Fields{"x": "y"}, Err: check.ErrFault}).
		I()("Charlie %d", 3)
	assert.Contains(t, buffer.String(), "level=info")
	assert.Contains(t, buffer.String(), "Charlie 3")
	assert.Contains(t, buffer.String(), "x=y")
	assert.NotContains(t, buffer.String(), "Error=\"Simulated\"")
	buffer.Reset()
}

type captureHook struct {
	levels []logrus.Level
	entry  *logrus.Entry
}

func (h *captureHook) Levels() []logrus.Level {
	return h.levels
}

func (h *captureHook) Fire(entry *logrus.Entry) error {
	h.entry = entry
	return nil
}

func TestWithScene_context(t *testing.T) {
	buffer := &bytes.Buffer{}
	lr := logrus.New()
	lr.SetOutput(buffer)
	s := scribe.New(Bind(lr))

	h := &captureHook{levels: []logrus.Level{logrus.InfoLevel}}
	lr.AddHook(h)

	ctx := context.Background()
	s.Capture(scribe.Scene{Ctx: ctx}).
		I()("Charlie %d", 3)

	assert.NotNil(t, h.entry)
	assert.Equal(t, h.entry.Context, ctx)
}
