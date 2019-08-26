// Copyright © 2018 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package blob provides a getter that loads and decodes configuration from a
// source where the configuration is stored in a known format.
package blob

import (
	"os"
	"reflect"
	"sync/atomic"

	"github.com/warthog618/config"
	"github.com/warthog618/config/blob/loader/file"
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

// ErrorHandler handles an error.
type ErrorHandler func(error)

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
	// handler for construction load errors
	ceh ErrorHandler
}

// New creates a new Blob using the provided loader and decoder.
// The configuration is loaded and decoded during construction, else an error is
// returned.
func New(l Loader, d Decoder, options ...Option) *Getter {
	g := Getter{l: l, d: d, pathSep: "."}
	for _, option := range options {
		option.applyOption(&g)
	}
	msi, err := load(l, d) // initial load
	if err == nil {
		g.msi.Store(msi)
	} else {
		if g.ceh != nil {
			g.ceh(err)
		}
	}
	return &g
}

// Get implements the Getter API.
func (g *Getter) Get(key string) (interface{}, bool) {
	msi := g.msi.Load()
	if msi == nil {
		return nil, false
	}
	msi = msi.(map[string]interface{})
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

// NewConfigFile is a helper function that creates a File getter.
// The config file path is defined in either the existing config, in a field
// indicated by pathfield, or a default path.
// If the config file is specified in the config then it must exist and load, or
// the function will panic.
// If the config file is not specified in the config then the default config
// file is used, if it exists, or an empty Getter.
// Any provided foptions are passed to the File constructor.
func NewConfigFile(cfg *config.Config, pathfield string,
	defpath string, fdec Decoder, foptions ...file.Option) config.Getter {
	path, err := cfg.Get(pathfield)
	if err == nil {
		// explicitly specified config file - must be there
		cfgFile := file.New(path.String(), foptions...)
		jget := New(cfgFile, fdec, MustLoad())
		return jget
	}
	// implicit and optional default config file
	cfgFile := file.New(defpath)
	jget := New(cfgFile, fdec, WithErrorHandler(func(e error) {
		if _, ok := e.(*os.PathError); !ok {
			panic(e)
		}
	}))
	return jget
}
