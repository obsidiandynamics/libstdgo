package glog

import (
	"testing"

	"github.com/obsidiandynamics/stdlibgo/check"
	"github.com/obsidiandynamics/stdlibgo/scribe"
)

func Example() {
	s := scribe.New(Bind())

	// Do some logging
	s.I()("Important application message")
}

func TestExample(t *testing.T) {
	check.RunTargetted(t, Example)
}
