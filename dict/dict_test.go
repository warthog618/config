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
	g := dict.New()
	require.NotNil(t, g)
	// test provides config.Getter interface.
	cfg := config.New()
	cfg.AppendGetter(g)
}

func TestGetter(t *testing.T) {
	g := dict.New()
	require.NotNil(t, g)
	v, ok := g.Get("a")
	assert.False(t, ok)
	assert.Nil(t, v)
	g.Set("a", 1)
	v, ok = g.Get("a")
	assert.True(t, ok)
	assert.Equal(t, 1, v)
}

func TestGetterWithConfig(t *testing.T) {
	config := map[string]interface{}{"a": 1}
	g := dict.New(dict.WithConfig(config))
	require.NotNil(t, g)
	v, ok := g.Get("a")
	assert.True(t, ok)
	assert.Equal(t, 1, v)
}

func TestGetterSet(t *testing.T) {
	g := dict.New()
	require.NotNil(t, g)
	v, ok := g.Get("a")
	assert.False(t, ok)
	assert.Nil(t, v)
	g.Set("a", 1)
	v, ok = g.Get("a")
	assert.True(t, ok)
	assert.Equal(t, 1, v)
	g.Set("a", 32)
	v, ok = g.Get("a")
	assert.True(t, ok)
	assert.Equal(t, 32, v)
}

func BenchmarkGet(b *testing.B) {
	g := dict.New(dict.WithConfig(map[string]interface{}{"nested.leaf": "44"}))
	for n := 0; n < b.N; n++ {
		g.Get("nested.leaf")
	}
}
