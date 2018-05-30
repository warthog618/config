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

func TestNewReplacer(t *testing.T) {
	patterns := []struct {
		fromSep   string
		toSep     string
		treatment keys.Treatment
		in        string
		expected  string
	}{
		{"-", ".", keys.Unchanged, "Nested-key", "Nested.key"},
		{"_", ":", keys.LowerCase, "Nested_Key", "nested:key"},
		{".", "_", keys.UpperCase, "Nested.Key", "NESTED_KEY"},
		{"_", "", keys.LowerCamelCase, "NesTed_Key", "nestedKey"},
		{".", "", keys.UpperCamelCase, "nested.key", "NestedKey"},
		{".", "_", 7, "NesteD.Key", "NesteD_Key"},
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			r := keys.NewReplacer(p.fromSep, p.toSep, p.treatment)
			require.NotNil(t, r)
			v := r.Replace(p.in)
			assert.Equal(t, p.expected, v)
		}
		t.Run(p.in, f)
	}
}

func TestNullReplacer(t *testing.T) {
	r := keys.NewNullReplacer()
	patterns := []string{
		"A.b.c_d-e",
		"A.b:c#d-e",
	}
	for _, p := range patterns {
		v := r.Replace(p)
		assert.Equal(t, p, v)
	}
}

func BenchmarkUnchangedReplacer(b *testing.B) {
	r := keys.NewUnchangedReplacer("_", ".")
	for n := 0; n < b.N; n++ {
		r.Replace("apple_Banana_Cantelope_date_Eggplant_fig")
	}
}

func BenchmarkNullReplacer(b *testing.B) {
	r := keys.NewNullReplacer()
	for n := 0; n < b.N; n++ {
		r.Replace("apple_Banana_Cantelope_date_Eggplant_fig")
	}
}

func BenchmarkLowerCaseReplacer(b *testing.B) {
	r := keys.NewLowerCaseReplacer("_", ".")
	for n := 0; n < b.N; n++ {
		r.Replace("apple_Banana_Cantelope_date_Eggplant_fig")
	}
}

func BenchmarkUpperCaseReplacer(b *testing.B) {
	r := keys.NewUpperCaseReplacer("_", ".")
	for n := 0; n < b.N; n++ {
		r.Replace("apple_Banana_Cantelope_date_Eggplant_fig")
	}
}

func BenchmarkLowerCamelCaseReplacer(b *testing.B) {
	r := keys.NewLowerCamelCaseReplacer("_", ".")
	for n := 0; n < b.N; n++ {
		r.Replace("apple_Banana_Cantelope_date_Eggplant_fig")
	}
}

func BenchmarkUpperCamelCaseReplacer(b *testing.B) {
	r := keys.NewUpperCamelCaseReplacer("_", ".")
	for n := 0; n < b.N; n++ {
		r.Replace("apple_Banana_Cantelope_date_Eggplant_fig")
	}
}

func BenchmarkCamelWord(b *testing.B) {
	for n := 0; n < b.N; n++ {
		keys.CamelWord("BaNaNa")
	}
}

func BenchmarkToLower(b *testing.B) {
	for n := 0; n < b.N; n++ {
		strings.ToLower("BaNaNa")
	}
}
