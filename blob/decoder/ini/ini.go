// Copyright © 2018 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package ini provides an INI format decoder for config.
package ini

import (
	"errors"
	"strings"

	ini "gopkg.in/ini.v1"
)

// NewDecoder returns a INI decoder.
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

// WithListSeparator sets the separator between slice fields in the ini space.
// The default separator is ","
func WithListSeparator(separator string) Option {
	return func(d *Decoder) {
		d.listSeparator = separator
	}
}

// Decoder provides the Decoder API required by config.Source.
type Decoder struct {
	listSeparator string
}

// Decode unmarshals an array of bytes containing ini text.
func (d Decoder) Decode(b []byte, v interface{}) error {
	mp, ok := v.(*map[string]interface{})
	if !ok {
		return errors.New("Decode only supports map[string]interface{}")
	}
	f, err := ini.Load(b)
	if err != nil {
		return err
	}
	for _, section := range f.Sections() {
		if section.Name() == "DEFAULT" {
			d.loadSection(section, *mp)
		} else {
			sm := make(map[string]interface{})
			(*mp)[section.Name()] = sm
			d.loadSection(section, sm)
		}
	}
	return nil
}

func (d Decoder) loadSection(s *ini.Section, m map[string]interface{}) {
	for _, key := range s.Keys() {
		v := key.String()
		k := key.Name()
		if len(d.listSeparator) > 0 && strings.Contains(v, d.listSeparator) {
			m[k] = strings.Split(v, d.listSeparator)
		} else {
			m[k] = v
		}
	}
}
