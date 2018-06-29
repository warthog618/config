// Copyright © 2018 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package config

import "sync"

// NewStack creates a Stack.
func NewStack(gg ...Getter) *Stack {
	s := Stack{}
	for _, g := range gg {
		if g != nil {
			s.gg = append(s.gg, g)
		}
	}
	return &s
}

// Stack attempts a get using a list of Getters, in the order provided,
// returning the first result found.
type Stack struct {
	// RWLock covering gg.
	// It does not prevent concurrent access to the getters themselves,
	// only to the gg array itself.
	mu sync.RWMutex
	// A list of Getters providing config key/value pairs.
	gg []Getter
}

// Append appends a getter to the set of getters for the Stack.
// This means this getter is only used as a last resort, relative to
// the existing getters.
func (s *Stack) Append(g Getter) {
	if g == nil {
		return
	}
	s.mu.Lock()
	s.gg = append(s.gg, g)
	s.mu.Unlock()
}

// Get gets the raw value corresponding to the key.
// It iterates through the list of getters, searching for a matching key.
// Returns the first match found, or an error if none is found.
func (s *Stack) Get(key string) (interface{}, bool) {
	s.mu.RLock()
	for _, g := range s.gg {
		if v, ok := g.Get(key); ok {
			s.mu.RUnlock()
			return v, ok
		}
	}
	s.mu.RUnlock()
	return nil, false
}

// Insert inserts a getter to the set of getters for the Stack.
// This means this getter is used before the existing getters.
func (s *Stack) Insert(g Getter) {
	if g == nil {
		return
	}
	s.mu.Lock()
	s.gg = append([]Getter{g}, s.gg...)
	s.mu.Unlock()
}
