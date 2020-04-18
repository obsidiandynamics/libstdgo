package overlog

import (
	"bytes"
	"testing"

	"github.com/obsidiandynamics/libstdgo/check"
	"github.com/obsidiandynamics/libstdgo/scribe"
	"github.com/stretchr/testify/assert"
)

func TestLogLevels(t *testing.T) {
	buffer := &bytes.Buffer{}
	logger := New(StandardFormat(), buffer)
	s := scribe.New(Bind(logger))

	s.T()("Alpha %d", 1)
	assert.Contains(t, buffer.String(), "TRC Alpha 1")
	buffer.Reset()

	s.D()("Bravo %d", 2)
	assert.Contains(t, buffer.String(), "DBG Bravo 2")
	buffer.Reset()

	s.I()("Charlie %d", 3)
	assert.Contains(t, buffer.String(), "INF Charlie 3")
	buffer.Reset()

	s.W()("Delta %d", 4)
	assert.Contains(t, buffer.String(), "WRN Delta 4")
	buffer.Reset()

	s.Capture(scribe.Scene{
		Fields: scribe.Fields{
			"foo": "bar",
		},
		Err: check.ErrFault}).
		E()("Echo %d", 5)
	assert.Contains(t, buffer.String(), "ERR Echo 5 <foo:bar> <Simulated>")
	buffer.Reset()
}
