package check

import "testing"

func TestAppendf(t *testing.T) {
	c := NewTestCapture()

	i := Intercept(c).Mutate(Appendf("A fish named %s", "Wanda"))
	i.Errorf("Some error")

	c.First().AssertFirstLineEqual(t, "Some error A fish named Wanda")
}

func TestThen(t *testing.T) {
	c := NewTestCapture()

	i := Intercept(c).Mutate(Append("First").Then(Append("Second")))
	i.Errorf("Some error")

	c.First().AssertFirstLineEqual(t, "Some error First Second")
}

func TestAddStack(t *testing.T) {
	c := NewTestCapture()

	i := Intercept(c).Mutate(AddStack())
	i.Errorf("Some error")

	c.First().AssertFirstLineEqual(t, "Some error")
	c.First().AssertContains(t, "interceptor_test.go")
}
