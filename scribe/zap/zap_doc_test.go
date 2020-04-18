package zap

import (
	"testing"

	"github.com/obsidiandynamics/stdlibgo/check"
	"github.com/obsidiandynamics/stdlibgo/scribe"
	"go.uber.org/zap"
)

func Example() {
	zap, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	s := scribe.New(Bind(zap.Sugar()))

	// Do some logging
	s.I()("Important application message")
}

func TestExample(t *testing.T) {
	check.RunTargetted(t, Example)
}
