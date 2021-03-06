package diags

import (
	"testing"
	"time"

	"github.com/obsidiandynamics/libstdgo/check"
	"github.com/obsidiandynamics/libstdgo/concurrent"
	"github.com/obsidiandynamics/libstdgo/scribe"
	"github.com/stretchr/testify/assert"
)

func TestWatch_ended(t *testing.T) {
	triggered := concurrent.NewAtomicCounter()
	trigger := func(watcher *Watcher) {
		triggered.Set(1)
	}

	w := Watch("op", time.Hour, trigger)
	defer w.End()
	time.Sleep(1 * time.Millisecond)
	assert.Equal(t, 0, triggered.GetInt())
}

func TestWatch_triggered(t *testing.T) {
	triggered := concurrent.NewAtomicCounter()
	trigger := func(watcher *Watcher) {
		triggered.Set(1)
	}

	w := Watch("op", time.Millisecond, trigger)
	defer w.End()
	check.Wait(t, 10*time.Second).UntilAsserted(func(t check.Tester) {
		assert.Equal(t, 1, triggered.GetInt())
	})
}

func TestPrint(t *testing.T) {
	m := scribe.NewMock()
	scr := scribe.New(m.Factories())

	w := Watch("op", time.Millisecond, Print(scr.W()))
	defer w.End()
	check.Wait(t, 10*time.Second).UntilAsserted(m.ContainsEntries().
		Having(scribe.LogLevel(scribe.Warn)).
		Having(scribe.MessageEqual("Operation 'op' took longer than 1ms")).
		Passes(scribe.Count(1)))
}
