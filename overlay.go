// Copyright © 2018 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package config

// Overlay attempts a get using a number of Getters, in the order provided,
// returning the first result found.
// This can be considered an immutable form of Stack.
func Overlay(gg ...Getter) Getter {
	if len(gg) == 1 {
		return gg[0]
	}
	return &overlay{gg}
}

type overlay struct {
	gg []Getter
}

// Get gets the raw value corresponding to the key.
// It iterates through the list of getters, searching for a matching key.
// Returns the first match found, or an error if none is found.
func (o *overlay) Get(key string) (interface{}, bool) {
	for _, g := range o.gg {
		if v, ok := g.Get(key); ok {
			return v, ok
		}
	}
	return nil, false
}

// Watcher implements the WatchableGetter interface.
func (o *overlay) Watcher() (GetterWatcher, bool) {
	ww := []GetterWatcher{}
	for _, g := range o.gg {
		if wg, ok := g.(WatchableGetter); ok {
			if w, ok := wg.Watcher(); ok {
				ww = append(ww, w)
			}
		}
	}
	if len(ww) != 0 {
		w := &stackWatcher{
			mu:    nullLocker{},
			gg:    ww,
			uchan: make(chan GetterWatcher),
			cchan: make(chan GetterWatcher, 1)}
		return w, true
	}
	return nil, false
}

// nullLocker implements a stubbed out locker for the stackWatcher for the
// Overlay which is immutable and so does not require mutex locks.
type nullLocker struct{}

func (n nullLocker) Lock() {}

func (n nullLocker) Unlock() {}
