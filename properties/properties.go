// Copyright © 2017 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package properties provides a Java properties format getter for config.
package properties

import (
	"strings"

	"github.com/magiconair/properties"
)

// Getter provides the mapping from a properties file to a config.Getter.
// The Getter parses the properties only at construction time, so its config state
// is effectively immutable.
type Getter struct {
	config *properties.Properties
	// The separator for slices stored in string values.
	listSeparator string
}

// New returns a properties Getter.
func New(options ...Option) (*Getter, error) {
	r := Getter{listSeparator: ","}
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
	if v, ok := r.config.Get(key); ok {
		if len(r.listSeparator) > 0 && strings.Contains(v, r.listSeparator) {
			return strings.Split(v, r.listSeparator), ok
		}
		return v, true
	}
	return nil, false
}

// Option is a function that modifies the Getter during construction,
// returning any error that may have occurred.
type Option func(*Getter) error

// WithListSeparator sets the separator between slice fields in the properties space.
// The default separator is ","
func WithListSeparator(separator string) Option {
	return func(r *Getter) error {
		r.listSeparator = separator
		return nil
	}
}

// FromBytes uses the []bytes as the source of properties configuration.
func FromBytes(cfg []byte) Option {
	return func(g *Getter) error {
		config, err := properties.Load(cfg, properties.UTF8)
		if err != nil {
			return err
		}
		g.config = config
		return nil
	}
}

// FromFile uses filename as the source of properties configuration.
func FromFile(filename string) Option {
	return func(g *Getter) error {
		config, err := properties.LoadFile(filename, properties.UTF8)
		if err != nil {
			return err
		}
		g.config = config
		return nil
	}
}
