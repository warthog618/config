// Copyright © 2017 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package toml provides a TOML format Getter for config.
package toml

import (
	gotoml "github.com/pelletier/go-toml"
	"github.com/warthog618/config/tree"
)

// Getter provides the mapping from TOML to a config.Getter.
// The Getter parses the TOML only at construction time, so its config state
// is effectively immutable.
type Getter struct {
	config map[string]interface{}
	sep    string
}

// New returns a properties Getter.
func New(options ...Option) (*Getter, error) {
	g := Getter{sep: "."}
	for _, option := range options {
		err := option(&g)
		if err != nil {
			return nil, err
		}
	}
	return &g, nil
}

// Get returns the value for a given key and true if found, or
// nil and false if not.
func (r *Getter) Get(key string) (interface{}, bool) {
	return tree.Get(r.config, key, r.sep)
}

// Option is a function that modifies the Getter during construction,
// returning any error that may have occurred.
type Option func(*Getter) error

// FromBytes uses the []bytes as the source of TOML configuration.
func FromBytes(cfg []byte) Option {
	return func(g *Getter) error {
		config, err := gotoml.Load(string(cfg))
		if err != nil {
			return err
		}
		g.config = config.ToMap()
		return nil
	}
}

// FromFile uses filename as the source of TOML configuration.
func FromFile(filename string) Option {
	return func(g *Getter) error {
		config, err := gotoml.LoadFile(filename)
		if err != nil {
			return err
		}
		g.config = config.ToMap()
		return nil
	}
}
