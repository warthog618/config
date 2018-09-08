// Copyright © 2018 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/warthog618/config"
)

func TestNewAlias(t *testing.T) {
	a := config.NewAlias()
	assert.NotNil(t, a)
}

func TestNewRegexAlias(t *testing.T) {
	r := config.NewRegexAlias()
	assert.NotNil(t, r)
}

func TestNewAliasWithSeparator(t *testing.T) {
	a := config.NewAlias(config.WithSeparator("_"))
	assert.NotNil(t, a)
	g := &mockGetter{
		"a.b.c_d": "d",
		"a.b.c_e": "e",
	}
	v, ok := a.Get(g, "new")
	assert.False(t, ok)
	assert.Nil(t, v)

	a.Append("leaf", "a.b.c_d")
	v, ok = a.Get(g, "leaf")
	assert.True(t, ok)
	assert.Equal(t, "d", v)

	a.Append("node", "a.b")
	v, ok = a.Get(g, "node.c_e")
	assert.False(t, ok)
	assert.Nil(t, v)

	a.Append("node", "a.b.c")
	v, ok = a.Get(g, "node_e")
	assert.True(t, ok)
	assert.Equal(t, "e", v)
}

func TestWithAlias(t *testing.T) {
	g := &mockGetter{
		"a.b.c_d": "d",
		"a.b.c_e": "e",
	}
	a := config.NewAlias()
	d := config.Decorate(g, config.WithAlias(a))
	require.NotNil(t, d)
	v, ok := d.Get("a.b.c_d")
	assert.True(t, ok)
	assert.Equal(t, "d", v)
}

func TestWithRegexAlias(t *testing.T) {
	g := &mockGetter{
		"a.b.c_d": "d",
		"a.b.c_e": "e",
	}
	r := config.NewRegexAlias()
	d := config.Decorate(g, config.WithRegexAlias(r))
	require.NotNil(t, d)
	v, ok := d.Get("a.b.c_d")
	assert.True(t, ok)
	assert.Equal(t, "d", v)
}

func TestAliasAppend(t *testing.T) {
	a := config.NewAlias()
	require.NotNil(t, a)
	g := mockGetter{
		"a.b.c_d": "d",
		"a.b.c_e": "e",
	}
	v, ok := a.Get(&g, "new")
	assert.False(t, ok)
	assert.Nil(t, v)

	a.Append("new", "a.b.c_d")
	v, ok = a.Get(&g, "new")
	assert.True(t, ok)
	assert.Equal(t, "d", v)

	a.Append("new", "a.b.c_e")
	v, ok = a.Get(&g, "new")
	assert.True(t, ok)
	assert.Equal(t, "d", v)

	delete(g, "a.b.c_d")
	v, ok = a.Get(&g, "new")
	assert.True(t, ok)
	assert.Equal(t, "e", v)
}

func TestAliasInsert(t *testing.T) {
	a := config.NewAlias()
	require.NotNil(t, a)
	g := mockGetter{
		"a.b.c_d": "d",
		"a.b.c_e": "e",
	}
	v, ok := a.Get(&g, "new")
	assert.False(t, ok)
	assert.Nil(t, v)

	a.Insert("new", "a.b.c_d")
	v, ok = a.Get(&g, "new")
	assert.True(t, ok)
	assert.Equal(t, "d", v)

	a.Insert("new", "a.b.c_e")
	v, ok = a.Get(&g, "new")
	assert.True(t, ok)
	assert.Equal(t, "e", v)

	delete(g, "a.b.c_e")
	v, ok = a.Get(&g, "new")
	assert.True(t, ok)
	assert.Equal(t, "d", v)
}

func TestAliasGet(t *testing.T) {
	mr := &mockGetter{
		"a":     "a",
		"foo.a": "foo.a",
		"foo.b": "foo.b",
		"bar.b": "bar.b",
		"bar.c": "bar.c",
	}
	type alias struct {
		new string
		old string
	}
	patterns := []struct {
		name     string
		aa       []alias
		tp       string
		expected interface{}
		ok       bool
	}{
		{"alias to alias", []alias{{"c", "foo.b"}, {"d", "c"}}, "d", nil, false},
		{"alias to node alias", []alias{{"baz", "bar"}, {"blob", "baz"}}, "blob.b", nil, false},
		{"leaf alias has priority over node alias", []alias{{"baz", "bar"}, {"baz.b", "a"}}, "baz.b", "a", true},
		{"leaf has priority over alias", []alias{{"a", "foo.a"}}, "a", "a", true},
		{"nested leaf to root leaf", []alias{{"baz.b", "a"}}, "baz.b", "a", true},
		{"nested leaf to self (ignored)", []alias{{"foo.a", "foo.a"}}, "foo.a", "foo.a", true},
		{"nested node to nested node", []alias{{"baz", "bar"}}, "baz.b", "bar.b", true},
		{"nested node to root node", []alias{{"node.a", "a"}}, "node.a", "a", true},
		{"root leaf to nested leaf", []alias{{"c", "foo.b"}}, "c", "foo.b", true},
		{"root node to nested node", []alias{{"", "foo"}}, "b", "foo.b", true},
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			a := config.NewAlias()
			c := config.WithAlias(a)(mr)
			for _, al := range p.aa {
				a.Append(al.new, al.old)
			}
			v, ok := c.Get(p.tp)
			assert.Equal(t, p.ok, ok)
			assert.Equal(t, p.expected, v)
		}
		t.Run(p.name, f)
	}
}

func TestRegexAliasAppend(t *testing.T) {
	r := config.NewRegexAlias()
	require.NotNil(t, r)
	g := mockGetter{
		"a.b.c_d": "d",
		"a.b.c_e": "e",
	}
	v, ok := r.Get(&g, "new")
	assert.False(t, ok)
	assert.Nil(t, v)

	err := r.Append("new", "a.b.c_d")
	assert.Nil(t, err)
	v, ok = r.Get(&g, "new")
	assert.True(t, ok)
	assert.Equal(t, "d", v)

	err = r.Append("new", "a.b.c_e")
	assert.Nil(t, err)
	v, ok = r.Get(&g, "new")
	assert.True(t, ok)
	assert.Equal(t, "d", v)

	delete(g, "a.b.c_d")
	v, ok = r.Get(&g, "new")
	assert.True(t, ok)
	assert.Equal(t, "e", v)

	err = r.Append("(?new", "a.b.c_e")
	assert.NotNil(t, err)
}

func TestRegexAliasGet(t *testing.T) {
	mr := &mockGetter{
		"a":      "a",
		"foo.a":  "foo.a",
		"foo.b":  "foo.b",
		"bar.b":  "bar.b",
		"bar.c":  "bar.c",
		"b[0].a": "b.a",
	}
	type alias struct {
		new string
		old string
	}
	patterns := []struct {
		name     string
		aa       []alias
		tp       string
		expected interface{}
		ok       bool
	}{
		{"array index", []alias{{`(.*)\[\d+\](.*)`, "$1[0]$2"}}, "b[2].a", "b.a", true},
		{"alias to alias", []alias{{"c", "foo.b"}, {"d", "c"}}, "d", nil, false},
		{"alias to node alias", []alias{{"baz", "bar"}, {"blob", "baz"}}, "blob.b", nil, false},
		{"priority order", []alias{{"baz", "bar"}, {"baz.b", "a"}}, "baz.b", "bar.b", true},
		{"leaf has priority over alias", []alias{{"a", "foo.a"}}, "a", "a", true},
		{"nested leaf to root leaf", []alias{{"baz.b", "a"}}, "baz.b", "a", true},
		{"nested leaf to self (ignored)", []alias{{"foo.a", "foo.a"}}, "foo.a", "foo.a", true},
		{"nested node to nested node", []alias{{"baz", "bar"}}, "baz.b", "bar.b", true},
		{"nested node to root node", []alias{{"node.a", "a"}}, "node.a", "a", true},
		{"root leaf to nested leaf", []alias{{"c", "foo.b"}}, "c", "foo.b", true},
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			r := config.NewRegexAlias()
			c := config.WithRegexAlias(r)(mr)
			for _, al := range p.aa {
				err := r.Append(al.new, al.old)
				assert.Nil(t, err, al.new)
			}
			v, ok := c.Get(p.tp)
			assert.Equal(t, p.ok, ok)
			assert.Equal(t, p.expected, v)
		}
		t.Run(p.name, f)
	}
}
