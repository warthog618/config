// Copyright © 2018 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package config

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
	// The Watch should return an error other than context.Canceled if the Watch
	// has failed due to some underlying error condition, but could recover if
	// the underlying error condition is cleared.
	Watch(context.Context) error
}

// Decoder unmarshals configuration from raw []byte into the provided type,
// typically a map[string]interface{}.
type Decoder interface {
	Decode(b []byte, v interface{}) error
}

// Source represents a two stage Getter.  The first stage is the Loader which
// retrieves the configuration as a []byte from an underlying source. The second
// stage is the Decoder, which converts the returned []byte into a
// map[string]interface{}.
type Source struct {
	l      Loader
	d      Decoder
	m      atomic.Value // map[string]interface{}
	sep    string
	update map[string]interface{}
}

// NewSource creates a new Source using the provided loader and decoder.
// The configuration is loaded and decoded during construction, else an error is
// returned.
func NewSource(l Loader, d Decoder, options ...SourceOption) (*Source, error) {
	s := Source{l: l, d: d, sep: "."}
	for _, option := range options {
		option.applySourceOption(&s)
	}
	m, err := s.load()
	if err != nil {
		return nil, err
	}
	if m == nil {
		m = make(map[string]interface{})
	}
	s.m.Store(m)
	return &s, nil
}

func (s *Source) load() (map[string]interface{}, error) {
	b, err := s.l.Load()
	if err != nil {
		return nil, err
	}
	m := make(map[string]interface{})
	err = s.d.Decode(b, &m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

// Close releases any resources allocated by the source.
// This implicitly closes any Watch goroutines when the fsnotify they are
// monitoring is closed.
// After closing, the source is still readable via Get, but will no longer be
// updated.
func (s *Source) Close() (err error) {
	if c, ok := s.l.(io.Closer); ok {
		err = c.Close()
	}
	return err
}

// Get implements the Getter API.
func (s *Source) Get(key string) (interface{}, bool) {
	m := s.m.Load().(map[string]interface{})
	v, ok := tree.Get(m, key, s.sep)
	return v, ok
}

// Watch initiates a goroutine that monitors the source and updates it if the
// underlying source changes.
// The Watch goroutine may be cancelled by providing a ctx WithCancel and
// calling its cancel function.
// It is assumed that Watch and CommitUpdate will only be called from a single
// goroutine, and with CommitUpdate only called after a successful return from
// Watch.
func (s *Source) Watch(ctx context.Context) error {
	w, ok := s.l.(WatchedLoader)
	if !ok {
		return ErrUnwatchable
	}
	for {
		if err := w.Watch(ctx); err != nil {
			return err
		}
		m, _ := s.load()
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
func (s *Source) CommitUpdate() {
	s.m.Store(s.update)
	s.update = nil
}
