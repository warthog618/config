// Copyright © 2018 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package bytes

// Bytes is an example
type Bytes struct {
	b []byte
}

func New(b []byte) *Bytes {
	return &Bytes{b: b}
}

func (b *Bytes) Load() ([]byte, error) {
	return b.b, nil
}
