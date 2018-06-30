// Copyright © 2017 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package json provides a JSON format Getter for config.
package json

import (
	"encoding/json"
	"io/ioutil"

	"github.com/warthog618/config/tree"
)

// Getter provides the mapping from JSON to a config.Getter.
// The Getter parses the JSON only at construction time, so its config state
// is effectively immutable.
type Getter struct {
	config map[string]interface{}
	sep    string
}

// Get returns the value for a given key and true if found, or
// nil and false if not.
func (r *Getter) Get(key string) (interface{}, bool) {
	return tree.GetFromMSI(r.config, key, r.sep)
}

// New returns a JSON Getter.
func New(options ...Option) (*Getter, error) {
	g := Getter{sep: "."}
	for _, option := range options {
		if err := option(&g); err != nil {
			return nil, err
		}
	}
	return &g, nil
}

// Option is a function that modifies the Getter during construction,
// returning any error that may have occurred.
type Option func(*Getter) error

// FromBytes uses the []bytes as the source of JSON configuration.
func FromBytes(cfg []byte) Option {
	return func(g *Getter) error {
		return fromBytes(g, cfg)
	}
}

// FromFile uses filename as the source of JSON configuration.
func FromFile(filename string) Option {
	return func(g *Getter) error {
		b, err := ioutil.ReadFile(filename)
		if err != nil {
			return err
		}
		return fromBytes(g, b)
	}
}

func fromBytes(g *Getter, cfg []byte) error {
	var config map[string]interface{}
	err := json.Unmarshal(cfg, &config)
	if err != nil {
		return err
	}
	g.config = config
	return nil
}
