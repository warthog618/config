// Copyright © 2017 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package dict provides a simple getter that wraps a key/value map.
package dict

import "sync"

// Getter is a simple getter that wraps a key/value map.
// The Getter is mutable, though only by setting keys, and
// is safe to call from multiple goroutines.
type Getter struct {
	mu sync.RWMutex
	// set of keys (node or leaf).
	config map[string]interface{}
}

// New returns a dict Getter.
// The key/value map is initially empty and must be populated using
// WithConfig or calls to Set.
func New(options ...Option) *Getter {
	g := Getter{}
	for _, option := range options {
		option(&g)
	}
	if g.config == nil {
		g.config = map[string]interface{}{}
	}
	return &g
}

// Option is a function which modifies a Getter at construction time.
type Option func(*Getter)

// WithConfig provides the config map, rather than having the Getter create a
// new one.
// Note that the Getter takes ownership of the map, so any subsequent sets
// will update the passed map.
func WithConfig(config map[string]interface{}) Option {
	return func(c *Getter) {
		c.config = config
	}
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
