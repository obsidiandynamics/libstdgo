package check

import "fmt"

// Mutation represents a transformation of the captured Tester.Errorf() invocation, where the formatted
// message is fed to a Mutation function, and the output is some transformation of the original.
type Mutation func(original string) string

// Append adds a string to the end of the original captured message. The appended suffixed is delimited from the
// original capture with a single whitespace character.
func Append(suffix interface{}) Mutation {
	return func(original string) string {
		return fmt.Sprint(original, " ", suffix)
	}
}

// Appendf adds a (printf-style) formmatted message to the end of the capture. This is a convenience
// for calling Append(), having applied fmt.Sprintf().
func Appendf(format string, args ...interface{}) Mutation {
	return Append(fmt.Sprintf(format, args...))
}

// AddStack adds a stack trace to the end of the capture.
func AddStack() Mutation {
	return func(original string) string {
		return fmt.Sprint(original, PrintStack(2))
	}
}

// Then chains mutations, feeding the outcome of the 'before' mutation to the after' mutation. It allows you to do things like:
//  Append("Hello").Then(AddStack())
func (before Mutation) Then(after Mutation) Mutation {
	return func(original string) string {
		return after(before(original))
	}
}

type interceptor struct {
	t Tester
	m Mutation
}

// Interceptor represents a mechanism for transforming the result of invoking Tester.Errorf().
// This is useful when you need to modify the output of a Tester, for instance, to enrich it
// with additional information that is not available to the assertion framework
// at the point where an assertion fails.
type Interceptor interface {
	Tester
}

// InterceptorStub is an element in a fluid chain.
type InterceptorStub struct {
	t Tester
}

// Intercept starts a fluid chain for specifying testing mutations.
func Intercept(t Tester) InterceptorStub {
	return InterceptorStub{t}
}

// Mutate sets up the given mutation that will be triggered on Tester.Errorf(). To chain mutations, use the Then() method.
func (is InterceptorStub) Mutate(m Mutation) Interceptor {
	return &interceptor{is.t, m}
}

// Errorf passes a formatted error message to the underlying Tester, having first applied a mutation.
func (i *interceptor) Errorf(format string, args ...interface{}) {
	original := fmt.Sprintf(format, args...)
	mutated := i.m(original)
	i.t.Errorf("%s", mutated)
}
