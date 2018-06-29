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

func TestNewConfigWithSeparator(t *testing.T) {
	g := mockGetter{"a": 42}
	c := NewConfig(g)
	require.NotNil(t, c)
	assert.Equal(t, ".", c.pathSep)
	c = NewConfig(g, WithSeparator("X"))
	require.NotNil(t, c)
	assert.Equal(t, "X", c.pathSep)
}

func TestNewConfigWithTag(t *testing.T) {
	g := mockGetter{"a": 42}
	c := NewConfig(g)
	require.NotNil(t, c)
	assert.Equal(t, "config", c.tag)
	c = NewConfig(g, WithTag("bogus"))
	require.NotNil(t, c)
	assert.Equal(t, "bogus", c.tag)
}

func TestNewMustWithErrorHandler(t *testing.T) {
	mr := mockGetter{"a.b.c_d": true}
	var err error
	f := func(e error) {
		err = e
	}
	c := NewMust(mr, WithErrorHandler(f))
	v := c.Get("a.b.c_d")
	assert.Nil(t, err)
	assert.Equal(t, true, v)
	v = c.Get("")
	assert.IsType(t, NotFoundError{}, err)
	assert.Equal(t, nil, v)
}

func TestNewMustWithPanic(t *testing.T) {
	mr := mockGetter{"a.b.c_d": true}
	c := NewMust(mr, WithPanic())
	assert.NotPanics(t, func() { c.Get("a.b.c_d") })
	assert.Panics(t, func() { c.Get("") })
}

func TestNewMustWithTag(t *testing.T) {
	g := mockGetter{"a": 42}
	m := NewMust(g)
	require.NotNil(t, m)
	assert.Equal(t, "config", m.c.tag)
	m = NewMust(g, WithTag("bogus"))
	require.NotNil(t, m)
	assert.Equal(t, "bogus", m.c.tag)
}

type mockGetter map[string]interface{}

func (m mockGetter) Get(key string) (interface{}, bool) {
	v, ok := m[key]
	return v, ok
}
