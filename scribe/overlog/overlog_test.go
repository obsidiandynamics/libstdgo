package overlog

import (
	"bytes"
	"testing"
	"time"

	"github.com/obsidiandynamics/libstdgo/check"
	"github.com/obsidiandynamics/libstdgo/scribe"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrintfAndRaw(t *testing.T) {
	b := &bytes.Buffer{}
	s := New(Message(), b)
	assert.Equal(t, "", b.String())

	s.Tracef("alpha")
	assert.Equal(t, "alpha\n", b.String())

	s.Debugf("bravo")
	assert.Equal(t, "alpha\nbravo\n", b.String())

	s.Raw(".")
	assert.Equal(t, "alpha\nbravo\n.", b.String())

	s.Raw(".")
	assert.Equal(t, "alpha\nbravo\n..", b.String())

	s.Raw("")
	assert.Equal(t, "alpha\nbravo\n..", b.String())

	s.Infof("charlie %s", "company")
	assert.Equal(t, "alpha\nbravo\n..\ncharlie company\n", b.String())

	s.Warnf("")
	assert.Equal(t, "alpha\nbravo\n..\ncharlie company\n\n", b.String())

	s.Errorf("\n")
	assert.Equal(t, "alpha\nbravo\n..\ncharlie company\n\n\n\n", b.String())
}

func TestTimestamp_fullLayout(t *testing.T) {
	b := &bytes.Buffer{}
	s := New(Format(Timestamp(TimestampLayoutDateTime), Message()), b)
	assert.Equal(t, "", b.String())

	s.Infof("")
	timeBefore := time.Now()
	msg := b.String()
	timestamp, err := time.ParseInLocation(TimestampLayoutDateTime, msg[:len(msg)-2], time.Local)
	timeAfter := time.Now()
	require.Nil(t, err)

	const delta = 5 * time.Second

	if timestamp.Add(time.Millisecond).Sub(timeBefore) < -delta { // millisecond is added because of prior rounding down
		t.Errorf("Expected %v >= %v", timestamp, timeBefore)
	}
	if timestamp.Sub(timeAfter) > delta {
		t.Errorf("Expected %v <= %v", timestamp, timeAfter)
	}
}

func TestLevel(t *testing.T) {
	b := &bytes.Buffer{}
	s := New(Level(), b)

	s.Tracef("irrelevant")
	assert.Equal(t, "TRC\n", b.String())
	b.Reset()

	s.Debugf("irrelevant")
	assert.Equal(t, "DBG\n", b.String())
	b.Reset()

	s.Infof("irrelevant")
	assert.Equal(t, "INF\n", b.String())
	b.Reset()

	s.Warnf("irrelevant")
	assert.Equal(t, "WRN\n", b.String())
	b.Reset()

	s.Errorf("irrelevant")
	assert.Equal(t, "ERR\n", b.String())
	b.Reset()

	const X scribe.Level = 70
	s.With(X, scribe.Scene{})("irrelevant")
	assert.Equal(t, "<ordinal 70>\n", b.String())
	b.Reset()
}

func TestScene(t *testing.T) {
	b := &bytes.Buffer{}
	s := New(Scene(), b)
	s.With(scribe.Info, scribe.Scene{Fields: scribe.Fields{"foo": "bar"}, Err: check.ErrFault})("irrelevant")
	assert.Equal(t, "<foo:bar> <Simulated>\n", b.String())
}

func TestFormat(t *testing.T) {
	b := &bytes.Buffer{}
	s := New(Format(Level(), Message()), b)
	s.With(scribe.Info, scribe.Scene{})("important message %d", 42)
	assert.Equal(t, "INF important message 42\n", b.String())
}
