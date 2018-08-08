// Copyright © 2018 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package config

import "sync"

// Notifier provides broadcast signalling of an anonymous event.
// It provides many to many connectivity - any number of sources
// may trigger the notification and any number of sinks may wait on it.
type Notifier struct {
	mu sync.RWMutex
	ch chan struct{}
}

// NewNotifier creates and returns a new Signal.
func NewNotifier() *Notifier {
	return &Notifier{ch: make(chan struct{})}
}

// Notified returns a channel which is closed when the signal
// is triggered.
func (s *Notifier) Notified() <-chan struct{} {
	s.mu.RLock()
	ch := s.ch
	s.mu.RUnlock()
	return ch
}

// Notify triggers the signal event.
func (s *Notifier) Notify() {
	s.mu.Lock()
	close(s.ch)
	s.ch = make(chan struct{})
	s.mu.Unlock()
}
