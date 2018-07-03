// Copyright © 2017 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package properties provides a Java properties format Getter for config.
package properties

import (
	"strings"

	"github.com/magiconair/properties"
	"github.com/warthog618/config/keys"
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
	if p, ok := keys.IsArrayLen(key); ok {
		if v, ok := r.config.Get(p); ok {
			return strings.Count(v, r.listSeparator) + 1, ok
		}
	}
	if p, i := keys.ParseArrayElement(key); len(i) == 1 {
		if v, ok := r.config.Get(p); ok {
			if len(r.listSeparator) > 0 && strings.Contains(v, r.listSeparator) {
				l := strings.Split(v, r.listSeparator)
				if i[0] < len(l) {
					return l[i[0]], true
				}
				return nil, false
			}
		}
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
