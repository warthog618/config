// Copyright © 2018 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/warthog618/config"
	"github.com/warthog618/config/keys"
)

func TestDecorate(t *testing.T) {
	eg := echoGetter{}
	patterns := []struct {
		name     string
		k        string
		d        []config.Decorator
		expected string
	}{
		{"none", "a.b.c.d", nil, "a.b.c.d"},
		{"one", "A.B.C.D",
			[]config.Decorator{
				config.WithKeyReplacer(keys.LowerCaseReplacer()),
			},
			"a.b.c.d",
		},
		{"two", "C.D",
			[]config.Decorator{
				config.WithKeyReplacer(keys.LowerCaseReplacer()),
				config.WithPrefix("A.B."),
			},
			"A.B.c.d",
		},
		{"two reversed", "C.D",
			[]config.Decorator{
				config.WithPrefix("a.b."),
				config.WithKeyReplacer(keys.LowerCaseReplacer()),
			},
			"a.b.c.d",
		},
		{"three", "foo",
			[]config.Decorator{
				config.WithPrefix("a.B."),
				config.WithKeyReplacer(keys.LowerCaseReplacer()),
				config.WithKeyReplacer(keys.StringReplacer("foo", "C.D")),
			},
			"a.b.C.D",
		},
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			pr := config.Decorate(&eg, p.d...)
			v, ok := pr.Get(p.k)
			assert.Equal(t, true, ok, p.k)
			assert.Equal(t, p.expected, v)
		}
		t.Run(p.name, f)
	}
}

func TestOverlay(t *testing.T) {
	under := mockGetter{
		"a.b.c": 43,
		"a.b.d": 41,
	}
	over := mockGetter{
		"a.b.d": 42,
	}
	g := config.Overlay(over, under)

	// under
	c, ok := g.Get("a.b.c")
	assert.True(t, ok)
	assert.Equal(t, 43, c)

	// shadowed by over
	c, ok = g.Get("a.b.d")
	assert.True(t, ok)
	assert.Equal(t, 42, c)

	// neither
	c, ok = g.Get("a.b.e")
	assert.False(t, ok)
	assert.Nil(t, c)
}

func TestWithDefault(t *testing.T) {
	def := mockGetter{
		"a.b.c": 43,
		"a.b.d": 41,
	}
	nondef := mockGetter{
		"a.b.d": 42,
	}
	g := config.WithDefault(&def)(nondef)

	// defaulted
	c, ok := g.Get("a.b.c")
	assert.True(t, ok)
	assert.Equal(t, 43, c)

	// non-default
	c, ok = g.Get("a.b.d")
	assert.True(t, ok)
	assert.Equal(t, 42, c)

	// neither
	c, ok = g.Get("a.b.e")
	assert.False(t, ok)
	assert.Nil(t, c)

	// nil default
	g = config.WithDefault(nil)(nondef)

	// no longer defaulted
	c, ok = g.Get("a.b.c")
	assert.False(t, ok)
	assert.Nil(t, c)

	// non-default
	c, ok = g.Get("a.b.d")
	assert.True(t, ok)
	assert.Equal(t, 42, c)
}

func TestWithGraft(t *testing.T) {
	mg := mockGetter{
		"a":     "is a",
		"foo.b": "is foo.b",
	}
	patterns := []struct {
		name string
		k    string
		ok   bool
		v    interface{}
	}{
		{"level 1", "blah.a", true, "is a"},
		{"level 2", "blah.foo.b", true, "is foo.b"},
		{"not level 1", "notblah.a", false, nil},
		{"not level 2", "notblah.foo.a", false, nil},
		{"empty", "", false, nil},
		{"level 0", "a", false, nil},
	}
	pr := config.WithGraft("blah.")(&mg)
	for _, p := range patterns {
		f := func(t *testing.T) {
			v, ok := pr.Get(p.k)
			assert.Equal(t, p.ok, ok)
			assert.Equal(t, p.v, v)
		}
		t.Run(p.name, f)
	}
}
func TestWithKeyReplacer(t *testing.T) {
	patterns := []struct {
		name     string
		k        string
		d        config.Replacer
		expected string
	}{
		{"nil", "a.b.c.d", nil, "a.b.c.d"},
		{"lower", "A.B.C.D", keys.LowerCaseReplacer(), "a.b.c.d"},
		{"upper", "a.b.C.d", keys.UpperCaseReplacer(), "A.B.C.D"},
		{"string", "a.b.foo", keys.StringReplacer("foo", "C.D"), "a.b.C.D"},
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			pr := config.WithKeyReplacer(p.d)(echoGetter{})
			v, ok := pr.Get(p.k)
			assert.Equal(t, true, ok, p.k)
			assert.Equal(t, p.expected, v)
		}
		t.Run(p.name, f)
	}
}

func TestWithPrefix(t *testing.T) {
	patterns := []struct {
		name   string
		k      string
		prefix string
	}{
		{"none", "key", ""},
		{"one", "b.c", "a"},
		{"two", "c.d", "a.b"},
		{"three", "d", "a.b.c"},
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			pr := config.WithPrefix(p.prefix)(echoGetter{})
			v, ok := pr.Get(p.k)
			assert.Equal(t, true, ok, p.k)
			assert.Equal(t, p.prefix+p.k, v)
		}
		t.Run(p.name, f)
	}
}

type echoGetter struct{}

func (e echoGetter) Get(key string) (interface{}, bool) {
	return key, true
}
