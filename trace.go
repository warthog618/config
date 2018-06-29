// Copyright © 2018 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package config

// WithTrace provides a decorator that calls the Getter, and then
// calls a TraceFunc with the result.
func WithTrace(t TraceFunc) Decorator {
	return func(g Getter) Getter {
		return GetterFunc(func(key string) (interface{}, bool) {
			v, ok := g.Get(key)
			t(key, v, ok)
			return v, ok
		})
	}
}

// TraceFunc traces the parameters and results of a call to a Getter.
type TraceFunc func(k string, v interface{}, ok bool)
