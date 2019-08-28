// Copyright © 2018 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package config

import (
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

// GetterAsOption allows a Getter to be passed to New as an option.
type GetterAsOption struct {
}

func (s GetterAsOption) applyConfigOption(c *Config) {
}

// UpdateCommit is a function that commits an update to a getter.
// After the call the change becomes visible to Get.
type UpdateCommit func()

// WatchableGetter is the interface supported by Getters that may support being
// watched.
type WatchableGetter interface {
	// Create a watcher on the getter.
	// Watcher will exit if the done chan closes.
	// Watcher will send updates via the update channel.
	// Watcher will send terminal errors via the err channel.
	NewWatcher(done <-chan struct{}) GetterWatcher
}

// GetterWatcher contains channels returning updates and errors from Getter
// watchers.
type GetterWatcher interface {
	Update() <-chan GetterUpdate
}

// GetterUpdate contains an update from a getter.
type GetterUpdate interface {
	Commit()
}

type getterWatcher struct {
	uch chan GetterUpdate
}

func (g *getterWatcher) Update() <-chan GetterUpdate {
	return g.uch
}

func newGetterWatcher() *getterWatcher {
	return &getterWatcher{uch: make(chan GetterUpdate)}
}

// Decorator is a func that takes one Getter and returns
// a decorated Getter.
type Decorator func(Getter) Getter

// getterDecorator is a common implementation for decorators.
type getterDecorator struct {
	g Getter
}

// NewWatcher implements the WatchableGetter interface.
func (g getterDecorator) NewWatcher(done <-chan struct{}) GetterWatcher {
	if wg, ok := g.g.(WatchableGetter); ok {
		return wg.NewWatcher(done)
	}
	return nil
}

// Decorate applies an ordered list of decorators to a Getter.
// The decorators are applied in reverse order, to create a decorator chain with
// the first decorator being the first link in the chain.
// When the returned getter is used, the first decorator is called first, then
// the second, etc and finally the decorated Getter itself.
func Decorate(g Getter, dd ...Decorator) Getter {
	if g == nil {
		return nil
	}
	dg := g
	for i := len(dd) - 1; i >= 0; i-- {
		dg = dd[i](dg)
	}
	return dg
}

// WithFallback provides a Decorator that falls back to a default Getter if the
// key is not found in the decorated Getter.
func WithFallback(d Getter) Decorator {
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

// UpdateHandler receives an update, performs some transformation
// on it, and forwards (or not) the transformed update.
// Must return if either the done or in channels are closed.
// May close the out chan to indicate that no further updates are possible.
type UpdateHandler func(done <-chan struct{}, in <-chan GetterUpdate, out chan<- GetterUpdate)

// WithUpdateHandler adds an update processing decorator to a getter.
func WithUpdateHandler(handler UpdateHandler) Decorator {
	return func(g Getter) Getter {
		if _, ok := g.(WatchableGetter); !ok {
			return g
		}
		return updateDecorator{g: g, h: handler}
	}
}

type updateDecorator struct {
	g Getter
	h UpdateHandler
}

// NewWatcher implements the WatchableGetter interface.
func (g updateDecorator) NewWatcher(done <-chan struct{}) GetterWatcher {
	wg := g.g.(WatchableGetter)
	w := newGetterWatcher()
	gw := wg.NewWatcher(done)
	go g.h(done, gw.Update(), w.uch)
	return w
}

// Get implements the Watcher interface.
func (g updateDecorator) Get(key string) (interface{}, bool) {
	return g.g.Get(key)
}
