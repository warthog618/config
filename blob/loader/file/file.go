// Copyright © 2018 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package file

import (
	"context"
	"io/ioutil"

	"github.com/fsnotify/fsnotify"
)

// Loader provides a read-once source of configuration from the local filesystem.
type Loader struct {
	filename string
}

// New creates a File loader with the specified path.
func New(filename string) *Loader {
	return &Loader{filename: filename}
}

// Load returns the current content of the file.
// The file is expected to exist and be readable, else an error is returned.
func (l *Loader) Load() ([]byte, error) {
	return ioutil.ReadFile(l.filename)
}

// WatchedLoader provides an active source of configuration from the local filesystem.
type WatchedLoader struct {
	filename string
	watcher  *fsnotify.Watcher
}

// NewWatched creates a WatchedFile with the specified path.
// The file is expected to exist, else an error is returned.
func NewWatched(filename string) (*WatchedLoader, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	err = watcher.Add(filename)
	if err != nil {
		return nil, err
	}
	return &WatchedLoader{filename: filename, watcher: watcher}, nil
}

// Close releases any resources allocated to the WatchedFile.
// Once closed the file will no longer be monitored for changes.
func (l *WatchedLoader) Close() error {
	return l.watcher.Close()
}

// Load returns the current content of the watched file.
func (l *WatchedLoader) Load() ([]byte, error) {
	return ioutil.ReadFile(l.filename)
}

// Watch blocks until the watched file is altered.
// Alteration is relative to the construction of the WatchedFile, or the last
// call to Watch, whichever is more recent.
// The Watch may be cancelled by providing a context WithCancel and
// calling the cancel function.
// The returned error is nil if file is changed, or an error if the context has
// been cancelled or the WatchedFile closed.
func (l *WatchedLoader) Watch(ctx context.Context) error {
	for {
		select {
		case _, ok := <-l.watcher.Events:
			if !ok {
				return context.Canceled
			}
			return nil
		case e, ok := <-l.watcher.Errors:
			if !ok {
				return context.Canceled
			}
			return e
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
