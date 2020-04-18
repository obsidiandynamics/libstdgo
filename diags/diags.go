// Package diags contains helpers to assist in debugging and diagnostics.
package diags

import "runtime"

// DumpAllStacks produces a string dump of stack traces for all running goroutines.
func DumpAllStacks() string {
	bytes := make([]byte, 1<<20)
	len := runtime.Stack(bytes, true)
	return string(bytes[:len])
}
