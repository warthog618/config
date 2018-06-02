// Copyright © 2017 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package config_test

import (
	"errors"
	"fmt"
	"testing"

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
