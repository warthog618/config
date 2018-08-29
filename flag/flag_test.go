// Copyright © 2017 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package flag_test

import (
	goflag "flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/warthog618/config"
	"github.com/warthog618/config/flag"
	"github.com/warthog618/config/keys"
	"github.com/warthog618/config/list"
)

func init() {
	goflag.Bool("logging-verbose", false, "")
	goflag.Int("leaf", 2, "")
	goflag.Int("nested-leaf", 4, "")
	goflag.String("slice", "", "")
	goflag.String("nested-slice", "", "")
}
func TestNew(t *testing.T) {
	f, err := flag.New()
	assert.Nil(t, err)
	require.NotNil(t, f)
	// basic get
	v, ok := f.Get("config.file")
	assert.False(t, ok)
	assert.Nil(t, v)
	assert.Implements(t, (*config.Getter)(nil), f)
}

func TestGetterGet(t *testing.T) {
	type kv struct {
		k string
		v interface{}
	}
	bogus := []string{"bogus", "nested", "nonsense", "slice[3]"}
	patterns := []struct {
		name      string
		args      []string
		expected  []kv
		expectedZ []string
	}{
		{"empty", nil, nil, bogus},
		{"no flags", []string{"abc"}, nil, bogus},
		{"leaves",
			[]string{
				"--nested-leaf=44",
				"--logging-verbose",
				"--leaf", "42",
				"--slice=a,b",
				"--nested-slice", "c,d",
			},
			[]kv{
				{"leaf", "42"},
				{"logging.verbose", "true"},
				{"nested.leaf", "44"},
				{"nested.slice", []string{"c", "d"}},
				{"slice", []string{"a", "b"}},
				{"slice[]", 2},
				{"slice[1]", "b"},
			}, bogus,
		},
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			oldArgs := os.Args
			os.Args = append([]string{"flagTest"}, p.args...)
			goflag.Parse()
			f, err := flag.New()
			os.Args = oldArgs
			assert.Nil(t, err)
			require.NotNil(t, f)
			for _, x := range p.expected {
				v, ok := f.Get(x.k)
				assert.True(t, ok, x.k)
				assert.Equal(t, x.v, v, x.k)
			}
			for _, x := range p.expectedZ {
				v, ok := f.Get(x)
				assert.False(t, ok, x)
				assert.Nil(t, v, x)
			}
		}
		t.Run(p.name, f)
	}
}

func TestNewWithAllFlags(t *testing.T) {
	args := []string{"--nested-leaf=44", "--leaf", "42"}
	patterns := []struct {
		name     string
		key      string
		expected string
	}{
		{"leaf", "leaf", "42"},
		{"nested", "nested.leaf", "44"},
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			oldArgs := os.Args
			os.Args = append([]string{"flagTest"}, args...)
			goflag.Parse()
			r, err := flag.New(flag.WithAllFlags())
			os.Args = oldArgs
			assert.Nil(t, err)
			require.NotNil(t, r)
			v, ok := r.Get(p.key)
			assert.True(t, ok)
			assert.Equal(t, p.expected, v)
		}
		t.Run(p.name, f)
	}
}

func TestNewWithKeyReplacer(t *testing.T) {
	args := []string{"--nested-leaf=44", "--leaf", "42"}
	patterns := []struct {
		name     string
		r        keys.Replacer
		expected string
	}{
		{"standard", keys.StringReplacer("-", "_"), "nested_leaf"},
		{"multi old", keys.StringReplacer("ted-", "."), "nes.leaf"},
		{"multi new", keys.StringReplacer("-", "_X_"), "nested_X_leaf"},
		{"no new", keys.StringReplacer("-", ""), "nestedleaf"},
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			oldArgs := os.Args
			os.Args = append([]string{"flagTest"}, args...)
			goflag.Parse()
			r, err := flag.New(flag.WithKeyReplacer(p.r))
			os.Args = oldArgs
			assert.Nil(t, err)
			require.NotNil(t, r)
			v, ok := r.Get(p.expected)
			assert.True(t, ok)
			assert.Equal(t, "44", v)
			v, ok = r.Get("leaf")
			assert.True(t, ok)
			assert.Equal(t, "42", v)
		}
		t.Run(p.name, f)
	}
}

func TestNewWithListSplitter(t *testing.T) {
	args := []string{"--slice", "a,#b"}
	patterns := []struct {
		name     string
		sep      string
		expected interface{}
	}{
		{"single", ",", []string{"a", "#b"}},
		{"multi", ",#", []string{"a", "b"}},
		{"none", "-", "a,#b"},
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			oldArgs := os.Args
			os.Args = append([]string{"flagTest"}, args...)
			goflag.Parse()
			r, err := flag.New(flag.WithListSplitter(list.NewSplitter(p.sep)))
			os.Args = oldArgs
			assert.Nil(t, err)
			require.NotNil(t, r)
			v, ok := r.Get("slice")
			assert.True(t, ok)
			assert.Equal(t, p.expected, v)
		}
		t.Run(p.name, f)
	}
}

func BenchmarkNew(b *testing.B) {
	b.StopTimer()
	oldArgs := os.Args
	os.Args = append([]string{"flagTest"}, "--leaf", "44")
	goflag.Parse()
	os.Args = oldArgs
	b.StartTimer()
	for n := 0; n < b.N; n++ {
		flag.New()
	}
}

func BenchmarkGet(b *testing.B) {
	b.StopTimer()
	oldArgs := os.Args
	os.Args = append([]string{"flagTest"}, "--leaf", "44")
	goflag.Parse()
	g, _ := flag.New()
	os.Args = oldArgs
	b.StartTimer()
	for n := 0; n < b.N; n++ {
		g.Get("leaf")
	}
}

func BenchmarkGetNested(b *testing.B) {
	b.StopTimer()
	oldArgs := os.Args
	os.Args = append([]string{"flagTest"}, "--nested-leaf", "44")
	goflag.Parse()
	g, _ := flag.New()
	os.Args = oldArgs
	b.StartTimer()
	for n := 0; n < b.N; n++ {
		g.Get("nested.leaf")
	}
}

func BenchmarkGetArray(b *testing.B) {
	b.StopTimer()
	oldArgs := os.Args
	os.Args = append([]string{"flagTest"}, "--slice", "42,44")
	goflag.Parse()
	g, _ := flag.New()
	os.Args = oldArgs
	b.StartTimer()
	for n := 0; n < b.N; n++ {
		g.Get("slice")
	}
}

func BenchmarkGetArrayLen(b *testing.B) {
	b.StopTimer()
	oldArgs := os.Args
	os.Args = append([]string{"flagTest"}, "--slice", "42,44")
	goflag.Parse()
	g, _ := flag.New()
	os.Args = oldArgs
	b.StartTimer()
	for n := 0; n < b.N; n++ {
		g.Get("slice[]")
	}
}

func BenchmarkGetArrayElement(b *testing.B) {
	b.StopTimer()
	oldArgs := os.Args
	os.Args = append([]string{"flagTest"}, "--slice", "42,44")
	goflag.Parse()
	g, _ := flag.New()
	os.Args = oldArgs
	b.StartTimer()
	for n := 0; n < b.N; n++ {
		g.Get("slice[1]")
	}
}

func BenchmarkDefaultReplacer(b *testing.B) {
	r := keys.StringReplacer("-", ".")
	for n := 0; n < b.N; n++ {
		r.Replace("apple-Banana-Cantelope-date-Eggplant-fig")
	}
}
