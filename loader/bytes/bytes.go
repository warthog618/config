// Copyright © 2018 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package bytes

// Bytes provides a source of configuration from an array of bytes in memory.
type Bytes struct {
	b []byte
}

// New creates a loader that returns the provided bytes.
// The bytes should not be changed after being passed to New.
func New(b []byte) *Bytes {
	return &Bytes{b: b}
}

// Load returns the bytes provided to New.
func (b *Bytes) Load() ([]byte, error) {
	return b.b, nil
}
