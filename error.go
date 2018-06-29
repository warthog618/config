// Copyright © 2017 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package config

import "errors"

// NotFoundError indicates that the Key could not be found in the config tree.
type NotFoundError struct {
	Key string
}

func (e NotFoundError) Error() string {
	return "config: key '" + e.Key + "' not found"
}

// UnmarshalError indicates an error occurred while unmarhalling config into
// a struct or map.  The error indicates the problematic Key and the specific
// error.
type UnmarshalError struct {
	Key string
	Err error
}

func (e UnmarshalError) Error() string {
	return "config: cannot unmarshal " + e.Key + " - " + e.Err.Error()
}

// ErrInvalidStruct indicates Unmarshal was provided an object to populate
// which is not a pointer to struct.
var ErrInvalidStruct = errors.New("unmarshal: provided obj is not pointer to struct")
