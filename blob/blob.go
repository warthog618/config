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

// Blob represents a two stage Getter.  The first stage is the Loader which
// retrieves the configuration as a []byte blob from an underlying source. The
// second stage is the Decoder, which converts the returned []byte blob into a
// map[string]interface{}.
type Blob struct {
	m   map[string]interface{}
	sep string
}

// New creates a new Blob using the provided loader and decoder.
// The configuration is loaded and decoded during construction, else an error is
// returned.
func New(l Loader, d Decoder, options ...Option) (*Blob, error) {
	s := Blob{sep: "."}
	for _, option := range options {
		option.applyBlobOption(&s)
	}
	m, err := load(l, d)
	if err != nil {
		return nil, err
	}
	s.m = m
	return &s, nil
}

// Get implements the Getter API.
func (s *Blob) Get(key string) (interface{}, bool) {
	v, ok := tree.Get(s.m, key, s.sep)
	return v, ok
}

// WatchedBlob represents a Blob that can be watched for changes.
type WatchedBlob struct {
	l      WatchedLoader
	d      Decoder
	m      atomic.Value // map[string]interface{}
	sep    string
	update map[string]interface{}
}

// NewWatched creates a new Blob using the provided loader and decoder.
// The configuration is loaded and decoded during construction, else an error is
// returned.
func NewWatched(l WatchedLoader, d Decoder, options ...WatchedBlobOption) (*WatchedBlob, error) {
	s := WatchedBlob{l: l, d: d, sep: "."}
	for _, option := range options {
		option.applyWatchedBlobOption(&s)
	}
	m, err := load(l, d)
	if err != nil {
		return nil, err
	}
	if m == nil {
		m = make(map[string]interface{})
	}
	s.m.Store(m)
	return &s, nil
}

// Close releases any resources allocated by the blob.
// This implicitly closes any active Watches.
// After closing, the blob is still readable via Get, but will no longer be
// updated.
func (s *WatchedBlob) Close() (err error) {
	return s.l.Close()
}

// Get implements the Getter API.
func (s *WatchedBlob) Get(key string) (interface{}, bool) {
	m := s.m.Load().(map[string]interface{})
	v, ok := tree.Get(m, key, s.sep)
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
func (s *WatchedBlob) Watch(ctx context.Context) error {
	for {
		if err := s.l.Watch(ctx); err != nil {
			return err
		}
		m, err := load(s.l, s.d)
		if err != nil {
			return WithTemporary(err)
		}
		if m == nil {
			continue
		}
		oldm := s.m.Load().(map[string]interface{})
		if reflect.DeepEqual(m, oldm) {
			continue
		}
		s.update = m
		return nil
	}
}

// CommitUpdate commits a change to the configuration detected by Watch, making
// the change visible to Get.
// It is assumed that Watch and CommitUpdate will only be called from a single
// goroutine, and with CommitUpdate only called after a successful return from
// Watch.
func (s *WatchedBlob) CommitUpdate() {
	s.m.Store(s.update)
	s.update = nil
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
