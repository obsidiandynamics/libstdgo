package diags

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDumpAllStacks(t *testing.T) {
	dump := DumpAllStacks()
	assert.Contains(t, dump, "diags_test.go")
}
