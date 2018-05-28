// Copyright © 2017 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package properties provides a Java properties format Getter for config.
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

// Option is a function which modifies a Getter at construction time.
type Option func(*Getter)

// WithListSeparator sets the separator between slice fields in the env namespace.
// The default separator is ":"
func WithListSeparator(separator string) Option {
	return func(r *Getter) {
		r.listSeparator = separator
	}
}

// NewBytes returns a properties Getter that reads config from []byte.
func NewBytes(cfg []byte, options ...Option) (*Getter, error) {
	config, err := properties.Load(cfg, properties.UTF8)
	if err != nil {
		return nil, err
	}
	return new(config, options...), nil
}

// NewFile returns a properties Getter that reads config from a named file.
func NewFile(filename string, options ...Option) (*Getter, error) {
	config, err := properties.LoadFile(filename, properties.UTF8)
	if err != nil {
		return nil, err
	}
	return new(config, options...), nil
}

func new(config *properties.Properties, options ...Option) *Getter {
	r := Getter{config, ","}
	for _, option := range options {
		option(&r)
	}
	return &r
}
