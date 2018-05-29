// Copyright © 2017 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package toml provides a TOML format getter for config.
package toml

import gotoml "github.com/pelletier/go-toml"

// Getter provides the mapping from TOML to a config.Getter.
// The Getter parses the TOML only at construction time, so its config state
// is effectively immutable.
type Getter struct {
	config *gotoml.Tree
}

// Get returns the value for a given key and true if found, or
// nil and false if not.
func (r *Getter) Get(key string) (interface{}, bool) {
	v := r.config.Get(key)
	if v == nil {
		return nil, false
	}
	if _, ok := v.(*gotoml.Tree); ok {
		return nil, false
	}
	return v, true
}

// NewBytes returns a TOML Getter that reads config from []byte.
func NewBytes(cfg []byte) (*Getter, error) {
	config, err := gotoml.Load(string(cfg))
	if err != nil {
		return nil, err
	}
	return &Getter{config}, nil
}

// NewFile returns a TOML Getter that reads config from a named file.
func NewFile(filename string) (*Getter, error) {
	config, err := gotoml.LoadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Getter{config}, nil
}
