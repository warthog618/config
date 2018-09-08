// Copyright © 2018 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package file

// Option is a construction option for a Blob.
type Option interface {
	applyOption(l *Loader)
}

// WatcherOption enables a watcher on the file.
type WatcherOption struct {
}

func (WatcherOption) applyOption(l *Loader) {
	l.watcher = true
}

// WithWatcher is an Option that enables watching of the file. This is an option
// to ensure it can only set at construction time, so the watcher is a singleton
// and is created before the file is initially loaded.
// Note that watched files must be closed to terminate the watch.
func WithWatcher() WatcherOption {
	return WatcherOption{}
}
