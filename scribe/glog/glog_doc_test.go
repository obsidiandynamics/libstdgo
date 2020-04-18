package glog

import (
	"testing"

	"github.com/obsidiandynamics/libstdgo/check"
	"github.com/obsidiandynamics/libstdgo/scribe"
)

func Example() {
	s := scribe.New(Bind())

	// Do some logging
	s.I()("Important application message")
}

func TestExample(t *testing.T) {
	check.RunTargetted(t, Example)
}
