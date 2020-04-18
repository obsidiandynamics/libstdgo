package scribe

import (
	"fmt"
	"testing"

	"github.com/obsidiandynamics/libstdgo/check"
	"github.com/stretchr/testify/assert"
)

type logCapture struct {
	scene *Scene
	msg   *string
}

func (c *logCapture) capturing() LoggerFactory {
	return func(level Level, scene Scene) Logger {
		c.scene = &scene
		return func(format string, args ...interface{}) {
			out := fmt.Sprintf(format, args...)
			c.msg = &out
		}
	}
}

func (c *logCapture) reset() {
	c.scene = nil
	c.msg = nil
}

func TestLevelSpec_String(t *testing.T) {
	assert.Contains(t, Levels[Off].String(), "Off")
}

func TestLevelNameAbbreviated(t *testing.T) {
	nameAbbr, err := LevelNameAbbreviated(Info)
	assert.Equal(t, "INF", nameAbbr)
	assert.Nil(t, err)

	nameAbbr, err = LevelNameAbbreviated(75)
	assert.Equal(t, "<ordinal 75>", nameAbbr)
	assert.NotNil(t, err)
}

func TestParseLevelName(t *testing.T) {
	cases := []struct {
		in        string
		wantSpec  LevelSpec
		wantError string
	}{
		{in: "All", wantSpec: Levels[All], wantError: ""},
		{in: "Trace", wantSpec: Levels[Trace], wantError: ""},
		{in: "Off", wantSpec: Levels[Off], wantError: ""},
		{in: "Foo", wantSpec: LevelSpec{}, wantError: "No level specification for name 'Foo'"},
	}

	for _, c := range cases {
		gotSpec, gotErr := ParseLevelName(c.in)
		assert.Equal(t, c.wantSpec, gotSpec)
		if c.wantError != "" {
			if assert.NotNil(t, gotErr) {
				assert.Equal(t, c.wantError, gotErr.Error())
			}
		} else {
			assert.Nil(t, gotErr)
		}
	}
}

func TestBasicInit(t *testing.T) {
	c := logCapture{}
	l := New(LoggerFactories{All: c.capturing()})

	l.L(Off)("should not appear")
	assertNoCaptures(t, c)

	l.I()("Pi is %f", 3.14)
	assertCaptured(t, Scene{}, "Pi is 3.140000", c)
}

func TestMultipleLevels(tst *testing.T) {
	t := logCapture{}
	d := logCapture{}
	i := logCapture{}
	w := logCapture{}
	e := logCapture{}

	l := New(LoggerFactories{
		Trace: t.capturing(),
		Debug: d.capturing(),
		Info:  i.capturing(),
		Warn:  w.capturing(),
		Error: e.capturing(),
	})
	l.SetEnabled(All)
	assert.Equal(tst, All, l.Enabled())

	l.L(Off)("Nothing")
	assertNoCaptures(tst, t, d, i, w, e)

	l.T()("Tracing")
	assertCaptured(tst, Scene{}, "Tracing", t)
	assertNoCaptures(tst, d, i, w, e)
	t.reset()

	l.D()("Debugging")
	assertCaptured(tst, Scene{}, "Debugging", d)
	assertNoCaptures(tst, t, i, w, e)
	d.reset()

	l.I()("Informing")
	assertCaptured(tst, Scene{}, "Informing", i)
	assertNoCaptures(tst, t, d, w, e)
	i.reset()

	l.W()("Warning")
	assertCaptured(tst, Scene{}, "Warning", w)
	assertNoCaptures(tst, t, d, i, e)
	w.reset()

	l.E()("Erring")
	assertCaptured(tst, Scene{}, "Erring", e)
	assertNoCaptures(tst, t, d, i, w)
	e.reset()
}

func TestDefaultEnabledLevels(tst *testing.T) {
	const X Level = Trace - 1

	x := logCapture{}
	t := logCapture{}
	d := logCapture{}
	i := logCapture{}
	w := logCapture{}
	e := logCapture{}

	l := New(LoggerFactories{
		X:     x.capturing(),
		Trace: t.capturing(),
		Debug: d.capturing(),
		Info:  i.capturing(),
		Warn:  w.capturing(),
		Error: e.capturing(),
	})
	assert.Equal(tst, Trace, l.Enabled())

	l.L(Off)("Nothing")
	assertNoCaptures(tst, x, t, d, i, w, e)

	l.L(X)("Nothing")
	assertNoCaptures(tst, x, t, d, i, w, e)

	l.L(Trace)("Something")
	assertCaptured(tst, Scene{}, "Something", t)
	assertNoCaptures(tst, x, d, i, w, e)
}

func assertCaptured(t *testing.T, expScene Scene, expMsg string, capture logCapture) {
	assert.NotNil(t, capture.msg)
	assert.Equal(t, expScene, *capture.scene)
	assert.Equal(t, expMsg, *capture.msg)
}

func assertNoCaptures(t *testing.T, captures ...logCapture) {
	for _, capture := range captures {
		assert.Nil(t, capture.scene)
		assert.Nil(t, capture.msg)
	}
}

func TestMissingLevel(t *testing.T) {
	l := New(LoggerFactories{All: nopFac})
	check.ThatPanicsAsExpected(t, check.ErrorWithValue("Missing logger factory for level <ordinal 80>"), func() {
		logger := l.L(80)
		t.Log(logger)
	})
}

func TestInitWithoutDefault(t *testing.T) {
	check.ThatPanicsAsExpected(t, check.ErrorWithValue("Missing logger factory for level Trace; no default has been provided"), func() {
		New(LoggerFactories{
			Debug: nopFac,
			Info:  nopFac,
			Warn:  nopFac,
			Error: nopFac,
		})
	})

	check.ThatPanicsAsExpected(t, check.ErrorWithValue("Missing logger factory for level Error; no default has been provided"), func() {
		New(LoggerFactories{
			Trace: nopFac,
			Debug: nopFac,
			Info:  nopFac,
			Warn:  nopFac,
		})
	})
}

func TestName(t *testing.T) {
	cases := []struct {
		in            Level
		expectedName  string
		expectedError error
	}{
		{All, "All", nil},
		{Trace, "Trace", nil},
		{Debug, "Debug", nil},
		{Info, "Info", nil},
		{Warn, "Warn", nil},
		{Error, "Error", nil},
		{11, "<ordinal 11>", fmt.Errorf("No level for ordinal 11")},
	}

	for _, c := range cases {
		name, err := LevelName(c.in)
		assert.Equal(t, name, c.expectedName)
		assert.Equal(t, err, c.expectedError)
	}
}
