// Copyright © 2018 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package config

import "sync"

// Signal provides broadcast signalling of an event.
// It provides many to many connectivity - any number of sources
// may trigger the signal and any number of sinks may wait on it.
type Signal struct {
	mu sync.RWMutex
	ch chan struct{}
}

// NewSignal creates and returns a new Signal.
func NewSignal() *Signal {
	return &Signal{ch: make(chan struct{})}
}

// Signalled returns a channel which is closed when the signal
// is triggered.
func (s *Signal) Signalled() <-chan struct{} {
	s.mu.RLock()
	ch := s.ch
	s.mu.RUnlock()
	return ch
}

// Signal triggers the signal event.
func (s *Signal) Signal() {
	s.mu.Lock()
	close(s.ch)
	s.ch = make(chan struct{})
	s.mu.Unlock()
}
