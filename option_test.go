// Copyright © 2018 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAliasWithSeparator(t *testing.T) {
	a := NewAlias()
	require.NotNil(t, a)
	assert.Equal(t, ".", a.pathSep)
	a = NewAlias(WithSeparator("X"))
	require.NotNil(t, a)
	assert.Equal(t, "X", a.pathSep)
}

func TestNewWithDefault(t *testing.T) {
	g := mockGetter{"a": 42}
	d := mockGetter{"a": 45, "d": 47}
	c := New(g, WithDefault(d))
	require.NotNil(t, c)
	val, err := c.Get("a")
	assert.Nil(t, err)
	assert.Equal(t, 42, val.Int())
	val, err = c.Get("d")
	assert.Nil(t, err)
	assert.Equal(t, 47, val.Int())
}

func TestNewWithSeparator(t *testing.T) {
	g := mockGetter{"a": 42}
	c := New(g)
	require.NotNil(t, c)
	assert.Equal(t, ".", c.pathSep)
	c = New(g, WithSeparator("X"))
	require.NotNil(t, c)
	assert.Equal(t, "X", c.pathSep)
}

func TestNewWithTag(t *testing.T) {
	g := mockGetter{"a": 42}
	c := New(g)
	require.NotNil(t, c)
	assert.Equal(t, "config", c.tag)
	c = New(g, WithTag("bogus"))
	require.NotNil(t, c)
	assert.Equal(t, "bogus", c.tag)
}

type mockGetter map[string]interface{}

func (m mockGetter) Get(key string) (interface{}, bool) {
	v, ok := m[key]
	return v, ok
}
