// Copyright © 2018 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package bytes provides a loader from []byte for config.
package bytes

// Loader provides a source of configuration from an array of bytes in memory.
type Loader struct {
	b []byte
}

// New creates a loader that returns the provided bytes.
// The bytes should not be changed after being passed to New.
func New(b []byte) *Loader {
	return &Loader{b: b}
}

// Load returns the bytes provided to New.
func (l *Loader) Load() ([]byte, error) {
	return l.b, nil
}
