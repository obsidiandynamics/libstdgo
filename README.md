<img src="https://raw.githubusercontent.com/wiki/obsidiandynamics/libstdgo/images/libstdgo-logo.png" width="90px" alt="logo"/> `libstdgo`
===
![Go version](https://img.shields.io/github/go-mod/go-version/obsidiandynamics/libstdgo)
[![Build](https://travis-ci.org/obsidiandynamics/libstdgo.svg?branch=master) ](https://travis-ci.org/obsidiandynamics/libstdgo#)
![Release](https://img.shields.io/github/v/release/obsidiandynamics/libstdgo?color=ff69b4)
[![Codecov](https://codecov.io/gh/obsidiandynamics/libstdgo/branch/master/graph/badge.svg)](https://codecov.io/gh/obsidiandynamics/libstdgo)
[![Go Report Card](https://goreportcard.com/badge/github.com/obsidiandynamics/libstdgo)](https://goreportcard.com/report/github.com/obsidiandynamics/libstdgo)
[![Total alerts](https://img.shields.io/lgtm/alerts/g/obsidiandynamics/libstdgo.svg?logo=lgtm&logoWidth=18)](https://lgtm.com/projects/g/obsidiandynamics/libstdgo/alerts/)
[![GoDoc Reference](https://img.shields.io/badge/docs-GoDoc-blue.svg)](https://pkg.go.dev/github.com/obsidiandynamics/libstdgo?tab=subdirectories)

**Standard libraries for Go**, taking care of things like:
* `concurrent`: **concurrent and thread-safe data structures** —
  - `AtomicCounter`: atomic `int64` counter
  - `Scoreboard`: a space-efficient map of `string`-keyed `int64` counters
  - `AtomicReference` an atomic reference that allows for `nil` pointers
  - `Deadline` - conditional running of tasks within a deadline
* `scribe`: **logging façade that features logger mocking and assertions**, and comes with **ready-to-go bindings** for —
  - The built-in `os.stdout` function
  - The built-in `log.Printf` logger
  - [Glog](https://github.com/golang/glog)
  - [Log15](https://github.com/inconshreveable/log15)
  - [Logrus](https://github.com/sirupsen/logrus)
  - [Seelog](https://github.com/cihub/seelog)
  - [Zap](https://github.com/uber-go/zap)
  - Overlog — a thread-safe logger for debugging concurrent apps, built into Scribe
* `check`: **assertion utilities**
  - `ThatPanicsAsExpected(func)`: asserting panic expectations
  - `Wait(t, timeout).UntilAsserted(assertion)`: time-based assertions
  - `TestCapture`: capture of `testing.T` failures (for self-testing of assertion libraries)
  - `Intercept(t).Mutate(...)`: enrichment of assertion failure messages
* `commander`: **schema-less command-line argument parsing**
  - `Parse(os.Args).Mappify`
* `fault`: **fault injection**
* `arity`: **extraction of optional arguments to variadic functions**
  - `arg := arity.SoleUntyped("a_default", args).(string)`
* `diags`: **debugging and diagnostics**


Check out the [GoDocs](https://pkg.go.dev/github.com/obsidiandynamics/libstdgo?tab=subdirectories).
