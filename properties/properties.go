// Copyright © 2017 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package properties provides a Java properties format reader for config.
package properties

import (
	"strings"

	"github.com/magiconair/properties"
)

// Reader provides the mapping from JSON to a config.Reader.
type Reader struct {
	config *properties.Properties
	// The separator for slices stored in string values.
	listSeparator string
}

// Read returns the value for a given key and true if found, or
// nil and false if not.
func (r *Reader) Read(key string) (interface{}, bool) {
	if v, ok := r.config.Get(key); ok {
		if len(r.listSeparator) > 0 && strings.Contains(v, r.listSeparator) {
			return strings.Split(v, r.listSeparator), ok
		}
		return v, true
	}
	return nil, false
}

// Option is a function which modifies a Reader at construction time.
type Option func(*Reader)

// WithListSeparator sets the separator between slice fields in the env namespace.
// The default separator is ":"
func WithListSeparator(separator string) Option {
	return func(r *Reader) {
		r.listSeparator = separator
	}
}

// NewBytes returns a properties reader that reads config from []byte.
func NewBytes(cfg []byte, options ...Option) (*Reader, error) {
	config, err := properties.Load(cfg, properties.UTF8)
	if err != nil {
		return nil, err
	}
	return new(config, options...), nil
}

// NewFile returns a properties reader that reads config from a named file.
func NewFile(filename string, options ...Option) (*Reader, error) {
	config, err := properties.LoadFile(filename, properties.UTF8)
	if err != nil {
		return nil, err
	}
	return new(config, options...), nil
}

func new(config *properties.Properties, options ...Option) *Reader {
	r := Reader{config, ","}
	for _, option := range options {
		option(&r)
	}
	return &r
}
