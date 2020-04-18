`libstdgo`
===
[![Build](https://travis-ci.org/obsidiandynamics/libstdgo.svg?branch=master) ](https://travis-ci.org/obsidiandynamics/libstdgo#)
[![Codecov](https://codecov.io/gh/obsidiandynamics/libstdgo/branch/master/graph/badge.svg)](https://codecov.io/gh/obsidiandynamics/libstdgo)
[![Language grade: Go](https://img.shields.io/lgtm/grade/go/g/obsidiandynamics/libstdgo.svg?logo=lgtm&logoWidth=18)](https://lgtm.com/projects/g/obsidiandynamics/libstdgo/context:go)
[![Total alerts](https://img.shields.io/lgtm/alerts/g/obsidiandynamics/libstdgo.svg?logo=lgtm&logoWidth=18)](https://lgtm.com/projects/g/obsidiandynamics/libstdgo/alerts/)

Standard libraries for Go, taking care of things like:

* Concurrent and thread-safe data structures —
  - `AtomicCounter`: atomic `int64` counter
  - `Scoreboard`: a space-efficient map of `string`-keyed `int64` counters
  - `AtomicReference` an atomic reference that allows for `nil` pointers
  — `Deadline` - conditional running of tasks within a deadline
* Logging façade that features logger mocking and assertions, and comes with ready-to-go bindings for —
  - The built-in `os.stdout` function
  - The built-in `log.Printf` logger
  - Glog
  - Log15
  - Logrus
  - Seelog
  - Zap
  - Overlog (a thread-safe logger for debugging concurrent apps)
* Assertion utilities
* Schema-less command-line argument parsing
* Fault injection
* Extraction of optional arguments to variadic functions
* Debugging and diagnostics


Check out the [GoDocs](https://pkg.go.dev/github.com/obsidiandynamics/libstdgo?tab=subdirectories).
