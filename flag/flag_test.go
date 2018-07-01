// Copyright © 2017 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package flag_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/warthog618/config"
	"github.com/warthog618/config/flag"
	"github.com/warthog618/config/keys"
)

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

// TestArgs tests the Args and Narg functions.
func TestArgs(t *testing.T) {
	patterns := []struct {
		name   string
		in     []string
		shorts map[byte]string
		args   []string
	}{
		{"nil", nil, nil, nil},
		{"empty", nil, nil, nil},
		{"only one", []string{"arg1"}, nil, []string{"arg1"}},
		{"only two", []string{"arg1", "arg2"}, nil, []string{"arg1", "arg2"}},
		{"two with flags",
			[]string{
				"-v",
				"-n=44",
				"--leaf", "42",
				"--slice=a,b",
				"--nested-slice", "c,d", "arg1", "arg2"},
			nil, []string{"arg1", "arg2"}},
		{"terminated parsing",
			[]string{
				"-v", "-n=44", "--leaf", "42", "--",
				"--slice=a,b", "--nested-slice", "c,d", "arg1"},
			nil,
			[]string{"--slice=a,b", "--nested-slice", "c,d", "arg1"}},
		{"non flag after group",
			[]string{
				"--addon", "first string",
				"-ab",
				"stophere",
				"--addon", "second string"},
			nil, []string{"stophere", "--addon", "second string"}},
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			f, err := flag.New(
				flag.WithCommandLine(p.in),
				flag.WithShortFlags(p.shorts))
			assert.Nil(t, err)
			assert.Equal(t, p.args, f.Args())
			assert.Equal(t, len(p.args), f.NArg())
		}
		t.Run(p.name, f)
		// test default to os.Args
		f = func(t *testing.T) {
			oldArgs := os.Args
			os.Args = append([]string{"flagTest"}, p.in...)
			f, err := flag.New(flag.WithShortFlags(p.shorts))
			os.Args = oldArgs
			assert.Nil(t, err)
			assert.Equal(t, p.args, f.Args())
			assert.Equal(t, len(p.args), f.NArg())
		}
		t.Run(p.name+"-oa", f)
	}
}

func TestNFlag(t *testing.T) {
	patterns := []struct {
		name   string
		in     []string
		shorts map[byte]string
		nflag  int
	}{
		{"empty", []string{}, nil, 0},
		{"only args", []string{"arg1"}, nil, 0},
		{"leaf", []string{
			"-v",    // ignored as no shorts
			"-n=44", // ignored as no shorts
			"--leaf", "42"},
			nil, 1},
		{"slice", []string{
			"-v",    // ignored as no shorts
			"-n=44", // ignored as no shorts
			"--slice=a,b"},
			nil, 1},
		{"multiple", []string{
			"-v",    // ignored as no shorts
			"-n=44", // ignored as no shorts
			"--leaf", "42",
			"--slice=a,b",
			"--nested-slice", "c,d", "arg1"},
			nil, 3},
		{"shorts", []string{
			"-v",
			"-n=44",
			"--leaf", "42"},
			map[byte]string{
				'n': "nested-leaf",
				'v': "logging-verbosity",
			}, 3},
		{"malformed shorts", []string{
			"-vn=44",
			"--leaf", "42"},
			map[byte]string{
				'n': "nested-leaf",
				'v': "logging-verbosity",
			}, 1},
		{"short leaf", []string{
			"-l", "42"},
			map[byte]string{
				'l': "leaf",
			}, 1},
		{"short and long leaf", []string{
			"-l", "42",
			"--leaf", "43"},
			map[byte]string{
				'l': "leaf",
			}, 1},
		{"terminated parsing", []string{
			"-v",    // ignored as no shorts
			"-n=44", // ignored as no shorts
			"--leaf", "42",
			"--",
			"--slice=a,b", "--nested-slice", "c,d", "arg1"},
			nil, 1},
		{"non flag after group",
			[]string{
				"--addon", "first string",
				"-ab",
				"stophere",
				"--addon", "second string"},
			map[byte]string{
				'a': "angle",
				'b': "bonus",
			}, 3},
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			f, err := flag.New(
				flag.WithCommandLine(p.in),
				flag.WithShortFlags(p.shorts))
			assert.Nil(t, err)
			assert.Equal(t, p.nflag, f.NFlag())
		}
		t.Run(p.name, f)
	}
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
		shorts    map[byte]string
		expected  []kv
		expectedZ []string
	}{
		{"empty", nil, nil, nil, bogus},
		{"no shorts", []string{"-abc"}, nil, nil, bogus},
		{"a short",
			[]string{"-abc"},
			map[byte]string{
				'a': "nested-leaf",
				'v': "logging-verbosity",
			},
			[]kv{{"nested.leaf", 1}}, bogus},
		{"two shorts",
			[]string{"-a", "-b", "-c"},
			map[byte]string{
				'a': "nested-leaf",
				'b': "bonus",
				'v': "logging-verbosity",
			},
			[]kv{{"nested.leaf", 1}, {"bonus", 1}}, bogus},
		{"grouped shorts",
			[]string{"-abc"}, map[byte]string{
				'a': "nested-leaf",
				'b': "bonus",
				'v': "logging-verbosity",
			},
			[]kv{{"nested.leaf", 1}, {"bonus", 1}}, bogus},
		{"repeated long",
			[]string{"--bonus", "--bonus", "--bonus"}, nil,
			[]kv{{"bonus", 3}}, bogus},
		{"leaves",
			[]string{
				"-vvv",
				"-n=44",
				"--logging-verbosity",
				"--leaf", "42",
				"--slice=a,b",
				"--nested-slice", "c,d",
			},
			map[byte]string{
				'n': "nested-leaf",
				'v': "logging-verbosity",
			},
			[]kv{
				{"leaf", "42"},
				{"logging.verbosity", 4},
				{"nested.leaf", "44"},
				{"nested.slice", []string{"c", "d"}},
				{"slice", []string{"a", "b"}},
				{"slice[]", 2},
				{"slice[1]", "b"},
			}, bogus,
		},
		{"precedence",
			[]string{
				"--addon", "first string",
				"-abc",
			},
			map[byte]string{
				'a': "addon",
				'v': "logging-verbosity",
			},
			[]kv{{"addon", 1}}, bogus},
		{"precedence2",
			[]string{
				"--addon", "first string",
				"-abc",
				"--addon", "second string",
			},
			map[byte]string{
				'a': "addon",
				'v': "logging-verbosity",
			},
			[]kv{{"addon", "second string"}}, bogus},
		{"non flag after group",
			[]string{
				"--addon", "first string",
				"-ab",
				"stophere",
				"--addon", "second string",
			}, map[byte]string{
				'a': "addon",
				'b': "bonus",
				'v': "logging-verbosity",
			},
			[]kv{{"addon", 1}, {"bonus", 1}}, bogus},
		{"malformed flag",
			[]string{
				"--addon", "first string",
				"-abc=42",
			},
			map[byte]string{
				'a': "addon",
				'b': "bonus",
				'v': "logging-verbosity",
			},
			[]kv{
				{"addon", "first string"}}, append(bogus, "bonus")},
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			f, err := flag.New(
				flag.WithCommandLine(p.args),
				flag.WithShortFlags(p.shorts))
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

func TestNewWithKeyReplacer(t *testing.T) {
	args := []string{"-n=44", "--leaf", "42"}
	shorts := map[byte]string{'n': "nested-leaf"}
	patterns := []struct {
		name     string
		r        flag.Replacer
		expected string
	}{
		{"standard", keys.StringReplacer("-", "_"), "nested_leaf"},
		{"multi old", keys.StringReplacer("ted-", "."), "nes.leaf"},
		{"multi new", keys.StringReplacer("-", "_X_"), "nested_X_leaf"},
		{"no new", keys.StringReplacer("-", ""), "nestedleaf"},
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			r, err := flag.New(
				flag.WithCommandLine(args),
				flag.WithShortFlags(shorts),
				flag.WithKeyReplacer(p.r))
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

func TestNewWithCommandLine(t *testing.T) {
	args := []string{"-avbcvv", "--config-file", "woot"}
	f, err := flag.New(flag.WithCommandLine(args))
	assert.Nil(t, err)
	require.NotNil(t, f)
	// basic get
	v, ok := f.Get("config.file")
	assert.True(t, ok)
	assert.Equal(t, "woot", v)
	assert.Implements(t, (*config.Getter)(nil), f)
}

func TestNewWithListSeparator(t *testing.T) {
	args := []string{"-s", "a,#b"}
	shorts := map[byte]string{'s': "slice"}
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
			r, err := flag.New(
				flag.WithCommandLine(args),
				flag.WithShortFlags(shorts),
				flag.WithListSeparator(p.sep))
			assert.Nil(t, err)
			require.NotNil(t, r)
			v, ok := r.Get("slice")
			assert.True(t, ok)
			assert.Equal(t, p.expected, v)
		}
		t.Run(p.name, f)
	}
}

func TestNewWithShortFlags(t *testing.T) {
	args := []string{"-avbcvv", "-c", "woot"}
	shorts := map[byte]string{'c': "config-file"}
	f, err := flag.New(
		flag.WithCommandLine(args),
		flag.WithShortFlags(shorts),
	)
	assert.Nil(t, err)
	require.NotNil(t, f)
	// basic get
	v, ok := f.Get("config.file")
	assert.True(t, ok)
	assert.Equal(t, "woot", v)
	assert.Implements(t, (*config.Getter)(nil), f)
}

func BenchmarkGet(b *testing.B) {
	g, _ := flag.New(flag.WithCommandLine([]string{"--nested-leaf", "44"}))
	for n := 0; n < b.N; n++ {
		g.Get("nested.leaf")
	}
}

func BenchmarkDefaultReplacer(b *testing.B) {
	r := keys.StringReplacer("-", ".")
	for n := 0; n < b.N; n++ {
		r.Replace("apple-Banana-Cantelope-date-Eggplant-fig")
	}
}
