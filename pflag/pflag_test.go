// Copyright © 2017 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package pflag_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/warthog618/config"
	"github.com/warthog618/config/keys"
	"github.com/warthog618/config/list"
	"github.com/warthog618/config/pflag"
)

func TestGetterAsOption(t *testing.T) {
	c := config.New(pflag.New(), pflag.New())
	c.Close()
}

func TestNew(t *testing.T) {
	f := pflag.New()
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
		name  string
		in    []string
		flags []pflag.Flag
		args  []string
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
			f := pflag.New(
				pflag.WithCommandLine(p.in),
				pflag.WithFlags(p.flags))
			assert.Equal(t, p.args, f.Args())
			assert.Equal(t, len(p.args), f.NArg())
		}
		t.Run(p.name, f)
		// test default to os.Args
		f = func(t *testing.T) {
			oldArgs := os.Args
			os.Args = append([]string{"flagTest"}, p.in...)
			f := pflag.New(pflag.WithFlags(p.flags))
			os.Args = oldArgs
			assert.Equal(t, p.args, f.Args())
			assert.Equal(t, len(p.args), f.NArg())
		}
		t.Run(p.name+"-oa", f)
	}
}

func TestNFlag(t *testing.T) {
	patterns := []struct {
		name  string
		in    []string
		flags []pflag.Flag
		nflag int
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
			[]pflag.Flag{
				{Short: 'n', Name: "nested-leaf"},
				{Short: 'v', Name: "logging-verbosity"},
			}, 3},
		{"malformed shorts", []string{
			"-vn=44",
			"--leaf", "42"},
			[]pflag.Flag{
				{Short: 'n', Name: "nested-leaf"},
				{Short: 'v', Name: "logging-verbosity"},
			}, 1},
		{"short leaf", []string{
			"-l", "42"},
			[]pflag.Flag{
				{Short: 'l', Name: "leaf"},
			}, 1},
		{"short and long leaf", []string{
			"-l", "42",
			"--leaf", "43"},
			[]pflag.Flag{
				{Short: 'l', Name: "leaf"},
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
			[]pflag.Flag{
				{Short: 'a', Name: "angle"},
				{Short: 'b', Name: "bonus"},
			}, 3},
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			f := pflag.New(
				pflag.WithCommandLine(p.in),
				pflag.WithFlags(p.flags))
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
		flags     []pflag.Flag
		expected  []kv
		expectedZ []string
	}{
		{"empty", nil, nil, nil, bogus},
		{"no shorts", []string{"-abc"}, nil, nil, bogus},
		{"a short",
			[]string{"-abc"},
			[]pflag.Flag{
				{Short: 'a', Name: "nested-leaf"},
				{Short: 'v', Name: "logging-verbosity"},
			},
			[]kv{{"nested.leaf", 1}}, bogus},
		{"two shorts",
			[]string{"-a", "-b", "-c"},
			[]pflag.Flag{
				{Short: 'a', Name: "nested-leaf"},
				{Short: 'b', Name: "bonus"},
				{Short: 'v', Name: "logging-verbosity"},
			},
			[]kv{{"nested.leaf", 1}, {"bonus", 1}}, bogus},
		{"grouped shorts",
			[]string{"-abc"},
			[]pflag.Flag{
				{Short: 'a', Name: "nested-leaf"},
				{Short: 'b', Name: "bonus"},
				{Short: 'v', Name: "logging-verbosity"},
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
			[]pflag.Flag{
				{Short: 'n', Name: "nested-leaf"},
				{Short: 'v', Name: "logging-verbosity"},
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
			[]pflag.Flag{
				{Short: 'a', Name: "addon"},
				{Short: 'v', Name: "logging-verbosity"},
			},
			[]kv{{"addon", 1}}, bogus},
		{"precedence2",
			[]string{
				"--addon", "first string",
				"-abc",
				"--addon", "second string",
			},
			[]pflag.Flag{
				{Short: 'a', Name: "addon"},
				{Short: 'v', Name: "logging-verbosity"},
			},
			[]kv{{"addon", "second string"}}, bogus},
		{"non flag after group",
			[]string{
				"--addon", "first string",
				"-ab",
				"stophere",
				"--addon", "second string",
			},
			[]pflag.Flag{
				{Short: 'a', Name: "addon"},
				{Short: 'b', Name: "bonus"},
				{Short: 'v', Name: "logging-verbosity"},
			},
			[]kv{{"addon", 1}, {"bonus", 1}}, bogus},
		{"malformed flag",
			[]string{
				"--addon", "first string",
				"-abc=42",
			},
			[]pflag.Flag{
				{Short: 'a', Name: "addon"},
				{Short: 'b', Name: "bonus"},
				{Short: 'v', Name: "logging-verbosity"},
			},
			[]kv{
				{"addon", "first string"}}, append(bogus, "bonus")},
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			f := pflag.New(
				pflag.WithCommandLine(p.args),
				pflag.WithFlags(p.flags))
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
	flags := []pflag.Flag{{Short: 'n', Name: "nested-leaf"}}
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
			r := pflag.New(
				pflag.WithCommandLine(args),
				pflag.WithFlags(flags),
				pflag.WithKeyReplacer(p.r))
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
	f := pflag.New(pflag.WithCommandLine(args))
	require.NotNil(t, f)
	// basic get
	v, ok := f.Get("config.file")
	assert.True(t, ok)
	assert.Equal(t, "woot", v)
	assert.Implements(t, (*config.Getter)(nil), f)
}

func TestNewWithListSplitter(t *testing.T) {
	args := []string{"-s", "a,#b"}
	flags := []pflag.Flag{{Short: 's', Name: "slice"}}
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
			r := pflag.New(
				pflag.WithCommandLine(args),
				pflag.WithFlags(flags),
				pflag.WithListSplitter(list.NewSplitter(p.sep)))
			require.NotNil(t, r)
			v, ok := r.Get("slice")
			assert.True(t, ok)
			assert.Equal(t, p.expected, v)
		}
		t.Run(p.name, f)
	}
}

func TestNewWithFlags(t *testing.T) {
	patterns := []struct {
		name  string
		args  []string
		flags []pflag.Flag
		key   string
		xval  interface{}
		narg  int
	}{
		{"short bool",
			[]string{"-avbcvv", "-c", "woot"},
			[]pflag.Flag{{Short: 'c', Name: "config-file", Options: pflag.IsBool}},
			"config.file", 2, 1,
		},
		{"short bool val",
			[]string{"-avbcvv", "-c=true", "woot"},
			[]pflag.Flag{{Short: 'c', Name: "config-file", Options: pflag.IsBool}},
			"config.file", "true", 1,
		},
		{"long bool",
			[]string{"--config-file", "woot"},
			[]pflag.Flag{{Name: "config-file", Options: pflag.IsBool}},
			"config.file", 1, 1,
		},
		{"long bool val",
			[]string{"--config-file=false", "woot"},
			[]pflag.Flag{{Name: "config-file", Options: pflag.IsBool}},
			"config.file", "false", 1,
		},
		{"ignore unnamed",
			[]string{"-c", "woot"},
			[]pflag.Flag{
				{Short: 'c', Name: "config-file"},
				{Short: 'c', Options: pflag.IsBool}},
			"config.file", "woot", 0,
		},
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			f := pflag.New(
				pflag.WithCommandLine(p.args),
				pflag.WithFlags(p.flags),
			)
			require.NotNil(t, f)
			v, ok := f.Get(p.key)
			assert.True(t, ok)
			assert.Equal(t, p.xval, v)
			assert.Equal(t, p.narg, f.NArg())
			assert.Implements(t, (*config.Getter)(nil), f)
		}
		t.Run(p.name, f)
	}
}

func BenchmarkNew(b *testing.B) {
	for n := 0; n < b.N; n++ {
		pflag.New(pflag.WithCommandLine([]string{"--leaf", "44"}))
	}
}

func BenchmarkGet(b *testing.B) {
	b.StopTimer()
	g := pflag.New(pflag.WithCommandLine([]string{"--leaf", "44"}))
	b.StartTimer()
	for n := 0; n < b.N; n++ {
		g.Get("leaf")
	}
}

func BenchmarkGetNested(b *testing.B) {
	b.StopTimer()
	g := pflag.New(pflag.WithCommandLine([]string{"--nested-leaf", "44"}))
	b.StartTimer()
	for n := 0; n < b.N; n++ {
		g.Get("nested.leaf")
	}
}

func BenchmarkGetArray(b *testing.B) {
	b.StopTimer()
	g := pflag.New(pflag.WithCommandLine([]string{"--slice", "42,44"}))
	b.StartTimer()
	for n := 0; n < b.N; n++ {
		g.Get("slice")
	}
}

func BenchmarkGetArrayLen(b *testing.B) {
	b.StopTimer()
	g := pflag.New(pflag.WithCommandLine([]string{"--slice", "42,44"}))
	b.StartTimer()
	for n := 0; n < b.N; n++ {
		g.Get("slice[]")
	}
}

func BenchmarkGetArrayElement(b *testing.B) {
	b.StopTimer()
	g := pflag.New(pflag.WithCommandLine([]string{"--slice", "42,44"}))
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
