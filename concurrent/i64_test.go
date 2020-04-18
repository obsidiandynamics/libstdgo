package concurrent

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCounterNot(t *testing.T) {
	assert.True(t, I64Not(I64Equal(5))(6))
}

func TestCounterEqual(t *testing.T) {
	assert.True(t, I64Equal(5)(5))
}

func TestCounterLessThan(t *testing.T) {
	assert.True(t, I64LessThan(5)(4))
}

func TestCounterGreaterThan(t *testing.T) {
	assert.True(t, I64GreaterThan(5)(6))
}
