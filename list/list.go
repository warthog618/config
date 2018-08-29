// Copyright © 2018 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package list contains helpers to convert values from strings to lists.
package list

import "strings"

// NewSplitter creates a splitter that splits lists separated by sep.
func NewSplitter(sep string) Splitter {
	return splitter{sep}
}

// Splitter converts a string containing a slice into a slice,
// or returns the string unaltered.
type Splitter interface {
	Split(string) interface{}
}

// Splitter splits a string containing a colon separated list
// and returns it as a slice.
type splitter struct {
	sep string
}

// Split converts a string containing a separated list into a slice, or
// returns the string unaltered.
func (s splitter) Split(v string) interface{} {
	if strings.Contains(v, s.sep) {
		return strings.Split(v, s.sep)
	}
	return v
}
