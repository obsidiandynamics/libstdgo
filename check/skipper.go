package check

import (
	"os"
	"strings"

	"github.com/obsidiandynamics/stdlibgo/commander"
)

// TestSkipper is the API contract for testing.T.Skip().
type TestSkipper interface {
	Skip(args ...interface{})
}

// Runnable in any no-arg function.
//
// Typically, this is a 'go doc' example function, of the form —
//  func Example_suffix() { ... }
//  func ExampleF_suffix() { ... }
//  func ExampleT_suffix() { ... }
//  func ExampleT_M_suffix() { ... }
type Runnable func()

// RunTargetted runs a given function if it the test case was specified using a narrow-matching regular expression;
// for example, by running 'debug.test -test.run ^...$' or 'go test -run ^...$'. If tests were run without a regex, or with
// a regex designed to match several tests, the example will be skipped.
//
// The purpose of RunTargetted is to enable the selective running of 'go doc' example snippets. These should not normally be
// run as part of package tests, CI builds and so on, but may occasionally be run from the IDE.
func RunTargetted(t TestSkipper, r Runnable) {
	runTargetted(t, r, os.Args)
}

func runTargetted(t TestSkipper, r Runnable, cmdArgs []string) {
	if isTargetted(cmdArgs) {
		r()
	} else {
		t.Skip("Skipped")
	}
}

// EnvGolabels is the name of the environment variable used by RequireLabel.
const EnvGolabels = "GOLABELS"

// RequireLabel ensures that the given label is present in the value of the GOLABELS environment variable, where the
// latter is a comma-separated list of arbitrary labels, e.g. GOLABELS=prod,test. If the required label is absent, the
// test will be skipped.
func RequireLabel(t TestSkipper, required string) {
	requireLabel(t, required, os.Args, os.Getenv)
}

type getenv = func(key string) string

func requireLabel(t TestSkipper, required string, cmdArgs []string, getenv getenv) {
	if !isTargetted(cmdArgs) && !hasLabel(getenv(EnvGolabels), required) {
		t.Skip("Skipped")
	}
}

func contains(strings []string, contains string) bool {
	for _, str := range strings {
		if str == contains {
			return true
		}
	}
	return false
}

func isTargetted(cmdArgs []string) bool {
	parsed := commander.Parse(cmdArgs).Mappify()
	runArg, _ := parsed.Value("run", "")
	testRunArg, _ := parsed.Value("test.run", "")

	return hasNarrowMatch(runArg, testRunArg)
}

func hasLabel(labels string, requiredTag string) bool {
	return contains(strings.Split(labels, ","), requiredTag)
}

func hasNarrowMatch(cmdArgs ...string) bool {
	for _, cmdArg := range cmdArgs {
		if isNarrowMatch(cmdArg) {
			return true
		}
	}
	return false
}

func isNarrowMatch(cmdArg string) bool {
	length := len(cmdArg)
	if length == 0 {
		return false
	}

	first := cmdArg[0]
	last := cmdArg[length-1]
	if first == '^' && last == '$' {
		return !strings.ContainsAny(cmdArg, "|*")
	} else {
		return false
	}
}
