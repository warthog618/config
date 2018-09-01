// Copyright © 2018 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package config

import (
	"context"
	"io"
	"strings"

	"github.com/warthog618/config/keys"
)

// Getter specifies the minimal interface for a configuration Getter.
//
// A Getter must be safe for concurrent use by multiple goroutines.
type Getter interface {
	// Get the value of the named config leaf key.
	// Also returns an ok, similar to a map read, to indicate if the value was
	// found.
	// The type underlying the returned interface{} must be convertable to the
	// expected type by cfgconv.
	//
	// Get does not need to support getting of objects, as returning of complete
	// objects is neither supported nor required.
	//
	// But it does support getting of arrays.
	// For arrays, referenced by the array name say "ax", a []interface{} must
	// be returned.
	// Array elements are referenced using keys of form "ax[idx]", where idx is
	// the zero-based index into the array.
	// The length of the array is returned by a key of form "ax[]".
	// If the getter only contains part of the array then it should return only
	// the elements it contains, not "ax" or "ax[]".
	//
	// For arrays of objects the array must be returned, to be consistent with
	// other arrays, but the elements may be nil.
	//
	// Must be safe to call from multiple goroutines.
	Get(key string) (value interface{}, found bool)
}

// WatchableGetter is the interface supported by Getters that may support being
// watched.
type WatchableGetter interface {
	// Watcher returns the watcher for a Getter.
	// Returns false if the associated Getter does not support watches.
	Watcher() (GetterWatcher, bool)
}

// GetterWatcher watches a getter for updates.
type GetterWatcher interface {
	// Close releases any resources allocated to the watcher, and cancels any
	// active watches.
	io.Closer
	// Watch blocks until the source has changed, or an error is detected.
	Watch(context.Context) error
	// CommitUpdate commits a change detected by Watch so that it becomes
	// visible to Get.
	CommitUpdate()
}

// Decorator is a func that takes one Getter and returns
// a decorated Getter.
type Decorator func(Getter) Getter

// getterDecorator is a common implementation for decorators.
type getterDecorator struct {
	g Getter
}

// Watcher implements the WatchableGetter interface
func (g getterDecorator) Watcher() (GetterWatcher, bool) {
	if wg, ok := g.g.(WatchableGetter); ok {
		return wg.Watcher()
	}
	return nil, false
}

// Decorate applies an ordered list of decorators to a Getter.
// The decorators are applied in reverse order, to create a decorator chain with
// the first decorator being the first link in the chain.
// When the returned getter is used, the first decorator is called first, then
// the second, etc and finally the decorated Getter itself.
func Decorate(g Getter, dd ...Decorator) Getter {
	dg := g
	for i := len(dd) - 1; i >= 0; i-- {
		dg = dd[i](dg)
	}
	return dg
}

// WithDefault provides a Decorator that falls back to a default Getter if the
// key is not found in the decorated Getter.
func WithDefault(d Getter) Decorator {
	return func(g Getter) Getter {
		if d == nil {
			return g
		}
		return Overlay(g, d)
	}
}

// WithGraft returns a decorator that attaches the root of the decorated Getter
// to a node in the config space.
// The prefix defines where the root node of the getter is located in the config
// space. The prefix must include any separator prior to the first field.
//
// e.g. with a prefix "a.module.", reading the key "a.module.field" from the
// WithGraft will return the "field" from the wrapped Getter.
func WithGraft(prefix string) Decorator {
	return func(g Getter) Getter {
		return graftDecorator{getterDecorator{g}, prefix}
	}
}

type graftDecorator struct {
	getterDecorator
	prefix string
}

func (g graftDecorator) Get(key string) (interface{}, bool) {
	if !strings.HasPrefix(key, g.prefix) {
		return nil, false
	}
	key = key[len(g.prefix):]
	return g.g.Get(key)
}

// WithKeyReplacer provides a decorator which performs a transformation on the
// key using the ReplacerFunc before calling the Getter.
func WithKeyReplacer(r keys.Replacer) Decorator {
	return func(g Getter) Getter {
		if r == nil {
			return g
		}
		return keyReplacerDecorator{getterDecorator{g}, r}
	}
}

type keyReplacerDecorator struct {
	getterDecorator
	r keys.Replacer
}

func (g keyReplacerDecorator) Get(key string) (interface{}, bool) {
	return g.g.Get(g.r.Replace(key))
}

// WithMustGet provides a Decorator that panics if a key is not found by the
// decorated Getter.
func WithMustGet() Decorator {
	return func(g Getter) Getter {
		return mustDecorator{getterDecorator{g}}
	}
}

type mustDecorator struct {
	getterDecorator
}

func (g mustDecorator) Get(key string) (interface{}, bool) {
	v, found := g.g.Get(key)
	if !found {
		panic(NotFoundError{Key: key})
	}
	return v, true
}

// WithPrefix provides a Decorator that adds a prefix to the key before calling
// the Getter.
// This is a common special case of KeyReplacer where the key is prefixed with a
// fixed string.
func WithPrefix(prefix string) Decorator {
	return func(g Getter) Getter {
		return prefixDecorator{getterDecorator{g}, prefix}
	}
}

type prefixDecorator struct {
	getterDecorator
	prefix string
}

func (g prefixDecorator) Get(key string) (interface{}, bool) {
	return g.g.Get(g.prefix + key)
}
