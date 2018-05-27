// Copyright © 2017 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package dict_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/warthog618/config"
	"github.com/warthog618/config/dict"
)

func TestNew(t *testing.T) {
	r := dict.New()
	require.NotNil(t, r)
	// test provides config.Reader interface.
	cfg := config.New()
	cfg.AppendReader(r)
}

func TestReader(t *testing.T) {
	r := dict.New()
	require.NotNil(t, r)
	v, ok := r.Read("a")
	assert.False(t, ok)
	assert.Nil(t, v)
	r.Set("a", 1)
	v, ok = r.Read("a")
	assert.True(t, ok)
	assert.Equal(t, 1, v)
}
