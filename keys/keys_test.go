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

func TestCamelCaseMapper(t *testing.T) {
	patterns := []struct {
		sep      string
		in       string
		expected string
	}{
		{"-", "topKey", "Topkey"},
		{"", "topKEy", "TOPKEY"}, // splits on every character
		{"-", "nested-key", "Nested-Key"},
		{".", "Nested.key", "Nested.Key"},
		{"_", "Nested_Key", "Nested_Key"},
		{"_", "nested_Key", "Nested_Key"},
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			m := keys.CamelCaseMapper{Sep: p.sep}
			require.NotNil(t, m)
			v := m.Map(p.in)
			assert.Equal(t, p.expected, v)
		}
		t.Run(p.in, f)
	}
}

func TestLowerCamelCaseMapper(t *testing.T) {
	patterns := []struct {
		sep      string
		in       string
		expected string
	}{
		{"-", "topKey", "topkey"},
		{"", "topKEy", "tOPKEY"}, // splits on every character
		{"-", "nested-key", "nested-Key"},
		{".", "Nested.key", "nested.Key"},
		{"_", "Nested_Key", "nested_Key"},
		{"_", "nested_Key", "nested_Key"},
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			m := keys.LowerCamelCaseMapper{Sep: p.sep}
			require.NotNil(t, m)
			v := m.Map(p.in)
			assert.Equal(t, p.expected, v)
		}
		t.Run(p.in, f)
	}
}

func TestLowerCaseMapper(t *testing.T) {
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
			m := keys.LowerCaseMapper{}
			require.NotNil(t, m)
			v := m.Map(p.in)
			assert.Equal(t, p.expected, v)
		}
		t.Run(p.in, f)
	}
}

type mockMapper struct {
	name string
}

func (m mockMapper) Map(key string) string {
	return m.name + key
}

func TestMultiMapper(t *testing.T) {
	m1 := mockMapper{"m1"}
	m2 := mockMapper{"m2"}
	m3 := mockMapper{"m3"}
	patterns := []struct {
		name     string
		in       []keys.Mapper
		expected string
	}{
		{"none", []keys.Mapper{}, "banana"},
		{"one", []keys.Mapper{m1}, "m1banana"},
		{"two", []keys.Mapper{m1, m2}, "m2m1banana"},
		{"three", []keys.Mapper{m1, m2, m3}, "m3m2m1banana"},
		{"revtwo", []keys.Mapper{m2, m1}, "m1m2banana"},
		{"revthree", []keys.Mapper{m3, m2, m1}, "m1m2m3banana"},
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			m := keys.MultiMapper{MM: p.in}
			require.NotNil(t, m)
			v := m.Map("banana")
			assert.Equal(t, p.expected, v)
		}
		t.Run(p.name, f)
	}
}

func TestNullMapper(t *testing.T) {
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
			m := keys.NullMapper{}
			require.NotNil(t, m)
			v := m.Map(p.in)
			assert.Equal(t, p.expected, v)
		}
		t.Run(p.in, f)
	}
}

func TestPrefixMapper(t *testing.T) {
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
			m := keys.PrefixMapper{Prefix: p.prefix}
			require.NotNil(t, m)
			v := m.Map(p.in)
			assert.Equal(t, p.expected, v)
		}
		t.Run(p.prefix+p.in, f)
	}
}

func TestSeparatorMapper(t *testing.T) {
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
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			m := keys.ReplaceMapper{From: p.from, To: p.to}
			require.NotNil(t, m)
			v := m.Map(p.in)
			assert.Equal(t, p.expected, v)
		}
		t.Run(p.in, f)
	}
}

func TestUpperCaseMapper(t *testing.T) {
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
			m := keys.UpperCaseMapper{}
			require.NotNil(t, m)
			v := m.Map(p.in)
			assert.Equal(t, p.expected, v)
		}
		t.Run(p.in, f)
	}
}

func TestCamelCase(t *testing.T) {
	patterns := []struct {
		in       string
		expected string
	}{
		{"topKey", "Topkey"},
		{"topKEy", "Topkey"},
		{"TOPKEY", "Topkey"},
		{"nested-key", "Nested-key"},
		{"Nested-key", "Nested-key"},
		{"spaced key", "Spaced key"},
		{"Spaced key", "Spaced key"},
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			v := keys.CamelCase(p.in)
			assert.Equal(t, p.expected, v)
		}
		t.Run(p.in, f)
	}
}

func BenchmarkCamelCase(b *testing.B) {
	for n := 0; n < b.N; n++ {
		keys.CamelCase("BaNaNa")
	}
}

func BenchmarkToLower(b *testing.B) {
	for n := 0; n < b.N; n++ {
		strings.ToLower("BaNaNa")
	}
}

func BenchmarkCamelCaseMapper(b *testing.B) {
	m := keys.CamelCaseMapper{Sep: "_"}
	for n := 0; n < b.N; n++ {
		m.Map("apple_Banana_Cantelope_date_Eggplant_fig")
	}
}

func BenchmarkNullMapper(b *testing.B) {
	m := keys.NullMapper{}
	for n := 0; n < b.N; n++ {
		m.Map("apple_Banana_Cantelope_date_Eggplant_fig")
	}
}

func BenchmarkLowerCamelCaseMapper(b *testing.B) {
	m := keys.LowerCamelCaseMapper{Sep: "_"}
	for n := 0; n < b.N; n++ {
		m.Map("apple_Banana_Cantelope_date_Eggplant_fig")
	}
}

func BenchmarkLowerCaseMapper(b *testing.B) {
	m := keys.LowerCaseMapper{}
	for n := 0; n < b.N; n++ {
		m.Map("apple_Banana_Cantelope_date_Eggplant_fig")
	}
}

func BenchmarkPrefixMapper(b *testing.B) {
	m := keys.PrefixMapper{Prefix: "apple_"}
	for n := 0; n < b.N; n++ {
		m.Map("apple_Banana_Cantelope_date_Eggplant_fig")
	}
}
func BenchmarkSeparatorMapper(b *testing.B) {
	m := keys.ReplaceMapper{From: "_", To: "."}
	for n := 0; n < b.N; n++ {
		m.Map("apple_Banana_Cantelope_date_Eggplant_fig")
	}
}

func BenchmarkUpperCaseMapper(b *testing.B) {
	m := keys.UpperCaseMapper{}
	for n := 0; n < b.N; n++ {
		m.Map("apple_Banana_Cantelope_date_Eggplant_fig")
	}
}
