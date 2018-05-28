// Copyright © 2017 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package flag_test

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/warthog618/config"
	"github.com/warthog618/config/flag"
)

func TestNew(t *testing.T) {
	args := []string{"-avbcvv", "--config-file", "woot"}
	shorts := map[byte]string{
		'c': "config-file",
		'b': "bonus",
		'v': "logging-verbosity",
	}
	f, err := flag.New(args, shorts)
	assert.Nil(t, err)
	require.NotNil(t, f)
	// test provides config.Getter interface.
	cfg := config.New()
	cfg.AppendGetter(f)
}

// TestArgs tests the Args and Narg functions.
func TestArgs(t *testing.T) {
	patterns := []struct {
		name   string
		in     []string
		shorts map[byte]string
		args   []string
	}{
		{"empty", []string{}, nil, []string{}},
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
			f, err := flag.New(p.in, p.shorts)
			assert.Nil(t, err)
			assert.Equal(t, p.args, f.Args())
			assert.Equal(t, len(p.args), f.NArg())
		}
		t.Run(p.name, f)
		// test default to os.Args
		f = func(t *testing.T) {
			oldArgs := os.Args
			os.Args = append([]string{"flagTest"}, p.in...)
			f, err := flag.New(nil, p.shorts)
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
			f, err := flag.New(p.in, p.shorts)
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
	bogus := []string{"bogus", "nested", "nonsense"}
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
				{"nested.leaf", "44"},
				{"logging.verbosity", 4},
				{"slice", []string{"a", "b"}},
				{"nested.slice", []string{"c", "d"}},
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
			f, err := flag.New(p.args, p.shorts)
			assert.Nil(t, err)
			require.NotNil(t, f)
			assert.Equal(t, len(p.expected), f.NFlag())
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

func TestGetterWithCfgKeyReplacer(t *testing.T) {
	args := []string{"-n=44", "--leaf", "42"}
	shorts := map[byte]string{'n': "nested-leaf"}
	patterns := []struct {
		name     string
		old      string
		new      string
		expected string
	}{
		{"standard", "-", "_", "nested_leaf"},
		{"multi old", "ted-", ".", "nes.leaf"},
		{"multi new", "-", "_X_", "nested_X_leaf"},
		{"no new", "-", "", "nestedleaf"},
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			r, err := flag.New(args, shorts, flag.WithCfgKeyReplacer(strings.NewReplacer(p.old, p.new)))
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

func TestGetterWithListSeparator(t *testing.T) {
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
			r, err := flag.New(args, shorts, flag.WithListSeparator(p.sep))
			assert.Nil(t, err)
			require.NotNil(t, r)
			v, ok := r.Get("slice")
			assert.True(t, ok)
			assert.Equal(t, p.expected, v)
		}
		t.Run(p.name, f)
	}
}
