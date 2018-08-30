// Copyright © 2018 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package properties provides a Java properties format decoder for config.
package properties

import (
	"errors"
	"strings"

	"github.com/magiconair/properties"
)

// NewDecoder returns a properties decoder.
func NewDecoder(options ...Option) Decoder {
	d := Decoder{listSeparator: ","}
	for _, option := range options {
		option(&d)
	}
	return d
}

// Option is a function that modifies the Decdoder during construction,
// returning any error that may have occurred.
type Option func(*Decoder)

// WithListSeparator sets the separator between slice fields in the properties
// space. The default separator is ","
func WithListSeparator(separator string) Option {
	return func(d *Decoder) {
		d.listSeparator = separator
	}
}

// Decoder provides the Decoder API required by config.Source.
type Decoder struct {
	listSeparator string
}

// Decode unmarshals an array of bytes containing properties text.
func (d Decoder) Decode(b []byte, v interface{}) error {
	mp, ok := v.(*map[string]interface{})
	if !ok {
		return errors.New("Decode only supports map[string]interface{}")
	}
	config, err := properties.Load(b, properties.UTF8)
	if err != nil {
		return err
	}
	m := config.Map()
	for key, val := range m {
		if len(d.listSeparator) > 0 && strings.Contains(val, d.listSeparator) {
			(*mp)[key] = strings.Split(val, d.listSeparator)
		} else {
			(*mp)[key] = val
		}
	}
	return nil
}
