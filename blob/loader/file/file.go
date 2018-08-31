// Copyright © 2018 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package file provides a loader from file for config.
package file

import (
	"context"
	"io/ioutil"

	"github.com/fsnotify/fsnotify"
	"github.com/warthog618/config/blob"
)

// Loader provides reads configuration from the local filesystem.
type Loader struct {
	filename string
	w        *watcher
}

// New creates a loader with the specified path.
func New(filename string, options ...Option) (*Loader, error) {
	l := Loader{filename: filename}
	for _, option := range options {
		err := option.applyOption(&l)
		if err != nil {
			return nil, err
		}
	}
	return &l, nil
}

// Load returns the current content of the file.
// The file is expected to exist and be readable, else an error is returned.
func (l *Loader) Load() ([]byte, error) {
	return ioutil.ReadFile(l.filename)
}

// Watcher returns the watcher for the loader.
// The watcher must be created using the WithWatch construction option.
func (l *Loader) Watcher() blob.WatcherCloser {
	return l.w
}

// watcher watches a file for changes.
type watcher struct {
	*fsnotify.Watcher
}

// Watch blocks until the watched file is altered.
// Alteration is relative to the construction of the Watcher, or the previous
// call to Watch, whichever is more recent.
// The Watch may be cancelled by providing a context WithCancel and
// calling the cancel function.
// The returned error is nil if file is changed, or an error if the context has
// been cancelled or the WatchedFile closed.
func (w *watcher) Watch(ctx context.Context) error {
	for {
		select {
		case _, ok := <-w.Events:
			if !ok {
				return context.Canceled
			}
			return nil
		case e, ok := <-w.Errors:
			if !ok {
				return context.Canceled
			}
			return e
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
