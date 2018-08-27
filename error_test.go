// Copyright © 2017 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package config_test

import (
	"fmt"
	"testing"

	"github.com/pkg/errors"

	"github.com/stretchr/testify/assert"
	"github.com/warthog618/config"
)

func TestNotFoundError(t *testing.T) {
	patterns := []string{"one", "two", "three"}
	for _, p := range patterns {
		f := func(t *testing.T) {
			e := config.NotFoundError{Key: p}
			expected := "config: key '" + e.Key + "' not found"
			assert.Equal(t, expected, e.Error())
		}
		t.Run(fmt.Sprintf("%x", p), f)
	}
}

func TestUnmarshalError(t *testing.T) {
	patterns := []struct {
		k   string
		err error
	}{
		{"one", errors.New("two")},
		{"three", errors.New("four")},
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			e := config.UnmarshalError{Key: p.k, Err: p.err}
			expected := "config: cannot unmarshal " + e.Key + " - " + e.Err.Error()
			assert.Equal(t, expected, e.Error())
		}
		t.Run(p.k, f)
	}
}

func TestWithTemporary(t *testing.T) {
	te := config.WithTemporary(nil)
	assert.Nil(t, te)

	e := errors.New("base error")
	te = config.WithTemporary(e)
	assert.Equal(t, e, errors.Cause(te))
	assert.Equal(t, e.Error(), te.Error())
	assert.False(t, config.IsTemporary(e))
	assert.True(t, config.IsTemporary(te))
}
