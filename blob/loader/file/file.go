// Copyright © 2018 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package file provides a loader from file for config.
package file

import (
	"io/ioutil"

	"github.com/fsnotify/fsnotify"
)

// Loader provides reads configuration from the local filesystem.
type Loader struct {
	filename string
	watcher  bool
}

// New creates a loader with the specified path.
func New(filename string, options ...Option) (*Loader, error) {
	l := Loader{filename: filename}
	for _, option := range options {
		option.applyOption(&l)
	}
	return &l, nil
}

// Load returns the current content of the file.
// The file is expected to exist and be readable, else an error is returned.
func (l *Loader) Load() ([]byte, error) {
	return ioutil.ReadFile(l.filename)
}

// NewWatcher returns a channel of update events the loader.
// The watcher must be enabled using the WithWatch construction option.
// The watcher will send nil events when the loader has changed.
// If a terminal error occurs it is sent to the update channel which is then closed.
// The watcher will exit when the done is closed or a terminal error occurs.
func (l *Loader) NewWatcher(done <-chan struct{}) <-chan error {
	if !l.watcher {
		return nil
	}
	update := make(chan error)
	w := watcher{}
	go w.watcher(l.filename, done, update)
	return update
}

// watcher watches a file for changes.
type watcher struct {
}

func (w *watcher) watcher(filename string, done <-chan struct{}, updatech chan error) {
	update := func(err error) {
		select {
		case updatech <- err:
		case <-done:
		}
	}
	defer close(updatech)
	fsn, err := fsnotify.NewWatcher()
	if err != nil {
		update(err)
		return
	}
	err = fsn.Add(filename)
	if err != nil {
		update(err)
		return
	}
	defer fsn.Close()
	// immediate update to trigger load AFTER fsnotify is active
	update(nil)
	for {
		select {
		case _, ok := <-fsn.Events:
			if !ok {
				return
			}
			update(nil)
		case _, ok := <-fsn.Errors:
			if !ok {
				return
			}
		case <-done:
			return
		}
	}
}
