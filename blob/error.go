// Copyright © 2017 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package blob

type withTemporary struct {
	error
}

// WithTemporary wraps an error so it supports the temporary interface
// and will be marked as temporary.
func WithTemporary(err error) error {
	if err == nil {
		return nil
	}
	return withTemporary{error: err}
}

func (w withTemporary) Temporary() bool {
	return true
}

func (w withTemporary) Cause() error {
	return w.error
}

func (w withTemporary) Error() string {
	return w.error.Error()
}
