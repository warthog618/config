// Copyright © 2018 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package config

// WithTrace provides a decorator that calls the Getter, and then
// calls a TraceFunc with the result.
func WithTrace(t TraceFunc) Decorator {
	return func(g Getter) Getter {
		return traceDecorator{getterDecorator{g}, t}
	}
}

// TraceFunc traces the parameters and results of a call to a Getter.
type TraceFunc func(k string, v interface{}, ok bool)

type traceDecorator struct {
	getterDecorator
	t TraceFunc
}

func (g traceDecorator) Get(key string) (interface{}, bool) {
	v, ok := g.g.Get(key)
	g.t(key, v, ok)
	return v, ok
}
