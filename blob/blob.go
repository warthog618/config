// Copyright © 2018 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package blob

import (
	"context"
	"io"
	"reflect"
	"sync/atomic"

	"github.com/warthog618/config/tree"
)

// Loader retrieves raw configuration data, as []byte, from some source.
// The Loader may also support the WatchedLoader interface if it is watchable.
type Loader interface {
	Load() ([]byte, error)
}

// WatchedLoader represents an API supported by Loaders that can watch the
// underlying source for configuration changes.
type WatchedLoader interface {
	Loader
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

// Getter represents a two stage Getter.  The first stage is the Loader which
// retrieves the configuration as a []byte blob from an underlying source. The
// second stage is the Decoder, which converts the returned []byte blob into a
// map[string]interface{}.
type Getter struct {
	// current committed configuration
	msi map[string]interface{}
	// separator between tiers
	pathSep string
}

// New creates a new Blob using the provided loader and decoder.
// The configuration is loaded and decoded during construction, else an error is
// returned.
func New(l Loader, d Decoder, options ...Option) (*Getter, error) {
	g := Getter{pathSep: "."}
	for _, option := range options {
		option.applyBlobOption(&g)
	}
	msi, err := load(l, d)
	if err != nil {
		return nil, err
	}
	g.msi = msi
	return &g, nil
}

// Get implements the Getter API.
func (g *Getter) Get(key string) (interface{}, bool) {
	v, ok := tree.Get(g.msi, key, g.pathSep)
	return v, ok
}

// WatchedGetter represents a Blob that can be watched for changes.
type WatchedGetter struct {
	l WatchedLoader
	d Decoder
	// current committed configuration
	msi atomic.Value // map[string]interface{}
	// separator between tiers
	pathSep string
	// lastest uncommitted configuration
	update map[string]interface{}
}

// NewWatched creates a new Blob using the provided loader and decoder.
// The configuration is loaded and decoded during construction, else an error is
// returned.
func NewWatched(l WatchedLoader, d Decoder, options ...WatchedBlobOption) (*WatchedGetter, error) {
	g := WatchedGetter{l: l, d: d, pathSep: "."}
	for _, option := range options {
		option.applyWatchedBlobOption(&g)
	}
	m, err := load(l, d)
	if err != nil {
		return nil, err
	}
	if m == nil {
		m = make(map[string]interface{})
	}
	g.msi.Store(m)
	return &g, nil
}

// Close releases any resources allocated by the blob.
// This implicitly closes any active Watches.
// After closing, the blob is still readable via Get, but will no longer be
// updated.
func (g *WatchedGetter) Close() (err error) {
	return g.l.Close()
}

// Get implements the Getter API.
func (g *WatchedGetter) Get(key string) (interface{}, bool) {
	msi := g.msi.Load().(map[string]interface{})
	v, ok := tree.Get(msi, key, g.pathSep)
	return v, ok
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
func (g *WatchedGetter) Watch(ctx context.Context) error {
	for {
		if err := g.l.Watch(ctx); err != nil {
			return err
		}
		updatedmsi, err := load(g.l, g.d)
		if err != nil {
			return WithTemporary(err)
		}
		if updatedmsi == nil {
			continue
		}
		msi := g.msi.Load().(map[string]interface{})
		if reflect.DeepEqual(updatedmsi, msi) {
			continue
		}
		g.update = updatedmsi
		return nil
	}
}

// CommitUpdate commits a change to the configuration detected by Watch, making
// the change visible to Get.
// It is assumed that Watch and CommitUpdate will only be called from a single
// goroutine, and with CommitUpdate only called after a successful return from
// Watch.
func (g *WatchedGetter) CommitUpdate() {
	g.msi.Store(g.update)
	g.update = nil
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
