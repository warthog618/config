// Copyright © 2018 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package blob provides a getter that loads and decodes configuration from a
// source where the configuration is stored in a known format.
package blob

import (
	"reflect"
	"sync/atomic"

	"github.com/warthog618/config"
	"github.com/warthog618/config/tree"
)

// Loader retrieves raw configuration data, as []byte, from some source.
// The Loader may also support the watchableLoader interface if it is watchable.
type Loader interface {
	Load() ([]byte, error)
}

// WatchableLoader is the interface supported by Loaders that can be watched for
// changes.
type WatchableLoader interface {
	NewWatcher(done <-chan struct{}) <-chan error
}

// Decoder unmarshals configuration from raw []byte into the provided type,
// typically a map[string]interface{}.
type Decoder interface {
	Decode(b []byte, v interface{}) error
}

// Getter represents a two stage getter for blobs.  The first stage is the
// Loader which retrieves the configuration as a []byte blob from an underlying
// source. The second stage is the Decoder, which converts the returned []byte
// blob into a map[string]interface{}.
type Getter struct {
	l Loader
	d Decoder
	// current committed configuration
	msi atomic.Value // map[string]interface{}
	// separator between tiers
	pathSep string
}

// New creates a new Blob using the provided loader and decoder.
// The configuration is loaded and decoded during construction, else an error is
// returned.
func New(l Loader, d Decoder, options ...Option) (*Getter, error) {
	g := Getter{l: l, d: d, pathSep: "."}
	for _, option := range options {
		option.applyOption(&g)
	}
	msi, err := load(l, d) // initial load
	if err != nil {
		return nil, err
	}
	g.msi.Store(msi)
	return &g, nil
}

// Get implements the Getter API.
func (g *Getter) Get(key string) (interface{}, bool) {
	msi := g.msi.Load().(map[string]interface{})
	v, ok := tree.Get(msi, key, g.pathSep)
	return v, ok
}

// NewWatcher creates a watcher for the getter.
// Returns nil if the getter does not support being watched.
func (g *Getter) NewWatcher(done <-chan struct{}) config.GetterWatcher {
	if wl, ok := g.l.(WatchableLoader); ok {
		w := wl.NewWatcher(done)
		if w == nil {
			return nil
		}
		gw := &getterWatcher{uch: make(chan config.GetterUpdate)}
		go g.watch(done, w, gw)
		return gw
	}
	return nil
}

func (g *Getter) watch(done <-chan struct{}, update <-chan error, gw *getterWatcher) {
	defer close(gw.uch)
	send := func(u getterUpdate) {
		select {
		case gw.uch <- u:
		case <-done:
			return
		}
	}
	for {
		select {
		case <-done:
			return
		case err, ok := <-update:
			if !ok {
				return
			}
			if err != nil {
				send(getterUpdate{g: g, err: err})
				continue
			}
			msi, err := load(g.l, g.d)
			if err != nil {
				send(getterUpdate{g: g, err: err, temperr: true})
				continue
			}
			if msi == nil {
				continue
			}
			oldmsi := g.msi.Load().(map[string]interface{})
			if reflect.DeepEqual(msi, oldmsi) {
				continue
			}
			send(getterUpdate{g: g, commit: func() { g.msi.Store(msi) }})
		}
	}
}

func load(l Loader, d Decoder) (map[string]interface{}, error) {
	b, err := l.Load()
	if err != nil {
		return nil, err
	}
	m := make(map[string]interface{})
	err = d.Decode(b, &m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

type getterWatcher struct {
	uch chan config.GetterUpdate
}

func (g *getterWatcher) Update() <-chan config.GetterUpdate {
	return g.uch
}

type getterUpdate struct {
	g       config.Getter
	err     error
	commit  func()
	temperr bool
}

func (g getterUpdate) Getter() config.Getter {
	return g.g
}

func (g getterUpdate) Err() error {
	return g.err
}

func (g getterUpdate) TemporaryError() bool {
	return g.temperr
}

func (g getterUpdate) Commit() {
	if g.commit == nil {
		return
	}
	g.commit()
}
