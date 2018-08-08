// Copyright © 2017 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package config

import (
	"time"

	"github.com/warthog618/config/cfgconv"
)

// Value contains a value read from the configuration.
type Value struct {
	value interface{}
	// error handler for type conversions
	eh ErrorHandler
}

// Bool converts the value to a bool.
// Returns false if conversion is not possible.
func (v Value) Bool() bool {
	b, err := cfgconv.Bool(v.value)
	if err != nil && v.eh != nil {
		v.eh(err)
	}
	return b
}

// Duration gets the value corresponding to the key and converts it to
// a time.Duration.
// Returns 0 if conversion is not possible.
func (v Value) Duration() time.Duration {
	d, err := cfgconv.Duration(v.value)
	if err != nil && v.eh != nil {
		v.eh(err)
	}
	return d
}

// Float converts the value to a float64.
// Returns 0 if conversion is not possible.
func (v Value) Float() float64 {
	f, err := cfgconv.Float(v.value)
	if err != nil && v.eh != nil {
		v.eh(err)
	}
	return f
}

// Int converts the value to an int64.
// Returns 0 if conversion is not possible.
func (v Value) Int() int64 {
	i, err := cfgconv.Int(v.value)
	if err != nil && v.eh != nil {
		v.eh(err)
	}
	return i
}

// IntSlice converts the value to a slice of int64s.
// Returns nil if conversion is not possible.
func (v Value) IntSlice() []int64 {
	is, err := cfgconv.IntSlice(v.value)
	if err != nil && v.eh != nil {
		v.eh(err)
	}
	return is
}

// Slice converts the value to a slice of []interface{}.
// Returns nil if conversion is not possible.
func (v Value) Slice() []interface{} {
	s, err := cfgconv.Slice(v.value)
	if err != nil && v.eh != nil {
		v.eh(err)
	}
	return s
}

// String converts the value to a string.
// Returns an empty string if conversion is not possible.
func (v Value) String() string {
	s, err := cfgconv.String(v.value)
	if err != nil && v.eh != nil {
		v.eh(err)
	}
	return s
}

// StringSlice converts the value to a slice of string.
// Returns nil if conversion is not possible.
func (v Value) StringSlice() []string {
	ss, err := cfgconv.StringSlice(v.value)
	if err != nil && v.eh != nil {
		v.eh(err)
	}
	return ss
}

// Time converts the value to a time.Time.
// Returns time.Time{} if conversion is not possible.
func (v Value) Time() time.Time {
	t, err := cfgconv.Time(v.value)
	if err != nil && v.eh != nil {
		v.eh(err)
	}
	return t
}

// Uint converts the value to a iint64.
// Returns 0 if conversion is not possible.
func (v Value) Uint() uint64 {
	u, err := cfgconv.Uint(v.value)
	if err != nil && v.eh != nil {
		v.eh(err)
	}
	return u
}

// UintSlice converts the value to a slice of uint64.
// Returns nil if conversion is not possible.
func (v Value) UintSlice() []uint64 {
	us, err := cfgconv.UintSlice(v.value)
	if err != nil && v.eh != nil {
		v.eh(err)
	}
	return us
}

// Value returns the raw value.
func (v Value) Value() interface{} {
	return v.value
}
