// Copyright © 2017 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package dict provides a simple Reader that wraps a key/value map.
package dict

// Reader is a simple Reader that wraps a key/value map.
type Reader struct {
	// set of keys (node or leaf) masked.
	config map[string]interface{}
}

// New returns a dict Reader.
// The key/value map is initially empty and must be populated using calls to Set.
func New() *Reader {
	return &Reader{map[string]interface{}{}}
}

// Set adds a value to the key/value map.
func (r *Reader) Set(key string, v interface{}) {
	r.config[key] = v
}

// Read returns the value from the dict config.
func (r *Reader) Read(key string) (interface{}, bool) {
	v, ok := r.config[key]
	return v, ok
}
