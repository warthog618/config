// Copyright © 2018 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package file

import "github.com/fsnotify/fsnotify"

// Option is a construction option for a Blob.
type Option interface {
	applyOption(l *Loader) error
}

// WatcherOption enables a watcher on the file.
type WatcherOption struct {
}

func (WatcherOption) applyOption(l *Loader) error {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	err = w.Add(l.filename)
	if err != nil {
		return err
	}
	l.w = &watcher{w}
	return nil
}

// WithWatcher is an Option that enables watching of the file. This is an option
// to ensure it can only set at construction time, so the watcher is a singleton
// and is created before the file is initially loaded.
func WithWatcher() WatcherOption {
	return WatcherOption{}
}
