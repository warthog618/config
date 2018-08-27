// Copyright © 2018 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package yaml

import yaml "gopkg.in/yaml.v2"

// NewDecoder returns a YAML decoder.
func NewDecoder() Decoder {
	return Decoder{}
}

// Decoder provides the Decoder API required by config.Source.
type Decoder struct{}

// Decode unmarshals an array of bytes containing YAML text.
func (d Decoder) Decode(b []byte, v interface{}) error {
	return yaml.Unmarshal(b, v)
}
