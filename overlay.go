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
	mingg := []Getter{}
	for _, g := range gg {
		// ignore nils
		if g == nil {
			continue
		}
		// consolidate overlays
		if og, ok := g.(*overlay); ok {
			mingg = append(mingg, og.gg...)
			continue
		}
		mingg = append(mingg, g)
	}
	return &overlay{mingg}
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
func (o *overlay) NewWatcher(done <-chan struct{}) GetterWatcher {
	ww := []GetterWatcher{}
	for _, g := range o.gg {
		if wg, ok := g.(WatchableGetter); ok {
			w := wg.NewWatcher(done)
			if w != nil {
				ww = append(ww, w)
			}
		}
	}
	if len(ww) == 0 {
		return nil
	}
	if len(ww) == 1 {
		return ww[0]
	}
	s := &stackWatcher{
		done: done,
		gw:   newGetterWatcher()}
	for _, w := range ww {
		s.append(w)
	}
	return s.gw
}
