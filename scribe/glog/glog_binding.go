// Package glog provides a Glog binding for Scribe.
package glog

import (
	"github.com/golang/glog"
	"github.com/obsidiandynamics/libstdgo/scribe"
)

// Bind creates a binding for Glog.
func Bind() scribe.LoggerFactories {
	return scribe.LoggerFactories{
		scribe.Trace: scribe.Fac(glog.Infof),
		scribe.Debug: scribe.Fac(glog.Infof),
		scribe.Info:  scribe.Fac(glog.Infof),
		scribe.Warn:  scribe.Fac(glog.Warningf),
		scribe.Error: scribe.Fac(glog.Errorf),
	}
}
