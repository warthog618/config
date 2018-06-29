// Copyright © 2018 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package config

import "strings"

// Getter specifies the minimal interface for a configuration Getter.
//
// A Getter must be safe for concurrent use by multiple goroutines.
type Getter interface {
	// Get the value of the named config leaf key.
	// Also returns an ok, similar to a map read, to indicate if the value
	// was found.
	// The type underlying the returned interface{} must be convertable to
	// the expected type by cfgconv.
	// For arrays a []interface{} should be returned.
	// For objects a map[string]interface{} should be returned.
	//
	// Must be safe to call from multiple goroutines.
	Get(key string) (value interface{}, found bool)
}

// GetterFunc is a func that implements a Getter Get.
type GetterFunc func(key string) (interface{}, bool)

// Get calls the GetterFunc to perform the get, satisfying the Getter interface.
func (g GetterFunc) Get(key string) (interface{}, bool) {
	return g(key)
}

// Decorator is a func that takes one Getter and returns
// a decorated Getter.
type Decorator func(Getter) Getter

// Decorate applies an ordered list of decorators to a Getter.
// The decorators are applied in reverse order, to create a decorator chain
// with the first decorator being the first link in the chain.
// When the returned getter is used, the first decorator is called first,
// then the second, etc and finally the decorated Getter itself.
func Decorate(g Getter, dd ...Decorator) Getter {
	dg := g
	for i := len(dd) - 1; i >= 0; i-- {
		dg = dd[i](dg)
	}
	return dg
}

// Overlay attempts a get using a number of Getters, in the order provided,
// returning the first result found.
// This can be considered an immutable form of Stack.
func Overlay(gg ...Getter) Getter {
	return GetterFunc(func(key string) (interface{}, bool) {
		for _, g := range gg {
			if v, ok := g.Get(key); ok {
				return v, ok
			}
		}
		return nil, false
	})
}

// Replacer replaces one string with another.
type Replacer interface {
	Replace(string) string
}

// ReplacerFunc replaces one string with another.
type ReplacerFunc func(string) string

// WithDefault provides a Decorator that falls back to a default Getter
// if the key is not found in the decorated Getter.
func WithDefault(d Getter) Decorator {
	return func(g Getter) Getter {
		if d == nil {
			return g
		}
		return Overlay(g, d)
	}
}

// WithGraft returns a decorator that attaches the root of the decorated
// Getter to a node in the config space.
// The prefix defines where the root node of the getter is located in the config space.
// The prefix must include any separator prior to the first field.
//
// e.g. with a prefix "a.module.", reading the key "a.module.field" from the
// WithGraft will return the "field" from the wrapped Getter.
func WithGraft(prefix string) Decorator {
	return func(g Getter) Getter {
		return GetterFunc(func(key string) (interface{}, bool) {
			if !strings.HasPrefix(key, prefix) {
				return nil, false
			}
			key = key[len(prefix):]
			return g.Get(key)
		})
	}
}

// WithKeyReplacer provides a decorator which performs a transformation
// on the key using the ReplacerFunc before calling the Getter.
func WithKeyReplacer(r Replacer) Decorator {
	return func(g Getter) Getter {
		if r == nil {
			return g
		}
		return GetterFunc(func(key string) (interface{}, bool) {
			return g.Get(r.Replace(key))
		})
	}
}

// WithPrefix provides a Decorator that adds a prefix to the key before
// calling the Getter.
// This is a common special case of KeyReplacer where the key is
// prefixed with a fixed string.
func WithPrefix(prefix string) Decorator {
	return func(g Getter) Getter {
		return GetterFunc(func(key string) (interface{}, bool) {
			return g.Get(prefix + key)
		})
	}
}
