package logrus

import (
	"testing"

	"github.com/obsidiandynamics/stdlibgo/check"
	"github.com/obsidiandynamics/stdlibgo/scribe"
	"github.com/sirupsen/logrus"
)

func Example() {
	lr := logrus.StandardLogger()
	s := scribe.New(Bind(lr))

	// Do some logging
	s.I()("Important application message")
}

func TestExample(t *testing.T) {
	check.RunTargetted(t, Example)
}
