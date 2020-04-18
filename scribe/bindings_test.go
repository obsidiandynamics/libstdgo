package scribe

import (
	"bytes"
	"fmt"
	"log"
	"testing"

	"github.com/obsidiandynamics/libstdgo/check"
	"github.com/stretchr/testify/assert"
)

func TestBindFmtPrintf(t *testing.T) {
	l := New(BindFmt())
	l.SetEnabled(Debug)
	l.D()("Debugging %s", "something")
}

func TestStandardBinding(t *testing.T) {
	l := New(StandardBinding())
	l.SetEnabled(Debug)
	l.D()("Debugging %s", "something")
}

func TestStandardBinding_customWriter(t *testing.T) {
	buffer := bytes.Buffer{}
	logger := log.New(&buffer, "", log.Llongfile)
	l := New(BindLogPrintf(logger))
	l.SetEnabled(Debug)
	l.D()("Debugging %s", "something")
	assert.Contains(t, buffer.String(), "Debugging something")
	assert.Contains(t, buffer.String(), "bindings_test.go")
}

func TestAppendScene(t *testing.T) {
	cases := []struct {
		format string
		args   []interface{}
		scene  Scene
		expect string
	}{
		{
			format: "%d %d",
			args:   []interface{}{1, 2},
			scene:  Scene{},
			expect: "1 2",
		},
		{
			format: "%d %d",
			args:   []interface{}{1, 2},
			scene:  Scene{Fields: Fields{}},
			expect: "1 2",
		},
		{
			format: "%d %d",
			args:   []interface{}{1, 2},
			scene:  Scene{Fields: Fields{"alpha": "bravo"}},
			expect: "1 2 <alpha:bravo>",
		},
		{
			format: "%d %d",
			args:   []interface{}{1, 2},
			scene:  Scene{Err: check.ErrSimulated},
			expect: "1 2 <simulated>",
		},
		{
			format: "%d %d",
			args:   []interface{}{1, 2},
			scene:  Scene{Fields: Fields{"alpha": "bravo"}, Err: check.ErrSimulated},
			expect: "1 2 <alpha:bravo> <simulated>",
		},
	}

	appendScene := AppendScene()
	for _, c := range cases {
		format := c.format
		args := make([]interface{}, len(c.args))
		copy(args, c.args)
		appendScene(Info, &c.scene, &format, &args)
		t := check.Intercept(t).Mutate(check.Appendf("case %v", c))
		msg := fmt.Sprintf(format, args...)
		assert.Equal(t, c.expect, msg)
	}
}

// Done as a separate test because map iteration order is non-deterministic, which means we need
// assert either possibility.
func TestAppendScene_twoFields(t *testing.T) {
	scene := Scene{Fields: Fields{"alpha": "bravo", "charlie": "delta"}}
	format := "%d %d"
	args := []interface{}{1, 2}
	AppendScene()(Info, &scene, &format, &args)
	msg := fmt.Sprintf(format, args...)
	assert.Contains(t, msg, "1 2")
	assert.Contains(t, msg, "alpha:bravo")
	assert.Contains(t, msg, "charlie:delta")
}

func TestShimFacs_withAppendScene(t *testing.T) {
	captured := ""
	logger := func(format string, args ...interface{}) {
		captured = fmt.Sprintf(format, args...)
	}
	facs := LoggerFactories{
		Info: Fac(logger),
	}
	shimmed := ShimFacs(facs, AppendScene())
	assert.Len(t, shimmed, 1)

	shimmed[Info](Info, Scene{Err: check.ErrSimulated})("one %d %d", 2, 3)
	assert.Equal(t, "one 2 3 <simulated>", captured)
}

func TestShimFac_mutateAllCallArgs(t *testing.T) {
	var capturedScene Scene
	var capturedFormat string
	var capturedArgs []interface{}

	fac := func(level Level, scene Scene) Logger {
		capturedScene = scene
		return func(format string, args ...interface{}) {
			capturedFormat = format
			capturedArgs = args
		}
	}

	substituteScene := Scene{Err: check.ErrSimulated}
	shimmed := ShimFac(fac, func(level Level, scene *Scene, format *string, args *[]interface{}) {
		*scene = substituteScene
		*format = "tomarf"
		*args = []interface{}{"argX", "argY"}
	})
	shimmed(Info, Scene{})("format", "arg0")

	assert.Equal(t, substituteScene, capturedScene)
	assert.Equal(t, "tomarf", capturedFormat)
	assert.Equal(t, []interface{}{"argX", "argY"}, capturedArgs)
}
