// Copyright © 2017 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package yaml provides a YAML format Getter for config.
package yaml

import (
	"io/ioutil"

	"github.com/warthog618/config/tree"
	"gopkg.in/yaml.v2"
)

// Getter provides the mapping from YAML to a config.Getter.
// The Getter parses the YAML only at construction time, so its config state
// is effectively immutable.
type Getter struct {
	config map[interface{}]interface{}
	sep    string
}

// New returns a YAML Getter.
func New(options ...Option) (*Getter, error) {
	r := Getter{sep: "."}
	for _, option := range options {
		err := option(&r)
		if err != nil {
			return nil, err
		}
	}
	return &r, nil
}

// Get returns the value for a given key and true if found, or
// nil and false if not.
func (r *Getter) Get(key string) (interface{}, bool) {
	return tree.Get(r.config, key, r.sep)
}

// Option is a function that modifies the Getter during construction,
// returning any error that may have occurred.
type Option func(*Getter) error

// FromBytes uses the []bytes as the source of YAML configuration.
func FromBytes(cfg []byte) Option {
	return func(g *Getter) error {
		return fromBytes(g, cfg)
	}
}

// FromFile uses filename as the source of YAML configuration.
func FromFile(filename string) Option {
	return func(g *Getter) error {
		b, err := ioutil.ReadFile(filename)
		if err != nil {
			return err
		}
		return fromBytes(g, b)
	}
}

func fromBytes(g *Getter, b []byte) error {
	var config map[interface{}]interface{}
	err := yaml.Unmarshal(b, &config)
	if err != nil {
		return err
	}
	g.config = config
	return nil
}
