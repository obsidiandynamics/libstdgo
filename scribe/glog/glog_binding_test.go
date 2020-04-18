package glog

import (
	"testing"

	"github.com/obsidiandynamics/libstdgo/scribe"
)

// Just for coverage and to make sure that nothing panics, as Glog does not allow us to assert
// the stream.
func TestLogging(t *testing.T) {
	s := scribe.New(Bind())
	s.T()("Alpha %d", 1)
}
