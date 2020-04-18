package log15

import (
	"bytes"
	"testing"

	"github.com/inconshreveable/log15"
	"github.com/obsidiandynamics/libstdgo/check"
	"github.com/obsidiandynamics/libstdgo/scribe"
	"github.com/stretchr/testify/assert"
)

func TestLogLevels(t *testing.T) {
	buffer := &bytes.Buffer{}
	ctor := WithHandler(WithContext(log15.Root()), log15.StreamHandler(buffer, FullFormat{}))
	binding := Bind(ctor)
	s := scribe.New(binding.Factories())
	s.SetEnabled(scribe.All)

	s.T()("Alpha %d", 1)
	assert.Contains(t, buffer.String(), "dbug")
	assert.Contains(t, buffer.String(), "log15_binding_test")
	assert.Contains(t, buffer.String(), "Alpha 1")
	buffer.Reset()

	s.D()("Bravo %d", 2)
	assert.Contains(t, buffer.String(), "dbug")
	assert.Contains(t, buffer.String(), "log15_binding_test")
	assert.Contains(t, buffer.String(), "Bravo 2")
	buffer.Reset()

	s.I()("Charlie %d", 3)
	assert.Contains(t, buffer.String(), "info")
	assert.Contains(t, buffer.String(), "log15_binding_test")
	assert.Contains(t, buffer.String(), "Charlie 3")
	buffer.Reset()

	s.W()("Delta %d", 4)
	assert.Contains(t, buffer.String(), "warn")
	assert.Contains(t, buffer.String(), "log15_binding_test")
	assert.Contains(t, buffer.String(), "Delta 4")
	buffer.Reset()

	s.E()("Echo %d", 5)
	assert.Contains(t, buffer.String(), "eror")
	assert.Contains(t, buffer.String(), "log15_binding_test")
	assert.Contains(t, buffer.String(), "Echo 5")
	buffer.Reset()

	err := binding.Close()
	assert.Nil(t, err)
}

func TestWithScene_fieldsAndError(t *testing.T) {
	buffer := &bytes.Buffer{}
	ctor := WithHandler(WithContext(log15.Root()), log15.StreamHandler(buffer, FullFormat{}))
	binding := Bind(ctor)
	s := scribe.New(binding.Factories())
	s.SetEnabled(scribe.All)

	s.Capture(scribe.Scene{}).
		I()("Charlie %d", 3)
	assert.Contains(t, buffer.String(), "info")
	assert.Contains(t, buffer.String(), "Charlie 3")
	assert.NotContains(t, buffer.String(), "Fields")
	assert.NotContains(t, buffer.String(), "Err")
	buffer.Reset()

	s.Capture(scribe.Scene{Fields: scribe.Fields{"x": "y"}}).
		I()("Charlie %d", 3)
	assert.Contains(t, buffer.String(), "info")
	assert.Contains(t, buffer.String(), "Charlie 3")
	assert.Contains(t, buffer.String(), "x=y")
	assert.NotContains(t, buffer.String(), "Err")
	buffer.Reset()

	s.Capture(scribe.Scene{Fields: scribe.Fields{"x": "y"}, Err: check.ErrSimulated}).
		I()("Charlie %d", 3)
	assert.Contains(t, buffer.String(), "info")
	assert.Contains(t, buffer.String(), "Charlie 3")
	assert.Contains(t, buffer.String(), "x=y")
	assert.NotContains(t, buffer.String(), "Error=\"simulated\"")
	buffer.Reset()
}

func TestDestructor(t *testing.T) {
	ctor := WithContext(log15.Root())

	dtorInvoked := false
	dtor := func(logger log15.Logger) error {
		dtorInvoked = true
		return nil
	}

	binding := Bind(ctor, dtor)
	assert.False(t, dtorInvoked)
	err := binding.Close()
	assert.Nil(t, err)
	assert.True(t, dtorInvoked)
}
