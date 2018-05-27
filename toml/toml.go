// Copyright © 2017 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package toml provides a TOML format reader for config.
package toml

import gotoml "github.com/pelletier/go-toml"

// Reader provides the mapping from TOML to a config.Reader.
type Reader struct {
	config *gotoml.Tree
}

// Read returns the value for a given key and true if found, or
// nil and false if not.
func (r *Reader) Read(key string) (interface{}, bool) {
	v := r.config.Get(key)
	if v == nil {
		return nil, false
	}
	if _, ok := v.(*gotoml.Tree); ok {
		return nil, false
	}
	return v, true
}

// NewBytes returns a TOML reader that reads config from []byte.
func NewBytes(cfg []byte) (*Reader, error) {
	config, err := gotoml.Load(string(cfg))
	if err != nil {
		return nil, err
	}
	return &Reader{config}, nil
}

// NewFile returns a TOML reader that reads config from a named file.
func NewFile(filename string) (*Reader, error) {
	config, err := gotoml.LoadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Reader{config}, nil
}
