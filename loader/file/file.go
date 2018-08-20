// Copyright © 2018 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package file

import (
	"context"
	"io/ioutil"
	"os"

	"github.com/fsnotify/fsnotify"
)

// File provides a read-once source of configuration from the local filesystem.
type File struct {
	filename string
}

// New creates a File loader with the specified path.
func New(filename string) *File {
	return &File{filename: filename}
}

// Load returns the current content of the file.
// The file is expected to exist, so it it is missing, or otherwise unreadable,
// an error is returned.
func (f *File) Load() ([]byte, error) {
	return ioutil.ReadFile(f.filename)
}

// WatchedFile provides an active source of configuration from the local filesystem.
type WatchedFile struct {
	filename string
	watcher  *fsnotify.Watcher
}

// NewWatchedFile creates a WatchedFile with the specified path.
// The file is expected to exist, but if it does not then its configuration
// is assumed to be empty, rather than an error condition.
func NewWatchedFile(filename string) (*WatchedFile, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	err = watcher.Add(filename)
	if err != nil {
		return nil, err
	}
	return &WatchedFile{filename: filename, watcher: watcher}, nil
}

// Close releases any resources allocated to the WatchedFile.
// Once closed the file will no longer be monitored for changes.
func (f *WatchedFile) Close() error {
	return f.watcher.Close()
}

// Load returns the current content of the watched file.
// If the file is missing its configuration is assumed to be empty,
// rather than indicating an error condition.
func (f *WatchedFile) Load() ([]byte, error) {
	b, err := ioutil.ReadFile(f.filename)
	if err != nil {
		if _, ok := err.(*os.PathError); ok {
			// treat missing as empty
			return make([]byte, 0), nil
		}
		return nil, err
	}
	return b, nil
}

// Watch blocks until the watched file is altered.
// Alteration is relative to the construction of the WatchedFile, or the last
// call to Watch, whichever is more recent.
// The Watch may be cancelled by providing a context WithCancel and
// calling the cancel function.
// The returned error is nil if file is changed, or an error if the context has
// been cancelled or the WatchedFile closed.
func (f *WatchedFile) Watch(ctx context.Context) error {
	for {
		select {
		case e, ok := <-f.watcher.Events:
			if !ok {
				return context.Canceled
			}
			switch e.Op {
			// !!! any need to handle special cases here??
			case fsnotify.Chmod, fsnotify.Write, fsnotify.Rename:
				return nil
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
