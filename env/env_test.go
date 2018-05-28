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
)

func setup(prefix string) {
	os.Clearenv()
	os.Setenv(prefix+"LEAF", "42")
	os.Setenv(prefix+"SLICE", "a:b")
	os.Setenv(prefix+"NESTED_LEAF", "44")
	os.Setenv(prefix+"NESTED_SLICE", "c:d")
}

func TestNew(t *testing.T) {
	prefix := "CFGENV_"
	setup(prefix)
	e, err := env.New(prefix)
	assert.Nil(t, err)
	require.NotNil(t, e)
	// test provides config.Getter interface.
	cfg := config.New()
	cfg.AppendGetter(e)
}

func TestGetterGet(t *testing.T) {
	patterns := []struct {
		name string
		k    string
		v    interface{}
		ok   bool
	}{
		{"leaf", "leaf", "42", true},
		{"slice", "slice", []string{"a", "b"}, true},
		{"nested", "nested", nil, false},
		{"nested leaf", "nested.leaf", "44", true},
		{"nested slice", "nested.slice", []string{"c", "d"}, true},
		{"nonsense", "nonsense", nil, false},
		{"nested nonsense", "nested.nonsense", nil, false},
	}
	prefix := "CFGENV_"
	setup(prefix)
	e, err := env.New(prefix)
	assert.Nil(t, err)
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

func TestGetterWithCfgKeyReplacer(t *testing.T) {
	prefix := "CFGENV_"
	setup(prefix)
	patterns := []struct {
		name      string
		old       string
		new       string
		treatment keys.Treatment
		expected  string
	}{
		{"default", "_", ".", keys.LowerCase, "nested.leaf"},
		{"null", "_", "_", keys.Unchanged, "NESTED_LEAF"},
		{"lower", "_", "_", keys.LowerCase, "nested_leaf"},
		{"multi old", "TED_", ".", keys.LowerCase, "nes.leaf"},
		{"multi new", "_", "_X_", keys.Unchanged, "NESTED_X_LEAF"},
		{"multi lower", "_", "_X_", keys.LowerCase, "nested_x_leaf"},
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			e, err := env.New(prefix, env.WithCfgKeyReplacer(keys.NewReplacer(p.old, p.new, p.treatment)))
			assert.Nil(t, err)
			require.NotNil(t, e)
			v, ok := e.Get(p.expected)
			assert.True(t, ok)
			assert.Equal(t, "44", v)
		}
		t.Run(p.name, f)
	}
}

func TestGetterSetListSeparator(t *testing.T) {
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
		{"none", "", "a:#b"},
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			e, err := env.New(prefix, env.WithListSeparator(p.sep))
			assert.Nil(t, err)
			require.NotNil(t, e)
			v, ok := e.Get("slice")
			assert.True(t, ok)
			assert.Equal(t, p.expected, v)
		}
		t.Run(p.name, f)
	}
}

func TestGetterSetPrefix(t *testing.T) {
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
			e, err := env.New(prefix, env.WithEnvPrefix(p.prefix))
			assert.Nil(t, err)
			require.NotNil(t, e)
			v, ok := e.Get(p.k)
			assert.True(t, ok)
			assert.Equal(t, "44", v)
		}
		t.Run(p.name, f)
	}
}

func BenchmarkGet(b *testing.B) {
	prefix := "CFGENV_"
	setup(prefix)
	g, _ := env.New(prefix)
	for n := 0; n < b.N; n++ {
		g.Get("nested.leaf")
	}
}
