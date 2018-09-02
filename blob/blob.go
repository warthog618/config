// Copyright © 2018 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package blob provides a getter that loads and decodes configuration from a
// source where the configuration is stored in a known format.
package blob

import (
	"context"
	"io"
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
	Watcher() (WatcherCloser, bool)
}

// WatcherCloser represents an API supported by Loaders that can watch the
// underlying source for configuration changes.
type WatcherCloser interface {
	// Close releases any resources allocated to the watcher, and cancels any
	// active watches.
	io.Closer
	// Watch blocks until the underlying source has changed since construction
	// or the previous Watch call.
	// The Watch should return context.Canceled if it has been terminated for
	// any reason, including the context being done or the underlying source
	// closing.
	// The Watch should return an error supporting the temporary interface if
	// the Watch has failed due to some underlying error condition, but could
	// recover if the underlying error condition is cleared.
	Watch(context.Context) error
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
	// watcher on the loader
	w *watcher
}

// New creates a new Blob using the provided loader and decoder.
// The configuration is loaded and decoded during construction, else an error is
// returned.
func New(l Loader, d Decoder, options ...Option) (*Getter, error) {
	g := Getter{l: l, d: d, pathSep: "."}
	for _, option := range options {
		option.applyOption(&g)
	}
	if wl, ok := l.(WatchableLoader); ok {
		if w, ok := wl.Watcher(); ok {
			// must be created before the initial load
			g.w = &watcher{g: &g, w: w}
		}
	}
	msi, err := load(l, d) // initial load
	if err != nil {
		if g.w != nil {
			g.w.Close()
		}
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

// Watcher returns the watcher for the getter.
// The watcher is created at construction time if the Loader has a watcher.
// Returns false if the getter is not watchable.
func (g *Getter) Watcher() (config.GetterWatcher, bool) {
	if g.w == nil {
		return nil, false
	}
	return g.w, true
}

// watcher watches a Getter for changes.
type watcher struct {
	g *Getter
	w WatcherCloser
	// lastest uncommitted configuration
	update map[string]interface{}
}

// Close releases any resources allocated by the watcher.
// This implicitly closes any active watches.
// After closing, the getter is still readable via Get, but will no longer be
// updated by the watcher.
func (w *watcher) Close() (err error) {
	return w.w.Close()
}

// Watch blocks until the underlying source changes.
// The Watch returns nil if the source has changed and no error was encountered.
// Otherwise the Watch returns the error encountered.
// If the error is temporary it will support the temporary interface.
// The Watch may be cancelled by providing a ctx WithCancel and
// calling its cancel function.
// It is assumed that Watch and CommitUpdate will only be called from a single
// goroutine, and with CommitUpdate only called after a successful return from
// Watch.
func (w *watcher) Watch(ctx context.Context) error {
	for {
		if err := w.w.Watch(ctx); err != nil {
			return err
		}
		updatedmsi, err := load(w.g.l, w.g.d)
		if err != nil {
			return WithTemporary(err)
		}
		if updatedmsi == nil {
			continue
		}
		msi := w.g.msi.Load().(map[string]interface{})
		if reflect.DeepEqual(updatedmsi, msi) {
			continue
		}
		w.update = updatedmsi
		return nil
	}
}

// CommitUpdate commits a change to the configuration detected by Watch, making
// the change visible to Get.
// It is assumed that Watch and CommitUpdate will only be called from a single
// goroutine, and with CommitUpdate only called after a successful return from
// Watch.
func (w *watcher) CommitUpdate() {
	w.g.msi.Store(w.update)
	w.update = nil
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
