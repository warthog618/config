// Copyright © 2018 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package tree

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGet(t *testing.T) {
	patterns := []struct {
		name string
		n    interface{}
		k    string
		x    interface{}
		ok   bool
	}{
		{"mii", map[interface{}]interface{}{"a": 1}, "a", 1, true},
		{"msi", map[string]interface{}{"a": 1}, "a", 1, true},
		{"mii miss", map[interface{}]interface{}{"a": 1}, "b", nil, false},
		{"msi miss", map[string]interface{}{"a": 1}, "b", nil, false},
		{"neither", map[int]interface{}{1: 1}, "a", nil, false},
	}
	for _, p := range patterns {
		v, ok := Get(p.n, p.k, ".")
		assert.Equal(t, p.ok, ok, p.name)
		assert.Equal(t, p.x, v, p.name)
	}
}

func TestGetFromFunc(t *testing.T) {
	patterns := []struct {
		name string
		k    string
		sep  string
		x    interface{}
		ok   bool
	}{
		{"leaf", "a", "", 1, true},
		{"nested", "nested.b", ".", 2, true},
		{"nested np sep", "nested.b", "", nil, false},
		{"array", "array", "", []uint{1, 2, 3, 4}, true},
		{"array length", "array[]", ".", 4, true},
		{"array element", "array[2]", ".", uint(3), true},
		{"array element", "array[5]", ".", nil, false},
		{"miss", "b", "", nil, false},
		{"overshoot", "nested.b.c", ".", nil, false},
	}
	m := map[string]interface{}{
		"a":      1,
		"nested": map[string]interface{}{"b": 2},
		"array":  []uint{1, 2, 3, 4},
	}
	f := func(k string) (interface{}, bool) {
		v, ok := m[k]
		return v, ok
	}
	for _, p := range patterns {
		v, ok := getFromFunc(f, p.k, p.sep)
		assert.Equal(t, p.ok, ok, p.k)
		assert.Equal(t, p.x, v, p.k)
	}
}

func TestGetArrayElement(t *testing.T) {
	a := []int{1, 2, 3, 4, 5}
	b := []interface{}{
		[]interface{}{1, 2, 3, 4, 5},
		[]int{5, 6, 7, 8}}
	c := []interface{}{
		map[string]interface{}{"a": 1},
		map[interface{}]interface{}{"b": 2},
	}
	d := []interface{}{[]map[string]interface{}{
		{"a": 1},
		{"b": 2},
	}}

	patterns := []struct {
		name string
		n    interface{}
		p    []string
		i    []int
		l    bool
		x    interface{}
		ok   bool
	}{
		{"leaf", a, []string{"a"}, []int{2}, false, 3, true},
		{"nested", b, []string{"a"}, []int{1, 1}, false, 6, true},
		{"length", b, []string{"a"}, []int{1}, true, 4, true},
		{"mii", c, []string{"a", "b"}, []int{1}, false, 2, true},
		{"mii object", c, []string{"a"}, []int{1}, false, nil, false},
		{"msi", c, []string{"a", "a"}, []int{0}, false, 1, true},
		{"msi object", c, []string{"a"}, []int{0}, false, nil, false},
		{"array of object", d, []string{"a"}, []int{0}, false,
			[]interface{}{nil, nil}, true},
		{"path overshoot", c, []string{"a", "b.c"}, []int{1}, false, nil, false},
		{"index overshoot", c, []string{"a"}, []int{1, 1, 1}, false, nil, false},
	}
	for _, p := range patterns {
		v, ok := getArrayElement(p.n, p.p, ".", p.i, p.l)
		assert.Equal(t, p.ok, ok, p.name)
		assert.Equal(t, p.x, v, p.name)
	}
}

func TestGetLeafElement(t *testing.T) {
	patterns := []struct {
		name string
		n    interface{}
		x    interface{}
		ok   bool
	}{
		{"mii", map[interface{}]interface{}{"a": 1}, nil, false},
		{"msi", map[string]interface{}{"a": 1}, nil, false},
		{"array", []interface{}{1, 2, 3}, []interface{}{1, 2, 3}, true},
		{"empty array", []interface{}{}, []interface{}{}, true},
		{"array of msi",
			[]interface{}{map[string]interface{}{"a": 1}},
			[]interface{}{nil}, true},
		{"array of mii",
			[]interface{}{map[interface{}]interface{}{"a": 1}},
			[]interface{}{nil}, true},
		{"int", 3, 3, true},
	}
	for _, p := range patterns {
		v, ok := getLeafElement(p.n)
		assert.Equal(t, p.ok, ok, p.name)
		assert.Equal(t, p.x, v, p.name)
	}
}

func BenchmarkGet(b *testing.B) {
	g := map[string]interface{}{"leaf": "44"}
	for n := 0; n < b.N; n++ {
		Get(g, "leaf", ".")
	}
}

func BenchmarkGetNested(b *testing.B) {
	g := map[string]interface{}{"nested": map[string]interface{}{"leaf": "44"}}
	for n := 0; n < b.N; n++ {
		Get(g, "nested.leaf", ".")
	}
}

func BenchmarkGetArray(b *testing.B) {
	g := map[string]interface{}{"leaf": []int{1, 2, 3, 4}}
	for n := 0; n < b.N; n++ {
		Get(g, "leaf", ".")
	}
}

func BenchmarkGetArrayLen(b *testing.B) {
	g := map[string]interface{}{"leaf": []int{1, 2, 3, 4}}
	for n := 0; n < b.N; n++ {
		Get(g, "nested.leaf[]", ".")
	}
}

func BenchmarkGetArrayElement(b *testing.B) {
	g := map[string]interface{}{"leaf": []int{1, 2, 3, 4}}
	for n := 0; n < b.N; n++ {
		Get(g, "nested.leaf[2]", ".")
	}
}
