package zap

import (
	"bytes"
	"testing"

	"github.com/obsidiandynamics/libstdgo/check"
	"github.com/obsidiandynamics/libstdgo/scribe"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type syncBuffer struct {
	bytes.Buffer
}

func (b *syncBuffer) Sync() error {
	return nil
}

func TestLogLevels(t *testing.T) {
	buffer := &syncBuffer{}
	core := zapcore.NewCore(zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()), buffer, zapcore.DebugLevel)
	zap := zap.New(core).WithOptions(zap.AddCaller())
	s := scribe.New(Bind(zap.Sugar()))
	s.SetEnabled(scribe.All)

	s.T()("Alpha %d", 1)
	assert.Contains(t, buffer.String(), "zap_binding_test.go")
	assert.Contains(t, buffer.String(), "DEBUG")
	assert.Contains(t, buffer.String(), "Alpha 1")
	buffer.Reset()

	s.D()("Bravo %d", 2)
	assert.Contains(t, buffer.String(), "DEBUG")
	assert.Contains(t, buffer.String(), "Bravo 2")
	buffer.Reset()

	s.I()("Charlie %d", 3)
	assert.Contains(t, buffer.String(), "INFO")
	assert.Contains(t, buffer.String(), "Charlie 3")
	buffer.Reset()

	s.W()("Delta %d", 4)
	assert.Contains(t, buffer.String(), "WARN")
	assert.Contains(t, buffer.String(), "Delta 4")
	buffer.Reset()

	s.E()("Echo %d", 5)
	assert.Contains(t, buffer.String(), "ERROR")
	assert.Contains(t, buffer.String(), "Echo 5")
	buffer.Reset()
}

func TestWithScene(t *testing.T) {
	buffer := &syncBuffer{}
	core := zapcore.NewCore(zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()), buffer, zapcore.DebugLevel)
	zap := zap.New(core).WithOptions(zap.AddCaller())
	s := scribe.New(Bind(zap.Sugar()))
	s.SetEnabled(scribe.All)

	s.Capture(scribe.Scene{}).
		I()("Charlie %d", 3)
	assert.Contains(t, buffer.String(), "INFO")
	assert.Contains(t, buffer.String(), "Charlie 3")
	assert.NotContains(t, buffer.String(), "Fields")
	assert.NotContains(t, buffer.String(), "Err")
	buffer.Reset()

	s.Capture(scribe.Scene{Fields: scribe.Fields{"x": "y"}}).
		I()("Charlie %d", 3)
	assert.Contains(t, buffer.String(), "INFO")
	assert.Contains(t, buffer.String(), "Charlie 3")
	assert.Contains(t, buffer.String(), `{"x": "y"}`)
	assert.NotContains(t, buffer.String(), "Err")
	buffer.Reset()

	s.Capture(scribe.Scene{Fields: scribe.Fields{"x": "y"}, Err: check.ErrSimulated}).
		I()("Charlie %d", 3)
	assert.Contains(t, buffer.String(), "INF")
	assert.Contains(t, buffer.String(), `"x": "y"`)
	assert.Contains(t, buffer.String(), `"Err": "simulated"`)
	assert.Contains(t, buffer.String(), "Charlie 3")
	buffer.Reset()
}
