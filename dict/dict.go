// Copyright © 2017 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package dict provides a simple Getter that wraps a key/value map.
package dict

import "sync"

// Getter is a simple Getter that wraps a key/value map.
// The Getter is mutable, though only by setting keys, and
// is safe to call from multiple goroutines.
type Getter struct {
	mu sync.RWMutex
	// set of keys (node or leaf).
	config map[string]interface{}
}

// New returns a dict Getter.
// The key/value map is initially empty and must be populated using calls to Set.
func New() *Getter {
	return &Getter{config: map[string]interface{}{}}
}

// Set adds a value to the key/value map.
func (r *Getter) Set(key string, v interface{}) {
	r.mu.Lock()
	r.config[key] = v
	r.mu.Unlock()
}

// Get returns the value from the dict config.
func (r *Getter) Get(key string) (interface{}, bool) {
	r.mu.RLock()
	v, ok := r.config[key]
	r.mu.RUnlock()
	return v, ok
}
