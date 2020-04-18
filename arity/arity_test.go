package arity

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/obsidiandynamics/stdlibgo/check"
)

func TestRepack(t *testing.T) {
	cases := []struct {
		input  interface{}
		expect []interface{}
		err    error
	}{
		{[...]int{0, 1, 2}, []interface{}{0, 1, 2}, nil},
		{[]int{0, 1, 2}, []interface{}{0, 1, 2}, nil},
		{[]int{}, []interface{}{}, nil},
		{[][]int{{0, 1}, {2, 3}}, []interface{}{[]int{0, 1}, []int{2, 3}}, nil},
		{4, nil, fmt.Errorf("Unsupported type int")},
	}

	for _, c := range cases {
		t := check.Intercept(t).Mutate(check.Appendf("For case %v", c))
		actualVal, actualErr := Listify(c.input)
		if c.expect != nil {
			assert.Equal(t, c.expect, actualVal)
			assert.Nil(t, actualErr)
		} else {
			assert.Nil(t, actualVal)
			assert.Equal(t, c.err, actualErr)
		}
	}
}

func TestRepackOrPanic_success(t *testing.T) {
	orig := []int{0, 1, 2}
	repacked := ListifyOrPanic(orig)
	assert.ElementsMatch(t, orig, repacked)
}

func TestRepackOrPanic_panic(t *testing.T) {
	check.ThatPanicsAsExpected(t, check.ErrorWithValue("Unsupported type int"), func() {
		ListifyOrPanic(0)
	})
}

func TestSoleUntyped_zeroLength(t *testing.T) {
	args := [...]rune{}
	assert.Equal(t, 'a', SoleUntyped('a', args).(rune))
}

func TestSoleUntyped_oneLength(t *testing.T) {
	args := [...]rune{'b'}
	assert.Equal(t, 'b', SoleUntyped('a', args).(rune))
}

func TestSoleUntyped_tooMany(t *testing.T) {
	args := [...]rune{'b', 'c'}
	check.ThatPanicsAsExpected(t, check.ErrorWithValue("Expected at most 1 argument(s), got 2"), func() {
		SoleUntyped('a', args)
	})
}

func TestOptionalUntyped(t *testing.T) {
	const noError = ""
	cases := []struct {
		offset    int
		limit     int
		def       interface{}
		args      interface{}
		expectVal interface{}
		expectErr string
	}{
		{0, 0, 'd', [...]rune{'a', 'b'}, nil, errLimitLessThanOne},
		{-1, 3, 'd', [...]rune{'a', 'b'}, nil, errOffsetOutsideRange},
		{3, 3, 'd', [...]rune{'a', 'b'}, nil, errOffsetOutsideRange},
		{0, 1, 'd', [...]rune{}, 'd', noError},
		{0, 1, 'd', [...]rune{'a'}, 'a', noError},
		{0, 2, 'd', [...]rune{}, 'd', noError},
		{0, 2, 'd', [...]rune{'a'}, 'a', noError},
		{1, 2, 'd', [...]rune{'a'}, 'd', noError},
		{1, 2, 'b', [...]rune{'a', 'b'}, 'b', noError},
	}

	for _, c := range cases {
		t := check.Intercept(t).Mutate(check.Appendf("\nFor case %v", c))
		if c.expectErr == noError {
			check.ThatDoesNotPanic(t, func() {
				val := OptionalUntyped(c.offset, c.limit, c.def, c.args)
				assert.Equal(t, c.expectVal, val)
			})
		} else {
			check.ThatPanicsAsExpected(t, check.ErrorWithValue(c.expectErr), func() {
				OptionalUntyped(c.offset, c.limit, c.def, c.args)
			})
		}
	}
}
