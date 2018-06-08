// Copyright © 2017 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package getter

import (
	"strings"
)

// Getter provides the minimal interface for a configuration Getter.
type Getter interface {
	// Get the value of the named config leaf key.
	// Also returns an ok, similar to a map read, to indicate if the value
	// was found.
	// The type underlying the returned interface{} must be convertable to
	// the expected type by cfgconv.
	// Get is not expected to be performed on node keys, but in case it is
	// the Get should return a nil interface{} and false, even if the node
	// exists in the config tree.
	// Must be safe to call from multiple goroutines.
	Get(key string) (value interface{}, found bool)
}

// Func is a func that implements a Getter Get.
type Func func(key string) (interface{}, bool)

// Get calls the GetterFunc to perform the get.
func (g Func) Get(key string) (interface{}, bool) {
	return g(key)
}

// Decorator is a func that takes one Getter and returns
// a decorated Getter.
type Decorator func(Getter) Getter

// Decorate applies an ordered list of decorators to a Getter.
// The first decorator in the list is the one that decorates the
// provided Getter, the second decorates the first, etc.
func Decorate(g Getter, dd ...Decorator) Getter {
	dg := g
	for _, d := range dd {
		dg = d(dg)
	}
	return dg
}

// Replacer maps a key from one space to another.
type Replacer interface {
	Replace(key string) string
}

// Mapped returns a decorator that performs key mapping from config space
// prior to getting from the wrapped Getter.
func Mapped(r Replacer) Decorator {
	return func(g Getter) Getter {
		return Func(func(key string) (interface{}, bool) {
			return g.Get(r.Replace(key))
		})
	}
}

// Prefixed returns a decorator that moves the root of the decorated
// Getter down into the config space.
// The prefix defines where the root node of the getter is located in the config space.
// The prefix must include any separator prior to the first field.
//
// e.g. with a prefix "a.module.", reading the key "a.module.field" from the
// Prefixed will return the "field" from the wrapped Getter.
func Prefixed(prefix string) Decorator {
	return func(g Getter) Getter {
		return Func(func(key string) (interface{}, bool) {
			if !strings.HasPrefix(key, prefix) {
				return nil, false
			}
			key = key[len(prefix):]
			return g.Get(key)
		})
	}
}
