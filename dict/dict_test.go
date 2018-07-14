// Copyright © 2017 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package dict_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/warthog618/config"
	"github.com/warthog618/config/dict"
)

func TestNew(t *testing.T) {
	d := dict.New()
	require.NotNil(t, d)
	assert.Implements(t, (*config.Getter)(nil), d)
}

func TestGetterGet(t *testing.T) {
	patterns := []struct {
		name string
		k    string
		v    interface{}
		ok   bool
	}{
		{"leaf", "leaf", 42, true},
		{"nested leaf", "nested.leaf", 44, true},
		{"nested nonsense", "nested.nonsense", nil, false},
		{"nested slice", "nested.slice", []interface{}{"c", "d"}, true},
		{"nested", "nested", nil, false},
		{"nonsense", "nonsense", nil, false},
		{"slice", "slice", []string{"a", "b"}, true},
		{"bogusslice[]", "bogusslice[]", 3, true},
		{"slice[]", "slice[]", 2, true},
		{"slice[1]", "slice[1]", "b", true},
		{"slice[4]", "slice[3]", nil, false},
	}
	d := dict.New(dict.WithMap(map[string]interface{}{
		"leaf":         42,
		"bogusslice[]": 3,
		"slice":        []string{"a", "b"},
		"nested": map[string]interface{}{
			"leaf":  44,
			"slice": []interface{}{"c", "d"},
		},
	}))
	for _, p := range patterns {
		f := func(t *testing.T) {
			v, ok := d.Get(p.k)
			assert.Equal(t, p.ok, ok)
			assert.Equal(t, p.v, v)
		}
		t.Run(p.name, f)
	}
}

func TestGetterWithMap(t *testing.T) {
	config := map[string]interface{}{"a": 1}
	g := dict.New(dict.WithMap(config))
	require.NotNil(t, g)
	v, ok := g.Get("a")
	assert.True(t, ok)
	assert.Equal(t, 1, v)
}

func TestGetterSet(t *testing.T) {
	g := dict.New()
	require.NotNil(t, g)
	v, ok := g.Get("a")
	assert.False(t, ok)
	assert.Nil(t, v)
	g.Set("a", 1)
	v, ok = g.Get("a")
	assert.True(t, ok)
	assert.Equal(t, 1, v)
	g.Set("a", 32)
	v, ok = g.Get("a")
	assert.True(t, ok)
	assert.Equal(t, 32, v)
}

func BenchmarkNew(b *testing.B) {
	for n := 0; n < b.N; n++ {
		dict.New(dict.WithMap(map[string]interface{}{"leaf": "44"}))
	}
}

func BenchmarkGet(b *testing.B) {
	b.StopTimer()
	g := dict.New(dict.WithMap(map[string]interface{}{"leaf": "44"}))
	b.StartTimer()
	for n := 0; n < b.N; n++ {
		g.Get("leaf")
	}
}

func BenchmarkGetNested(b *testing.B) {
	b.StopTimer()
	g := dict.New(dict.WithMap(map[string]interface{}{
		"nested": map[string]interface{}{"leaf": "44"}}))
	b.StartTimer()
	for n := 0; n < b.N; n++ {
		g.Get("nested.leaf")
	}
}

func BenchmarkGetArray(b *testing.B) {
	b.StopTimer()
	g := dict.New(dict.WithMap(map[string]interface{}{
		"leaf": []int{1, 2, 3, 4}}))
	b.StartTimer()
	for n := 0; n < b.N; n++ {
		g.Get("nested.leaf")
	}
}

func BenchmarkGetArrayLen(b *testing.B) {
	b.StopTimer()
	g := dict.New(dict.WithMap(map[string]interface{}{
		"leaf": []int{1, 2, 3, 4}}))
	b.StartTimer()
	for n := 0; n < b.N; n++ {
		g.Get("leaf[]")
	}
}

func BenchmarkGetArrayElement(b *testing.B) {
	b.StopTimer()
	g := dict.New(dict.WithMap(map[string]interface{}{
		"leaf": []int{1, 2, 3, 4}}))
	b.StartTimer()
	for n := 0; n < b.N; n++ {
		g.Get("leaf[2]")
	}
}
