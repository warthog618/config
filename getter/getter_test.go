// Copyright © 2017 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package getter_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/warthog618/config"
	"github.com/warthog618/config/getter"
	"github.com/warthog618/config/keys"
)

func TestNewMapped(t *testing.T) {
	g := mockGetter{map[string]interface{}{}, ""}
	m := keys.NullReplacer()
	p := getter.Decorate(&g, getter.Mapped(m))
	if p == nil {
		t.Fatalf("new returned nil")
	}
	// test provides getter.Getter interface.
	cfg := config.New()
	cfg.AppendGetter(p)
}

func TestMappedGet(t *testing.T) {
	g := mockGetter{
		map[string]interface{}{
			"myapp.a":     "is a",
			"myapp.foo.b": "is foo.b",
			"Myapp.A":     "is a camel",
			"Myapp.Foo.B": "is Foo.B",
		}, "",
	}
	patterns := []struct {
		name string
		r    keys.ReplacerFunc
		k    string
		ok   bool
		v    interface{}
	}{
		{"null level 0", keys.NullReplacer(), "a", false, nil},
		{"null level 1", keys.NullReplacer(), "myapp.a", true, "is a"},
		{"null level 2", keys.NullReplacer(), "myapp.foo.b", true, "is foo.b"},
		{"prefix level 1", keys.PrefixReplacer("myapp."), "a", true, "is a"},
		{"prefix level 2", keys.PrefixReplacer("myapp."), "foo.b", true, "is foo.b"},
		{"prefix not level 1", keys.PrefixReplacer("myapp."), "myapp.a", false, nil},
		{"prefix not level 2", keys.PrefixReplacer("myapp."), "myapp.foo.a", false, nil},
		{"prefix empty", keys.PrefixReplacer("myapp."), "", false, nil},
		{"camel level 1", keys.CamelCaseReplacer(), "myapp.A", true, "is a camel"},
		{"camel level 2", keys.CamelCaseReplacer(), "Myapp.Foo.B", true, "is Foo.B"},
		{"camel lower level 1", keys.CamelCaseReplacer(), "myapp.a", true, "is a camel"},
		{"camel lower level 2", keys.CamelCaseReplacer(), "myapp.foo.b", true, "is Foo.B"},
		{"camel empty", keys.CamelCaseReplacer(), "", false, nil},
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			pr := getter.Decorate(&g, getter.Mapped(p.r))
			v, ok := pr.Get(p.k)
			assert.Equal(t, p.ok, ok)
			assert.Equal(t, p.v, v)
		}
		t.Run(p.name, f)
	}
}

func TestPrefixed(t *testing.T) {
	mg := mockGetter{map[string]interface{}{}, ""}
	p := getter.Decorate(&mg, getter.Prefixed("blah."))
	if p == nil {
		t.Fatalf("new returned nil")
	}
	// test provides getter.Getter interface.
	cfg := config.New()
	cfg.AppendGetter(p)
}

func TestDecorate(t *testing.T) {
	mg := mockGetter{
		map[string]interface{}{
			"a.b.c.d": "is found",
		}, "blah",
	}
	patterns := []struct {
		name string
		k    string
		d    []getter.Decorator
	}{
		{"none", "a.b.c.d", nil},
		{"one", "A.B.C.D", []getter.Decorator{
			getter.Mapped(keys.LowerCaseReplacer()),
		}},
		{"two", "C.D", []getter.Decorator{
			getter.Mapped(keys.LowerCaseReplacer()),
			getter.Mapped(keys.PrefixReplacer("A.B.")),
		}},
		{"two reversed", "C.D", []getter.Decorator{
			getter.Mapped(keys.PrefixReplacer("a.b.")),
			getter.Mapped(keys.LowerCaseReplacer()),
		}},
		{"three", "foo", []getter.Decorator{
			getter.Mapped(keys.PrefixReplacer("a.b.")),
			getter.Mapped(keys.LowerCaseReplacer()),
			getter.Mapped(keys.StringReplacer("foo", "C.D")),
		}},
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			pr := getter.Decorate(&mg, p.d...)
			v, ok := pr.Get(p.k)
			assert.Equal(t, true, ok, mg.lastKey)
			assert.Equal(t, "is found", v)
		}
		t.Run(p.name, f)
	}
}

func TestPrefixedGet(t *testing.T) {
	mg := mockGetter{
		map[string]interface{}{
			"a":     "is a",
			"foo.b": "is foo.b"}, "",
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
	pr := getter.Decorate(&mg, getter.Prefixed("blah."))
	for _, p := range patterns {
		f := func(t *testing.T) {
			v, ok := pr.Get(p.k)
			assert.Equal(t, p.ok, ok)
			assert.Equal(t, p.v, v)
		}
		t.Run(p.name, f)
	}
}

// A simple mock Getter wrapping an accessible map.

type mockGetter struct {
	config  map[string]interface{}
	lastKey string
}

func (mg *mockGetter) Get(key string) (interface{}, bool) {
	mg.lastKey = key
	v, ok := mg.config[key]
	return v, ok
}
