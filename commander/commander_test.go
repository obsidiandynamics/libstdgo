package commander

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestString(t *testing.T) {
	assert.Equal(t, "Part[Name=some name, Value=some value]", Part{"some name", "some value"}.String())
}

func TestIsFreeForm(t *testing.T) {
	assert.True(t, Part{"", "some value"}.IsFreeForm())
	assert.False(t, Part{"some name", "some value"}.IsFreeForm())
}

func TestDashes(t *testing.T) {
	cases := []struct {
		arg    string
		expect int
	}{
		{"", 0},
		{"a", 0},
		{"-", 0},
		{"--", 0},
		{"---", 0},
		{"---a", 0},
		{"-a", 1},
		{"--a", 2},
	}

	for _, c := range cases {
		assert.Equal(t, c.expect, dashes(c.arg))
	}
}

func TestParse(t *testing.T) {
	cases := []struct {
		cmdArgs []string
		expect  Parts
	}{
		{cmdArgs: []string{},
			expect: Parts{}},
		{cmdArgs: []string{"go"},
			expect: Parts{Part{"", "go"}}},
		{cmdArgs: []string{"go", "test", "-", "--", "---three"},
			expect: Parts{Part{"", "go"}, Part{"", "test"}, Part{"", "-"}, Part{"", "--"}, Part{"", "---three"}}},
		{cmdArgs: []string{"go", "-run", "^TestExample$"},
			expect: Parts{Part{"", "go"}, Part{"run", "^TestExample$"}}},
		{cmdArgs: []string{"go", "--run", "^TestExample$"},
			expect: Parts{Part{"", "go"}, Part{"run", "^TestExample$"}}},
		{cmdArgs: []string{"go", "-run", "^TestExample$", "-foo=bar"},
			expect: Parts{Part{"", "go"}, Part{"run", "^TestExample$"}, Part{"foo", "bar"}}},
		{cmdArgs: []string{"go", "--run", "^TestExample$", "--foo=bar", "trail"},
			expect: Parts{Part{"", "go"}, Part{"run", "^TestExample$"}, Part{"foo", "bar"}, Part{"", "trail"}}},
		{cmdArgs: []string{"go", "-run", "^TestExample$", "-yes"},
			expect: Parts{Part{"", "go"}, Part{"run", "^TestExample$"}, Part{"yes", "true"}}},
		{cmdArgs: []string{"go", "--run", "^TestExample$", "--yes"},
			expect: Parts{Part{"", "go"}, Part{"run", "^TestExample$"}, Part{"yes", "true"}}},
		{cmdArgs: []string{"go", "--yes", "-run", "^TestExample$"},
			expect: Parts{Part{"", "go"}, Part{"yes", "true"}, Part{"run", "^TestExample$"}}},
		{cmdArgs: []string{"go", "-run=^TestExample$"},
			expect: Parts{Part{"", "go"}, Part{"run", "^TestExample$"}}},
		{cmdArgs: []string{"go", "-run=^TestExample$", "-foo=bar"},
			expect: Parts{Part{"", "go"}, Part{"run", "^TestExample$"}, Part{"foo", "bar"}}},
	}

	for _, c := range cases {
		parsed := Parse(c.cmdArgs)
		assert.Equal(t, c.expect, parsed)
	}
}

func TestPartsMap(t *testing.T) {
	mapped := Parse([]string{"go", "--run", "^TestExample$", "--foo=bar", "-run=Another", "trail", "-verbose"}).Mappify()
	assert.Equal(t, PartsMap{
		FreeForm:  []string{"go", "trail"},
		"run":     []string{"^TestExample$", "Another"},
		"foo":     []string{"bar"},
		"verbose": []string{"true"},
	}, mapped)
}

func TestValue(t *testing.T) {
	withoutError := func(value string, err error) string {
		assert.Nil(t, err)
		return value
	}
	withError := func(value string, err error) error {
		assert.NotNil(t, err)
		return err
	}

	mapped := Parse([]string{"go", "--run", "^TestExample$", "--foo=bar", "-run=Another", "trail", "-verbose"}).Mappify()
	assert.Equal(t, errors.New("too many arguments: expected one or none, got 2"), withError(mapped.Value(FreeForm, "")))
	assert.Equal(t, "bar", withoutError(mapped.Value("foo", "")))
	assert.Equal(t, "true", withoutError(mapped.Value("verbose", "false")))
	assert.Equal(t, "some-default", withoutError(mapped.Value("missing", "some-default")))

	// This is the case where we might not care about the error. Should return the first value.
	value, err := mapped.Value(FreeForm, "some-default")
	assert.Equal(t, "go", value)
	assert.NotNil(t, err)
}
