// Copyright © 2018 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package toml

import (
	toml "github.com/BurntSushi/toml"
)

// NewDecoder returns a TOML decoder.
func NewDecoder() Decoder {
	return Decoder{}
}

// Decoder provides the Decoder API required by config.Source.
type Decoder struct{}

// Decode unmarshals an array of bytes containing TOML text.
func (d Decoder) Decode(b []byte, v interface{}) error {
	return toml.Unmarshal(b, v)
}
