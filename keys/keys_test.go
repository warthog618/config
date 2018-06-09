// Copyright © 2017 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package keys_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/warthog618/config/keys"
)

func TestCamelCaseReplacer(t *testing.T) {
	patterns := []struct {
		in       string
		expected string
	}{
		{"", ""},
		{"topKey", "Topkey"},
		{"nested.key", "Nested.Key"},
		{"Nested.key", "Nested.Key"},
		{"Nested.Key", "Nested.Key"},
		{"nested.Key", "Nested.Key"},
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			m := keys.CamelCaseReplacer()
			require.NotNil(t, m)
			v := m.Replace(p.in)
			assert.Equal(t, p.expected, v)
		}
		t.Run(p.in, f)
	}
}

func TestCamelCaseSepReplacer(t *testing.T) {
	patterns := []struct {
		sep      string
		in       string
		expected string
	}{
		{"", "", ""},
		{"-", "", ""},
		{"-", "topKey", "Topkey"},
		{"", "topKEy", "TOPKEY"}, // splits on every character
		{"-", "nested-key", "Nested-Key"},
		{".", "Nested.key", "Nested.Key"},
		{"_", "Nested_Key", "Nested_Key"},
		{"_", "nested_Key", "Nested_Key"},
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			m := keys.CamelCaseSepReplacer(p.sep)
			require.NotNil(t, m)
			v := m.Replace(p.in)
			assert.Equal(t, p.expected, v)
		}
		t.Run(p.in, f)
	}
}

func TestChainReplacer(t *testing.T) {
	patterns := []struct {
		name     string
		in       string
		expected string
		r        []keys.ReplacerFunc
	}{
		{"empty", "", "", nil},
		{"nil", "", "", []keys.ReplacerFunc{nil}},
		{"none", "a.b.c.d", "a.b.c.d", nil},
		{"one", "A.B.C.D", "a.b.c.d",
			[]keys.ReplacerFunc{
				keys.LowerCaseReplacer(),
			}},
		{"two", "C.D", "a.b.c.d",
			[]keys.ReplacerFunc{
				keys.LowerCaseReplacer(),
				keys.PrefixReplacer("a.b."),
			}},
		{"three", "C.D", "a.b.c.d",
			[]keys.ReplacerFunc{
				keys.LowerCaseReplacer(),
				keys.PrefixReplacer("foo."),
				keys.StringReplacer("foo", "a.b"),
			}},
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			m := keys.ChainReplacer(p.r...)
			require.NotNil(t, m)
			v := m.Replace(p.in)
			assert.Equal(t, p.expected, v)
		}
		t.Run(p.name, f)
	}
}

func TestLowerCamelCaseReplacer(t *testing.T) {
	patterns := []struct {
		in       string
		expected string
	}{
		{"", ""},
		{"topKey", "topkey"},
		{"nested.key", "nested.Key"},
		{"Nested.key", "nested.Key"},
		{"Nested.Key", "nested.Key"},
		{"nested.Key", "nested.Key"},
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			m := keys.LowerCamelCaseReplacer()
			require.NotNil(t, m)
			v := m.Replace(p.in)
			assert.Equal(t, p.expected, v)
		}
		t.Run(p.in, f)
	}
}

func TestLowerCamelCaseSepReplacer(t *testing.T) {
	patterns := []struct {
		sep      string
		in       string
		expected string
	}{
		{"", "", ""},
		{"-", "", ""},
		{"-", "topKey", "topkey"},
		{"", "topKEy", "tOPKEY"}, // splits on every character
		{"-", "nested-key", "nested-Key"},
		{".", "Nested.key", "nested.Key"},
		{"_", "Nested_Key", "nested_Key"},
		{"_", "nested_Key", "nested_Key"},
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			m := keys.LowerCamelCaseSepReplacer(p.sep)
			require.NotNil(t, m)
			v := m.Replace(p.in)
			assert.Equal(t, p.expected, v)
		}
		t.Run(p.in, f)
	}
}

func TestLowerCaseReplacer(t *testing.T) {
	patterns := []struct {
		in       string
		expected string
	}{
		{"topKey", "topkey"},
		{"topKEy", "topkey"},
		{"nested-key", "nested-key"},
		{"Nested-key", "nested-key"},
		{"Nested_Key", "nested_key"},
		{"nested_Key", "nested_key"},
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			m := keys.LowerCaseReplacer()
			require.NotNil(t, m)
			v := m.Replace(p.in)
			assert.Equal(t, p.expected, v)
		}
		t.Run(p.in, f)
	}
}

func TestNullReplacer(t *testing.T) {
	patterns := []struct {
		in       string
		expected string
	}{
		{"topKey", "topKey"},
		{"topKEy", "topKEy"},
		{"nested-key", "nested-key"},
		{"Nested-key", "Nested-key"},
		{"Nested_Key", "Nested_Key"},
		{"nested_Key", "nested_Key"},
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			m := keys.NullReplacer()
			require.NotNil(t, m)
			v := m.Replace(p.in)
			assert.Equal(t, p.expected, v)
		}
		t.Run(p.in, f)
	}
}

func TestPrefixReplacer(t *testing.T) {
	patterns := []struct {
		prefix   string
		in       string
		expected string
	}{
		{"", "topKey", "topKey"},
		{"pre", "topKEy", "pretopKEy"},
		{"pre-", "nested-key", "pre-nested-key"},
		{"pre-", "Nested-key", "pre-Nested-key"},
		{"Pre-", "Nested_Key", "Pre-Nested_Key"},
		{"PRE_", "nested_Key", "PRE_nested_Key"},
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			m := keys.PrefixReplacer(p.prefix)
			require.NotNil(t, m)
			v := m.Replace(p.in)
			assert.Equal(t, p.expected, v)
		}
		t.Run(p.prefix+p.in, f)
	}
}

func TestStringReplacer(t *testing.T) {
	patterns := []struct {
		from     string
		to       string
		in       string
		expected string
	}{
		{"", "-", "topKey", "-t-o-p-K-e-y-"},
		{"-", ".", "topKEy", "topKEy"},
		{"-", ".", "nested-key", "nested.key"},
		{"_", "-", "Nested_key", "Nested-key"},
		{".", "_", "Nested.Key", "Nested_Key"},
		{":", "#", "nested:Key", "nested#Key"},
		{"foo", "a.b", "foo.c.d", "a.b.c.d"},
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			r := keys.StringReplacer(p.from, p.to)
			require.NotNil(t, r)
			v := r.Replace(p.in)
			assert.Equal(t, p.expected, v)
		}
		t.Run(p.in, f)
	}
}

func TestUpperCaseReplacer(t *testing.T) {
	patterns := []struct {
		in       string
		expected string
	}{
		{"topKey", "TOPKEY"},
		{"topKEy", "TOPKEY"},
		{"nested-key", "NESTED-KEY"},
		{"Nested-key", "NESTED-KEY"},
		{"Nested_Key", "NESTED_KEY"},
		{"nested_Key", "NESTED_KEY"},
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			r := keys.UpperCaseReplacer()
			require.NotNil(t, r)
			v := r.Replace(p.in)
			assert.Equal(t, p.expected, v)
		}
		t.Run(p.in, f)
	}
}

func BenchmarkToLower(b *testing.B) {
	for n := 0; n < b.N; n++ {
		strings.ToLower("BaNaNa")
	}
}

func BenchmarkCamelCase(b *testing.B) {
	r := keys.CamelCaseSepReplacer("_")
	for n := 0; n < b.N; n++ {
		r.Replace("apple_Banana_Cantelope_date_Eggplant_fig")
	}
}

func BenchmarkNull(b *testing.B) {
	r := keys.NullReplacer()
	for n := 0; n < b.N; n++ {
		r.Replace("apple_Banana_Cantelope_date_Eggplant_fig")
	}
}

func BenchmarkLowerCamelCase(b *testing.B) {
	r := keys.LowerCamelCaseSepReplacer("_")
	for n := 0; n < b.N; n++ {
		r.Replace("apple_Banana_Cantelope_date_Eggplant_fig")
	}
}

func BenchmarkLowerCase(b *testing.B) {
	r := keys.LowerCaseReplacer()
	for n := 0; n < b.N; n++ {
		r.Replace("apple_Banana_Cantelope_date_Eggplant_fig")
	}
}

func BenchmarkPrefix(b *testing.B) {
	r := keys.PrefixReplacer("apple_")
	for n := 0; n < b.N; n++ {
		r.Replace("apple_Banana_Cantelope_date_Eggplant_fig")
	}
}
func BenchmarkStringReplace(b *testing.B) {
	r := keys.StringReplacer("_", ".")
	for n := 0; n < b.N; n++ {
		r.Replace("apple_Banana_Cantelope_date_Eggplant_fig")
	}
}

func BenchmarkUpperCase(b *testing.B) {
	r := keys.UpperCaseReplacer()
	for n := 0; n < b.N; n++ {
		r.Replace("apple_Banana_Cantelope_date_Eggplant_fig")
	}
}

type mockReplacer struct {
	name string
}

func (m mockReplacer) Replace(key string) string {
	return m.name + key
}
