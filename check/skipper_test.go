package check

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsNarrowMatch(t *testing.T) {
	cases := []struct {
		arg    string
		expect bool
	}{
		{"", false},
		{"Something", false},
		{"^Example$", true},
		{"(Ex1|Ex2}", false},
		{"^Example.*$", false},
	}

	for _, c := range cases {
		assert.Equal(t, isNarrowMatch(c.arg), c.expect)
	}
}

type skipper struct {
	skipArgs *[]interface{}
}

func (s *skipper) Skip(args ...interface{}) {
	s.skipArgs = &args
}

func TestRunTargetted_private(t *testing.T) {
	cases := []struct {
		args      []string
		expectRun bool
	}{
		{[]string{}, false},
		{[]string{"go", "test", "-run", "^TestExample$", "-foo=bar"}, true},
		{[]string{"go", "test", "-test.run=^TestExample$", "-foo=bar"}, true},
		{[]string{"go", "test", "-run", "^(Example)$"}, true},
		{[]string{"go", "test", "-run=^(Example)$"}, true},
		{[]string{"go", "test", "-run=^(Example)$"}, true},
		{[]string{"go", "test", "-run=^(ExampleA|ExampleB)$"}, false},
		{[]string{"go", "test", "-test.run=^(ExampleA.*)$"}, false},
	}

	for _, c := range cases {
		t := Intercept(t).Mutate(Appendf("case %v", c))
		s := &skipper{}
		ran := false
		runTargetted(s, func() { ran = true }, c.args)
		if c.expectRun {
			assert.Nil(t, s.skipArgs)
		} else {
			if assert.NotNil(t, s.skipArgs) {
				assert.Equal(t, *s.skipArgs, []interface{}{"Skipped"})
			}
		}
		assert.Equal(t, ran, c.expectRun)
	}
}

func TestRunTargetted(t *testing.T) {
	// This test is mainly for coverage.
	RunTargetted(&skipper{}, func() {})
}

func TestRequireLabel_private(t *testing.T) {
	cases := []struct {
		args      []string
		labels    string
		required  string
		expectRun bool
	}{
		{[]string{}, "", "int", false},
		{[]string{"-run=^.*$"}, "", "int", false},
		{[]string{"-run=^.*$"}, "foo", "int", false},
		{[]string{"-run=^TestExample$"}, "", "int", true},
		{[]string{"-run=^.*$"}, "int,foo", "int", true},
		{[]string{"-run=^TestExample$"}, "int,foo", "int", true},
	}

	for _, c := range cases {
		t := Intercept(t).Mutate(Appendf("case %v", c))
		s := &skipper{}
		requireLabel(s, c.required, c.args, func(key string) string { return c.labels })
		if c.expectRun {
			assert.Nil(t, s.skipArgs)
		} else {
			if assert.NotNil(t, s.skipArgs) {
				assert.Equal(t, *s.skipArgs, []interface{}{"Skipped"})
			}
		}
	}
}

func TestRequireLabel(t *testing.T) {
	// This test is mainly for coverage.
	RequireLabel(&skipper{}, "int")
}
