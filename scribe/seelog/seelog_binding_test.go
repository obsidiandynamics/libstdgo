package seelog

import (
	"bytes"
	"io"
	"testing"

	"github.com/cihub/seelog"
	"github.com/obsidiandynamics/libstdgo/check"
	"github.com/obsidiandynamics/libstdgo/scribe"
	"github.com/stretchr/testify/assert"
)

func createBindingForWriter(w io.Writer) Binding {
	return Bind(func() seelog.LoggerInterface {
		const formatStr = "%Time %Date %LEV %File:%Line: %Msg"
		logger, err := seelog.LoggerFromWriterWithMinLevelAndFormat(w, seelog.TraceLvl, formatStr)
		if err != nil {
			panic(err)
		}
		return logger
	})
}

func TestLogLevels(t *testing.T) {
	buffer := &bytes.Buffer{}
	binding := createBindingForWriter(buffer)
	defer binding.Close()
	s := scribe.New(binding.Factories())
	s.SetEnabled(scribe.All)

	s.T()("Alpha %d", 1)
	assert.Contains(t, buffer.String(), "TRC")
	assert.Contains(t, buffer.String(), "Alpha 1")
	buffer.Reset()

	s.D()("Bravo %d", 2)
	assert.Contains(t, buffer.String(), "DBG")
	assert.Contains(t, buffer.String(), "Bravo 2")
	buffer.Reset()

	s.I()("Charlie %d", 3)
	assert.Contains(t, buffer.String(), "INF")
	assert.Contains(t, buffer.String(), "Charlie 3")
	buffer.Reset()

	s.W()("Delta %d", 4)
	assert.Contains(t, buffer.String(), "WRN")
	assert.Contains(t, buffer.String(), "Delta 4")
	buffer.Reset()

	s.E()("Echo %d", 5)
	assert.Contains(t, buffer.String(), "ERR")
	assert.Contains(t, buffer.String(), "Echo 5")
	buffer.Reset()

	binding.Close()
}

func TestWithScene(t *testing.T) {
	buffer := &bytes.Buffer{}
	binding := createBindingForWriter(buffer)
	defer binding.Close()
	s := scribe.New(binding.Factories())

	s.Capture(scribe.Scene{}).
		I()("Charlie %d", 3)
	assert.Contains(t, buffer.String(), "INF")
	assert.Contains(t, buffer.String(), "Charlie 3")
	assert.NotContains(t, buffer.String(), "Fields")
	assert.NotContains(t, buffer.String(), "Err")
	buffer.Reset()

	s.Capture(scribe.Scene{Fields: scribe.Fields{"x": "y"}}).
		I()("Charlie %d", 3)
	assert.Contains(t, buffer.String(), "INF")
	assert.Contains(t, buffer.String(), "Charlie 3 <x:y>")
	assert.NotContains(t, buffer.String(), "Err")
	buffer.Reset()

	s.Capture(scribe.Scene{Fields: scribe.Fields{"x": "y"}, Err: check.ErrSimulated}).
		I()("Charlie %d", 3)
	assert.Contains(t, buffer.String(), "INF")
	assert.Contains(t, buffer.String(), "Charlie 3 <x:y> <simulated>")
	buffer.Reset()
}
