package log15

import (
	"testing"

	"github.com/inconshreveable/log15"
	"github.com/obsidiandynamics/libstdgo/check"
	"github.com/obsidiandynamics/libstdgo/scribe"
)

func Example() {
	binding := Bind(WithContext(log15.Root()))
	s := scribe.New(binding.Factories())

	// Do some logging
	s.I()("Important application message")

	// Eventually, when the logger is no longer required...
	err := binding.Close()
	if err != nil {
		panic(err)
	}
}

func TestExample(t *testing.T) {
	check.RunTargetted(t, Example)
}
