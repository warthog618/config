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

func TestGetFromMII(t *testing.T) {
	m := map[interface{}]interface{}{
		"a":      1,
		"nested": map[interface{}]interface{}{"b": 2},
		"array":  []interface{}{1, 2, 3, 4},
	}
	patterns := []struct {
		name string
		n    map[interface{}]interface{}
		k    string
		x    interface{}
		ok   bool
	}{
		{"leaf", m, "a", 1, true},
		{"nested", m, "nested.b", 2, true},
		{"array", m, "array", []interface{}{1, 2, 3, 4}, true},
		{"array length", m, "array[]", 4, true},
		{"array element", m, "array[2]", 3, true},
		{"array element", m, "array[5]", nil, false},
		{"miss", m, "b", nil, false},
	}
	for _, p := range patterns {
		v, ok := getFromMII(p.n, p.k, ".")
		assert.Equal(t, p.ok, ok, p.k)
		assert.Equal(t, p.x, v, p.k)
	}
}

func TestGetFromMSI(t *testing.T) {
	m := map[string]interface{}{
		"a":      1,
		"nested": map[string]interface{}{"b": 2},
		"array":  []interface{}{1, 2, 3, 4},
	}
	patterns := []struct {
		name string
		n    map[string]interface{}
		k    string
		x    interface{}
		ok   bool
	}{
		{"leaf", m, "a", 1, true},
		{"nested", m, "nested.b", 2, true},
		{"array", m, "array", []interface{}{1, 2, 3, 4}, true},
		{"array length", m, "array[]", 4, true},
		{"array element", m, "array[2]", 3, true},
		{"array element", m, "array[5]", nil, false},
		{"miss", m, "b", nil, false},
	}
	for _, p := range patterns {
		v, ok := getFromMSI(p.n, p.k, ".")
		assert.Equal(t, p.ok, ok, p.k)
		assert.Equal(t, p.x, v, p.k)
	}
}

func TestGetArrayElement(t *testing.T) {
	a := []interface{}{1, 2, 3, 4, 5}
	b := []interface{}{
		[]interface{}{1, 2, 3, 4, 5},
		[]interface{}{5, 6, 7, 8}}
	c := []interface{}{
		map[string]interface{}{"a": 1},
		map[interface{}]interface{}{"b": 2},
	}
	d := []interface{}{[]interface{}{
		map[string]interface{}{"a": 1},
		map[interface{}]interface{}{"b": 2},
	}}

	patterns := []struct {
		name string
		n    []interface{}
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

func TestGetNestedElement(t *testing.T) {
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
		{"neither", map[int]interface{}{1: 1}, "a", map[int]interface{}{1: 1}, true},
	}
	for _, p := range patterns {
		v, ok := getNestedElement(p.n, p.k, ".")
		assert.Equal(t, p.ok, ok, p.name)
		assert.Equal(t, p.x, v, p.name)
	}
}

func TestIsArrayLen(t *testing.T) {
	patterns := []struct {
		k  string
		x  string
		ok bool
	}{
		{"a", "a", false},
		{"a[]", "a", true},
		{"a[][]", "a[]", true},
		{"a[2][]", "a[2]", true},
		{"a[", "a[", false},
		{"a]", "a]", false},
		{"[]a", "[]a", false},
	}
	for _, p := range patterns {
		v, ok := isArrayLen(p.k)
		assert.Equal(t, p.ok, ok, p.k)
		assert.Equal(t, p.x, v, p.k)
	}
}

func TestParseArrayElement(t *testing.T) {
	patterns := []struct {
		k string
		x string
		i []int
	}{
		{"", "", nil},
		{"a", "a", nil},
		{"a[0]", "a", []int{0}},
		{"a[1][2]", "a", []int{1, 2}},
		{"a]", "a]", nil},
		{"a[1][notint]", "a[1][notint]", nil},
	}
	for _, p := range patterns {
		v, i := parseArrayElement(p.k)
		assert.Equal(t, p.i, i, p.k)
		assert.Equal(t, p.x, v, p.k)
	}
}
