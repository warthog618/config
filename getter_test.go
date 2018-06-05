// Copyright © 2017 Kent Gibson <warthog618@gmail.com>.
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

func TestNewMappedGetter(t *testing.T) {
	g := mockGetter{map[string]interface{}{}}
	m := keys.NullMapper{}
	p := config.NewMappedGetter(m, &g)
	if p == nil {
		t.Fatalf("new returned nil")
	}
	// test provides config.Getter interface.
	cfg := config.New()
	cfg.AppendGetter(p)
}

func TestMappedGetterGet(t *testing.T) {
	g := mockGetter{
		map[string]interface{}{
			"myapp.a":     "is a",
			"myapp.foo.b": "is foo.b",
			"Myapp.A":     "is a camel",
			"Myapp.Foo.B": "is Foo.B",
		},
	}
	patterns := []struct {
		name string
		m    keys.Mapper
		k    string
		ok   bool
		v    interface{}
	}{
		{"null level 0", keys.NullMapper{}, "a", false, nil},
		{"null level 1", keys.NullMapper{}, "myapp.a", true, "is a"},
		{"null level 2", keys.NullMapper{}, "myapp.foo.b", true, "is foo.b"},
		{"prefix level 1", keys.PrefixMapper{Prefix: "myapp."}, "a", true, "is a"},
		{"prefix level 2", keys.PrefixMapper{Prefix: "myapp."}, "foo.b", true, "is foo.b"},
		{"prefix not level 1", keys.PrefixMapper{Prefix: "myapp."}, "myapp.a", false, nil},
		{"prefix not level 2", keys.PrefixMapper{Prefix: "myapp."}, "myapp.foo.a", false, nil},
		{"prefix empty", keys.PrefixMapper{Prefix: "myapp."}, "", false, nil},
		{"camel level 1", keys.CamelCaseMapper{Sep: "."}, "myapp.A", true, "is a camel"},
		{"camel level 2", keys.CamelCaseMapper{Sep: "."}, "Myapp.Foo.B", true, "is Foo.B"},
		{"camel lower level 1", keys.CamelCaseMapper{Sep: "."}, "myapp.a", true, "is a camel"},
		{"camel lower level 2", keys.CamelCaseMapper{Sep: "."}, "myapp.foo.b", true, "is Foo.B"},
		{"camel empty", keys.CamelCaseMapper{Sep: "."}, "", false, nil},
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			pr := config.NewMappedGetter(p.m, &g)
			v, ok := pr.Get(p.k)
			assert.Equal(t, p.ok, ok)
			assert.Equal(t, p.v, v)
		}
		t.Run(p.name, f)
	}
}

func TestPrefixedGetter(t *testing.T) {
	m := mockGetter{map[string]interface{}{}}
	p := config.NewPrefixedGetter("blah.", &m)
	if p == nil {
		t.Fatalf("new returned nil")
	}
	// test provides config.Getter interface.
	cfg := config.New()
	cfg.AppendGetter(p)
}

func TestPrefixedGetterGet(t *testing.T) {
	m := mockGetter{
		map[string]interface{}{
			"a":     "is a",
			"foo.b": "is foo.b"},
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
	pr := config.NewPrefixedGetter("blah.", &m)
	for _, p := range patterns {
		f := func(t *testing.T) {
			v, ok := pr.Get(p.k)
			assert.Equal(t, p.ok, ok)
			assert.Equal(t, p.v, v)
		}
		t.Run(p.name, f)
	}
}
