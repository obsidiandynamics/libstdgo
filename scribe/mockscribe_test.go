package scribe

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/obsidiandynamics/libstdgo/check"
	"github.com/stretchr/testify/assert"
)

func TestBasicLogging(t *testing.T) {
	m := NewMock()
	l := New(m.Loggers())
	l.SetEnabled(All)

	l.T()("Trace %d %d", 0, 1)
	l.D()("Debug %d %d", 2, 3)
	l.I()("Info %d %d", 4, 5)
	l.W()("Warn %d %d", 6, 7)
	l.E()("Error %d %d", 8, 9)

	t.Log(m.Entries().List())
	assert.Equal(t, 5, len(m.Entries().List()))
	assert.Equal(t, 5, m.Entries().Length())

	m.Entries().Assert(t, Count(5))
	m.Entries().Assert(t, CountAtLeast(4))
	m.Entries().Assert(t, CountAtLeast(5))
	m.Entries().Assert(t, CountAtMost(5))
	m.Entries().Assert(t, CountAtMost(6))

	m.Entries().
		Having(LogLevel(Debug)).
		Assert(t, Count(1))

	m.Entries().
		Having(Not(LogLevel(Debug))).
		Assert(t, Count(4))

	m.Entries().
		Having(Not(LogLevel(Debug))).
		Having(Not(LogLevel(Error))).
		Assert(t, Count(3))

	m.Entries().
		Having(MessageContaining("n")).
		Assert(t, Count(2))

	m.Entries().
		Having(MessageEqual("Info 4 5")).
		Assert(t, Count(1))
}

func TestSceneLogging(t *testing.T) {
	m := NewMock()
	l := New(m.Loggers())
	l.SetEnabled(All)

	testScene := func(name, value string) Scene {
		return Scene{
			Fields: Fields{name: value},
			Err:    check.ErrFault,
		}
	}

	l.I()("Info %d %d", 4, 5)
	l.Capture(testScene("foo", "bar")).W()("Warn %d %d", 5, 6)

	m.Entries().
		Having(LogLevel(Info)).
		Having(ASceneWith(AFieldNamed("foo"))).
		Assert(t, Count(0))

	m.Entries().
		Having(LogLevel(Info)).
		Having(ASceneWith(AField("foo", "bar"))).
		Assert(t, Count(0))

	m.Entries().
		Having(LogLevel(Info)).
		Having(ASceneWith(AnError())).
		Assert(t, Count(0))

	m.Entries().
		Having(ASceneWith(AFieldNamed("foo"))).
		Assert(t, Count(1))

	m.Entries().
		Having(ASceneWith(AField("foo", "bar"))).
		Assert(t, Count(1))

	m.Entries().
		Having(ASceneWith(AField("foo", "other"))).
		Assert(t, Count(0))

	m.Entries().
		Having(ASceneWith(AnError())).
		Assert(t, Count(1))

	// Test the rest of the levels
	m.Reset()
	l.Capture(testScene("alpha", "bravo")).T()("Trace")
	l.Capture(testScene("charlie", "delta")).D()("Debug")
	l.Capture(testScene("echo", "foxtrot")).I()("Info")
	l.Capture(testScene("golf", "hotel")).E()("Error")

	t.Log(m.Entries())
	m.Entries().Having(LogLevel(Trace)).Having(ASceneWith(AField("alpha", "bravo"))).Assert(t, Count(1))
	m.Entries().Having(LogLevel(Debug)).Having(ASceneWith(AField("charlie", "delta"))).Assert(t, Count(1))
	m.Entries().Having(LogLevel(Info)).Having(ASceneWith(AField("echo", "foxtrot"))).Assert(t, Count(1))
	m.Entries().Having(LogLevel(Error)).Having(ASceneWith(AField("golf", "hotel"))).Assert(t, Count(1))

	m.Entries().Having(ASceneWith(Content())).Assert(t, Count(4))
	m.Entries().Having(ASceneWith(Content().Invert())).Assert(t, Count(0))
}

func TestCustomLevel(t *testing.T) {
	const BooYeah Level = 85
	var capture *string
	l := New(LoggerFactories{
		BooYeah: func(level Level, scene Scene) Logger {
			return func(format string, args ...interface{}) {
				c := fmt.Sprintf(format, args...)
				t.Logf("Called with %s", c)
				capture = &c
			}
		},
		All: nopFac,
	})

	l.L(BooYeah)("Yeah %s", "baby")
	assert.Equal(t, "Yeah baby", *capture)
}

func TestDynamicAssertions(t *testing.T) {
	m := NewMock()
	l := New(m.Loggers())

	l.I()("Info %d %d", 4, 5)
	l.W()("Warn %d %d", 5, 6)

	a0 := m.ContainsEntries().Having(LogLevel(Error)).Having(MessageContaining("boom")).Passes(Count(1))
	a1 := m.ContainsEntries().Passes(Count(2))
	a2 := m.ContainsEntries().Passes(Count(3))

	ensureFails := func(a check.Assertion) {
		c := check.NewTestCapture()
		a(c)
		assert.Equal(t, c.Length(), 1)
		c.Reset()
	}

	ensureFails(a0)
	a1(t)
	ensureFails(a2)

	// Log another entries; the assertions should be dynamically recalculated
	l.E()("boom")

	a0(t)
	ensureFails(a1)
	a2(t)
}

func TestMultithreadedLogging(t *testing.T) {
	m := NewMock()
	l := New(m.Loggers())
	l.SetEnabled(All)

	const routines = 100
	wg := sync.WaitGroup{}
	wg.Add(routines)

	levels := []Level{Trace, Debug, Info, Warn, Error}
	for i := 0; i < routines; i++ {
		i := i
		go func() {
			defer wg.Done()
			level := levels[i%5]
			l.L(level)("Logged from %d", i)
		}()
	}

	check.Wait(t, 10*time.Second).UntilAsserted(func(t check.Tester) {
		m.Entries().Having(LogLevel(Debug)).Assert(t, Count(routines/5))
	})
	check.Wait(t, 10*time.Second).UntilAsserted(func(t check.Tester) {
		m.Entries().Assert(t, Count(routines))
	})

	wg.Wait()
	m.Entries().Having(LogLevel(Debug)).Assert(t, Count(routines/5))
	m.Entries().Assert(t, Count(routines))
}

func TestRest(t *testing.T) {
	m := NewMock()
	l := New(m.Loggers())
	l.SetEnabled(All)

	l.T()("Trace %d %d", 0, 1)
	l.D()("Debug %d %d", 2, 3)
	m.Entries().Assert(t, Count(2))
	m.Reset()
	m.Entries().Assert(t, Count(0))

	l.I()("Info %d %d", 4, 5)
	m.Entries().Assert(t, Count(1))
}

func TestAssertionFailures(t *testing.T) {
	m := NewMock()
	l := New(m.Loggers())
	l.SetEnabled(All)

	l.T()("Trace %d %d", 0, 1)
	l.D()("Debug %d %d", 2, 3)
	l.I()("Info %d %d", 4, 5)
	l.W()("Warn %d %d", 6, 7)
	l.E()("Error %d %d", 8, 9)

	c := check.NewTestCapture()

	m.Entries().Assert(c, Count(3))
	c.First().AssertFirstLineEqual(t, "Expected 3 entries; got 5")
	assert.Equal(t, 2, c.First().NumCapturedLines())
	c.Reset()

	m.Entries().Assert(c, CountAtMost(3))
	c.First().AssertFirstLineEqual(t, "Expected at most 3 entries; got 5")
	assert.Equal(t, 2, c.First().NumCapturedLines())
	c.Reset()

	m.Entries().Assert(c, CountAtLeast(7))
	c.First().AssertFirstLineEqual(t, "Expected at least 7 entries; got 5")
	assert.Equal(t, 2, c.First().NumCapturedLines())
	c.Reset()
}
