// Copyright © 2017 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package dict provides a simple Reader that wraps a key/value map.
package dict

import "sync"

// Reader is a simple Reader that wraps a key/value map.
// The Reader is mutable, though only by setting keys, and
// is safe to call from multiple goroutines.
type Reader struct {
	mu sync.RWMutex
	// set of keys (node or leaf).
	config map[string]interface{}
}

// New returns a dict Reader.
// The key/value map is initially empty and must be populated using calls to Set.
func New() *Reader {
	return &Reader{config: map[string]interface{}{}}
}

// Set adds a value to the key/value map.
func (r *Reader) Set(key string, v interface{}) {
	r.mu.Lock()
	r.config[key] = v
	r.mu.Unlock()
}

// Read returns the value from the dict config.
func (r *Reader) Read(key string) (interface{}, bool) {
	r.mu.RLock()
	v, ok := r.config[key]
	r.mu.RUnlock()
	return v, ok
}
