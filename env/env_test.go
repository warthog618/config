// Copyright © 2017 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package env_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/warthog618/config"
	"github.com/warthog618/config/env"
	"github.com/warthog618/config/keys"
	"github.com/warthog618/config/list"
)

func setup(prefix string) {
	os.Clearenv()
	os.Setenv(prefix+"LEAF", "42")
	os.Setenv(prefix+"SLICE", "a:b")
	os.Setenv(prefix+"NESTED_LEAF", "44")
	os.Setenv(prefix+"NESTED_SLICE", "c:d")
}

func TestNew(t *testing.T) {
	e := env.New()
	require.NotNil(t, e)
	assert.Implements(t, (*config.Getter)(nil), e)
}

func TestGetterGet(t *testing.T) {
	patterns := []struct {
		name string
		k    string
		v    interface{}
		ok   bool
	}{
		{"leaf", "leaf", "42", true},
		{"nested leaf", "nested.leaf", "44", true},
		{"nested nonsense", "nested.nonsense", nil, false},
		{"nested slice", "nested.slice", []string{"c", "d"}, true},
		{"nested", "nested", nil, false},
		{"nonsense", "nonsense", nil, false},
		{"slice", "slice", []string{"a", "b"}, true},
		{"slice[]", "slice[]", 2, true},
		{"slice[1]", "slice[1]", "b", true},
		{"slice[3]", "slice[3]", nil, false},
	}
	prefix := "CFGENV_"
	setup(prefix)
	e := env.New(env.WithEnvPrefix(prefix))
	require.NotNil(t, e)

	for _, p := range patterns {
		f := func(t *testing.T) {
			v, ok := e.Get(p.k)
			assert.Equal(t, p.ok, ok)
			assert.Equal(t, p.v, v)
		}
		t.Run(p.name, f)
	}
}

func TestNewWithKeyReplacer(t *testing.T) {
	prefix := "CFGENV_"
	setup(prefix)
	patterns := []struct {
		name     string
		r        keys.Replacer
		expected string
	}{
		{"default",
			keys.ChainReplacer(
				keys.StringReplacer("_", "."),
				keys.LowerCaseReplacer()),
			"nested.leaf"},
		{"null", keys.NullReplacer(), "NESTED_LEAF"},
		{"lower", keys.LowerCaseReplacer(), "nested_leaf"},
		{"multi old",
			keys.ChainReplacer(
				keys.StringReplacer("TED_", "."),
				keys.LowerCaseReplacer()),
			"nes.leaf"},
		{"multi new",
			keys.StringReplacer("_", "_X_"),
			"NESTED_X_LEAF"},
		{"multi lower",
			keys.ChainReplacer(
				keys.LowerCaseReplacer(),
				keys.StringReplacer("_", "_X_")),
			"nested_X_leaf"},
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			e := env.New(
				env.WithEnvPrefix(prefix),
				env.WithKeyReplacer(p.r))
			require.NotNil(t, e)
			v, ok := e.Get(p.expected)
			assert.True(t, ok)
			assert.Equal(t, "44", v)
		}
		t.Run(p.name, f)
	}
}

func TestNewWithListSplitter(t *testing.T) {
	prefix := "CFGENV_"
	setup(prefix)
	os.Setenv(prefix+"SLICE", "a:#b")
	patterns := []struct {
		name     string
		sep      string
		expected interface{}
	}{
		{"default", ":", []string{"a", "#b"}},
		{"multi", ":#", []string{"a", "b"}},
		{"none", "", []string{"a", ":", "#", "b"}},
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			e := env.New(
				env.WithEnvPrefix(prefix),
				env.WithListSplitter(list.NewSplitter(p.sep)))
			require.NotNil(t, e)
			v, ok := e.Get("slice")
			assert.True(t, ok)
			assert.Equal(t, p.expected, v)
		}
		t.Run(p.name, f)
	}
}

func TestNewWithEnvPrefix(t *testing.T) {
	prefix := "CFGENV_"
	setup(prefix)
	patterns := []struct {
		name   string
		prefix string
		k      string
	}{
		{"default", "CFGENV_", "nested.leaf"},
		{"multi", "CFG", "env.nested.leaf"},
		{"none", "", "cfgenv.nested.leaf"},
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			e := env.New(env.WithEnvPrefix(p.prefix))
			require.NotNil(t, e)
			v, ok := e.Get(p.k)
			assert.True(t, ok)
			assert.Equal(t, "44", v)
		}
		t.Run(p.name, f)
	}
}

func BenchmarkNew(b *testing.B) {
	b.StopTimer()
	prefix := "CFGENV_"
	setup(prefix)
	b.StartTimer()
	for n := 0; n < b.N; n++ {
		env.New(env.WithEnvPrefix(prefix))
	}
}

func BenchmarkGet(b *testing.B) {
	b.StopTimer()
	prefix := "CFGENV_"
	setup(prefix)
	b.StartTimer()
	g := env.New(env.WithEnvPrefix(prefix))
	for n := 0; n < b.N; n++ {
		g.Get("leaf")
	}
}

func BenchmarkGetNested(b *testing.B) {
	b.StopTimer()
	prefix := "CFGENV_"
	setup(prefix)
	g := env.New(env.WithEnvPrefix(prefix))
	b.StartTimer()
	for n := 0; n < b.N; n++ {
		g.Get("nested.leaf")
	}
}

func BenchmarkGetArray(b *testing.B) {
	b.StopTimer()
	prefix := "CFGENV_"
	setup(prefix)
	g := env.New(env.WithEnvPrefix(prefix))
	b.StartTimer()
	for n := 0; n < b.N; n++ {
		g.Get("slice")
	}
}

func BenchmarkGetArrayLen(b *testing.B) {
	b.StopTimer()
	prefix := "CFGENV_"
	setup(prefix)
	g := env.New(env.WithEnvPrefix(prefix))
	b.StartTimer()
	for n := 0; n < b.N; n++ {
		g.Get("slice[]")
	}
}

func BenchmarkGetArrayElement(b *testing.B) {
	b.StopTimer()
	prefix := "CFGENV_"
	setup(prefix)
	g := env.New(env.WithEnvPrefix(prefix))
	b.StartTimer()
	for n := 0; n < b.N; n++ {
		g.Get("slice[1]")
	}
}

func BenchmarkDefaultReplacer(b *testing.B) {
	b.StopTimer()
	r := keys.ChainReplacer(
		keys.StringReplacer("_", "."),
		keys.LowerCaseReplacer())
	b.StartTimer()
	for n := 0; n < b.N; n++ {
		r.Replace("apple_Banana_Cantelope_date_Eggplant_fig")
	}
}
