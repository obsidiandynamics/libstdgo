package check

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Basic capture of a single Errorf call
func TestBasicCapture(t *testing.T) {
	c := NewTestCapture()
	first := c.First()
	first.AssertNil(t)
	assert.Nil(t, first.Captured())
	assert.Equal(t, 0, first.NumCapturedLines())
	assert.Nil(t, first.CapturedLines())

	c.Errorf("alpha %s", "bravo")
	first = c.First()
	first.AssertFirstLineEqual(t, "alpha bravo")
	first.AssertNotNil(t)
	assert.Equal(t, 1, first.NumCapturedLines())
	assert.NotNil(t, first.CapturedLines())
	assert.Equal(t, "alpha bravo", *first.Captured())
}

// Inductive test of itself.
func TestCaptureSelf(t *testing.T) {
	g := NewTestCapture() // working capture
	c := NewTestCapture() // capture under test

	// Test with empty capture
	c.First().AssertNil(g)
	g.First().AssertNil(t)

	c.First().AssertNotNil(g)
	g.First().AssertNotNil(t)
	g.First().AssertFirstLineEqual(t, "Expected not nil")
	g.Reset()

	c.First().AssertFirstLineEqual(g, "alpha bravo")
	g.First().AssertNotNil(t)
	g.First().AssertFirstLineEqual(t, "Expected 'alpha bravo'; got nil")
	g.Reset()

	c.First().AssertFirstLineContains(g, "alpha bravo")
	g.First().AssertNotNil(t)
	g.First().AssertFirstLineEqual(t, "Expected string containing 'alpha bravo'; got nil")
	g.Reset()

	c.First().AssertContains(g, "alpha bravo")
	g.First().AssertNotNil(t)
	g.First().AssertFirstLineEqual(t, "Expected string containing 'alpha bravo'; got nil")
	g.Reset()

	// Test with non-empty capture
	c.Errorf("alpha %s", "bravo")

	c.First().AssertNil(g)
	g.First().AssertNotNil(t)
	g.First().AssertFirstLineEqual(t, "Expected nil; got 'alpha bravo'")
	g.Reset()

	c.First().AssertNotNil(g)
	g.First().AssertNil(t)

	c.First().AssertFirstLineEqual(g, "alpha bravo")
	g.First().AssertNil(t)

	c.First().AssertFirstLineEqual(g, "charlie delta")
	g.First().AssertNotNil(t)
	g.First().AssertFirstLineEqual(t, "Expected 'charlie delta'; got 'alpha bravo'")
	g.Reset()

	c.First().AssertFirstLineContains(g, "charlie delta")
	g.First().AssertNotNil(t)
	g.First().AssertFirstLineEqual(t, "Expected string containing 'charlie delta'; got 'alpha bravo'")
	g.Reset()

	c.First().AssertContains(g, "charlie delta")
	g.First().AssertNotNil(t)
	g.First().AssertFirstLineEqual(t, "Expected string containing 'charlie delta'; got 'alpha bravo'")
	g.Reset()
}

// Capturing multiple calls to Errorf
func TestMultipleCaptures(t *testing.T) {
	c := NewTestCapture()
	c.Errorf("One %d", 1)
	c.Errorf("Two %d", 2)

	assert.Equal(t, 2, c.Length())
	c.Captures()[0].AssertFirstLineEqual(t, "One 1")
	c.Captures()[1].AssertFirstLineEqual(t, "Two 2")
}
