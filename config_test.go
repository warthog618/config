// Copyright © 2017 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package config_test

import (
	"bytes"
	"encoding/gob"
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/warthog618/config"
	"github.com/warthog618/config/cfgconv"
)

func TestNewConfig(t *testing.T) {
	mr := mockGetter{"a.b.c_d": true}
	c := config.NewConfig(&mr)
	v, err := c.Get("")
	assert.IsType(t, config.NotFoundError{}, err)
	assert.Equal(t, nil, v)
	v, err = c.Get("a.b.c_d")
	assert.Nil(t, err)
	assert.Equal(t, true, v)
}

func TestNewMust(t *testing.T) {
	mr := mockGetter{"a.b.c_d": true}
	c := config.NewMust(&mr)
	v := c.Get("")
	assert.Equal(t, nil, v)
	v = c.Get("a.b.c_d")
	assert.Equal(t, true, v)
}

func TestGetBool(t *testing.T) {
	mr := mockGetter{
		"bool":       true,
		"boolString": "true",
		"boolInt":    1,
		"notabool":   "bogus",
	}
	patterns := []struct {
		k   string
		v   bool
		err error
	}{
		{"bool", true, nil},
		{"boolString", true, nil},
		{"boolInt", true, nil},
		{"notabool", false, &strconv.NumError{}},
		{"notsuchbool", false, config.NotFoundError{}},
	}
	c := config.NewConfig(&mr)
	f := func(t *testing.T) {
		for _, p := range patterns {
			v, err := c.GetBool(p.k)
			assert.IsType(t, p.err, err, p.k)
			assert.Equal(t, p.v, v, p.k)
		}
	}
	t.Run("Config", f)
	var err error
	m := config.NewMust(&mr, config.WithErrorHandler(
		func(e error) {
			err = e
		}))
	f = func(t *testing.T) {
		for _, p := range patterns {
			err = nil
			v := m.GetBool(p.k)
			assert.IsType(t, p.err, err)
			assert.Equal(t, p.v, v, p.k)
		}
	}
	t.Run("Must", f)
}

func TestGetDuration(t *testing.T) {
	mr := mockGetter{
		"duration":     "123ms",
		"notaduration": "bogus",
	}
	patterns := []struct {
		k   string
		v   time.Duration
		err error
	}{
		{"duration", time.Duration(123000000), nil},
		{"notaduration", time.Duration(0), errors.New("")},
		{"nosuchduration", time.Duration(0), config.NotFoundError{}},
	}
	c := config.NewConfig(&mr)
	f := func(t *testing.T) {
		for _, p := range patterns {
			v, err := c.GetDuration(p.k)
			assert.IsType(t, p.err, err, p.k)
			assert.Equal(t, p.v, v, p.k)
		}
	}
	t.Run("Config", f)
	var err error
	m := config.NewMust(&mr, config.WithErrorHandler(
		func(e error) {
			err = e
		}))
	f = func(t *testing.T) {
		for _, p := range patterns {
			err = nil
			v := m.GetDuration(p.k)
			assert.IsType(t, p.err, err)
			assert.Equal(t, p.v, v, p.k)
		}
	}
	t.Run("Must", f)
}

func TestGetFloat(t *testing.T) {
	mr := mockGetter{
		"float":        3.1415,
		"floatString":  "3.1415",
		"floatInt":     1,
		"notafloatInt": "bogus",
	}
	patterns := []struct {
		k   string
		v   float64
		err error
	}{
		{"float", 3.1415, nil},
		{"floatString", 3.1415, nil},
		{"floatInt", 1, nil},
		{"notafloatInt", 0, &strconv.NumError{}},
		{"nosuchfloat", 0, config.NotFoundError{}},
	}
	c := config.NewConfig(&mr)
	f := func(t *testing.T) {
		for _, p := range patterns {
			v, err := c.GetFloat(p.k)
			assert.IsType(t, p.err, err, p.k)
			assert.Equal(t, p.v, v, p.k)
		}
	}
	t.Run("Config", f)
	var err error
	m := config.NewMust(&mr, config.WithErrorHandler(
		func(e error) {
			err = e
		}))
	f = func(t *testing.T) {
		for _, p := range patterns {
			err = nil
			v := m.GetFloat(p.k)
			assert.IsType(t, p.err, err)
			assert.Equal(t, p.v, v, p.k)
		}
	}
	t.Run("Must", f)
}

func TestGetInt(t *testing.T) {
	mr := mockGetter{
		"int":       42,
		"intString": "43",
		"notaint":   "bogus",
	}
	patterns := []struct {
		k   string
		v   int64
		err error
	}{
		{"int", 42, nil},
		{"intString", 43, nil},
		{"notaint", 0, &strconv.NumError{}},
		{"nosuchint", 0, config.NotFoundError{}},
	}
	c := config.NewConfig(&mr)
	f := func(t *testing.T) {
		for _, p := range patterns {
			v, err := c.GetInt(p.k)
			assert.IsType(t, p.err, err, p.k)
			assert.Equal(t, p.v, v, p.k)
		}
	}
	t.Run("Config", f)
	var err error
	m := config.NewMust(&mr, config.WithErrorHandler(
		func(e error) {
			err = e
		}))
	f = func(t *testing.T) {
		for _, p := range patterns {
			err = nil
			v := m.GetInt(p.k)
			assert.IsType(t, p.err, err)
			assert.Equal(t, p.v, v, p.k)
		}
	}
	t.Run("Must", f)
}

func TestGetString(t *testing.T) {
	mr := mockGetter{
		"string":     "a string",
		"stringInt":  42,
		"notastring": struct{}{},
	}
	patterns := []struct {
		k   string
		v   string
		err error
	}{
		{"string", "a string", nil},
		{"stringInt", "42", nil},
		{"notastring", "", cfgconv.TypeError{}},
		{"nosuchstring", "", config.NotFoundError{}},
	}
	c := config.NewConfig(&mr)
	f := func(t *testing.T) {
		for _, p := range patterns {
			v, err := c.GetString(p.k)
			assert.IsType(t, p.err, err, p.k)
			assert.Equal(t, p.v, v, p.k)
		}
	}
	t.Run("Config", f)
	var err error
	m := config.NewMust(&mr, config.WithErrorHandler(
		func(e error) {
			err = e
		}))
	f = func(t *testing.T) {
		for _, p := range patterns {
			err = nil
			v := m.GetString(p.k)
			assert.IsType(t, p.err, err)
			assert.Equal(t, p.v, v, p.k)
		}
	}
	t.Run("Must", f)
}

func TestGetTime(t *testing.T) {
	mr := mockGetter{
		"time":     "2017-03-01T01:02:03Z",
		"notatime": "bogus",
	}
	patterns := []struct {
		k   string
		v   time.Time
		err error
	}{
		{"time", time.Date(2017, 3, 1, 1, 2, 3, 0, time.UTC), nil},
		{"notatime", time.Time{}, &time.ParseError{}},
		{"nosuchtime", time.Time{}, config.NotFoundError{}},
	}
	c := config.NewConfig(&mr)
	f := func(t *testing.T) {
		for _, p := range patterns {
			v, err := c.GetTime(p.k)
			assert.IsType(t, p.err, err, p.k)
			assert.Equal(t, p.v, v, p.k)
		}
	}
	t.Run("Config", f)
	var err error
	m := config.NewMust(&mr, config.WithErrorHandler(
		func(e error) {
			err = e
		}))
	f = func(t *testing.T) {
		for _, p := range patterns {
			err = nil
			v := m.GetTime(p.k)
			assert.IsType(t, p.err, err)
			assert.Equal(t, p.v, v, p.k)
		}
	}
	t.Run("Must", f)
}

func TestGetUint(t *testing.T) {
	mr := mockGetter{
		"uint":       42,
		"uintString": "43",
		"notaUint":   "bogus",
	}
	patterns := []struct {
		k   string
		v   uint64
		err error
	}{
		{"uint", 42, nil},
		{"uintString", 43, nil},
		{"notaUint", 0, &strconv.NumError{}},
		{"nosuchUint", 0, config.NotFoundError{}},
	}
	c := config.NewConfig(&mr)
	f := func(t *testing.T) {
		for _, p := range patterns {
			v, err := c.GetUint(p.k)
			assert.IsType(t, p.err, err, p.k)
			assert.Equal(t, p.v, v, p.k)
		}
	}
	t.Run("Config", f)
	var err error
	m := config.NewMust(&mr, config.WithErrorHandler(
		func(e error) {
			err = e
		}))
	f = func(t *testing.T) {
		for _, p := range patterns {
			err = nil
			v := m.GetUint(p.k)
			assert.IsType(t, p.err, err)
			assert.Equal(t, p.v, v, p.k)
		}
	}
	t.Run("Must", f)
}

func TestGetSlice(t *testing.T) {
	mr := mockGetter{
		"slice": []interface{}{1, 2, 3, 4},
		"animals": []interface{}{
			map[string]interface{}{"Name": "Platypus", "Order": "Monotremata"},
			map[string]interface{}{"Order": "Dasyuromorphia", "Name": "Quoll"},
		},
		"casttoslice": "bogus",
		"notaslice":   struct{}{},
	}
	patterns := []struct {
		k   string
		v   []interface{}
		err error
	}{
		{"slice", []interface{}{1, 2, 3, 4}, nil},
		{"animals", []interface{}{
			map[string]interface{}{"Name": "Platypus", "Order": "Monotremata"},
			map[string]interface{}{"Order": "Dasyuromorphia", "Name": "Quoll"},
		}, nil},
		{"casttoslice", []interface{}{"bogus"}, nil},
		{"notaslice", nil, cfgconv.TypeError{}},
		{"nosuchslice", nil, config.NotFoundError{}},
	}
	c := config.NewConfig(&mr)
	f := func(t *testing.T) {
		for _, p := range patterns {
			v, err := c.GetSlice(p.k)
			assert.IsType(t, p.err, err, p.k)
			assert.Equal(t, p.v, v, p.k)
		}
	}
	t.Run("Config", f)
	var err error
	m := config.NewMust(&mr, config.WithErrorHandler(
		func(e error) {
			err = e
		}))
	f = func(t *testing.T) {
		for _, p := range patterns {
			err = nil
			v := m.GetSlice(p.k)
			assert.IsType(t, p.err, err)
			assert.Equal(t, p.v, v, p.k)
		}
	}
	t.Run("Must", f)

}

func TestGetIntSlice(t *testing.T) {
	mr := mockGetter{
		"slice":       []int64{1, 2, -3, 4},
		"casttoslice": "42",
		"stringslice": []string{"one", "two", "three"},
		"notaslice":   "bogus",
	}
	patterns := []struct {
		k   string
		x   []int64
		err error
	}{
		{"slice", []int64{1, 2, -3, 4}, nil},
		{"casttoslice", []int64{42}, nil},
		{"stringslice", []int64{0, 0, 0}, &strconv.NumError{}},
		{"notaslice", []int64{0}, &strconv.NumError{}},
		{"nosuchslice", nil, config.NotFoundError{}},
	}
	c := config.NewConfig(&mr)
	f := func(t *testing.T) {
		for _, p := range patterns {
			v, err := c.GetIntSlice(p.k)
			assert.IsType(t, p.err, err, p.k)
			assert.Equal(t, p.x, v, p.k)
		}
	}
	t.Run("Config", f)
	var err error
	m := config.NewMust(&mr, config.WithErrorHandler(
		func(e error) {
			err = e
		}))
	f = func(t *testing.T) {
		for _, p := range patterns {
			err = nil
			v := m.GetIntSlice(p.k)
			assert.IsType(t, p.err, err)
			assert.Equal(t, p.x, v, p.k)
		}
	}
	t.Run("Must", f)
}

func TestGetStringSlice(t *testing.T) {
	mr := mockGetter{
		"intslice":        []int64{1, 2, -3, 4},
		"stringslice":     []string{"one", "two", "three"},
		"uintslice":       []uint64{1, 2, 3, 4},
		"notastringslice": []interface{}{1, 2, struct{}{}},
		"casttoslice":     "bogus",
		"notaslice":       struct{}{},
	}
	patterns := []struct {
		k   string
		x   []string
		err error
	}{
		{"intslice", []string{"1", "2", "-3", "4"}, nil},
		{"uintslice", []string{"1", "2", "3", "4"}, nil},
		{"stringslice", []string{"one", "two", "three"}, nil},
		{"casttoslice", []string{"bogus"}, nil},
		{"notastringslice", []string{"1", "2", ""}, cfgconv.TypeError{}},
		{"notaslice", nil, cfgconv.TypeError{}},
		{"nosuchslice", nil, config.NotFoundError{}},
	}
	c := config.NewConfig(&mr)
	f := func(t *testing.T) {
		for _, p := range patterns {
			v, err := c.GetStringSlice(p.k)
			assert.IsType(t, p.err, err, p.k)
			assert.Equal(t, p.x, v, p.k)
		}
	}
	t.Run("Config", f)
	var err error
	m := config.NewMust(&mr, config.WithErrorHandler(
		func(e error) {
			err = e
		}))
	f = func(t *testing.T) {
		for _, p := range patterns {
			err = nil
			v := m.GetStringSlice(p.k)
			assert.IsType(t, p.err, err)
			assert.Equal(t, p.x, v, p.k)
		}
	}
	t.Run("Must", f)
}

func TestGetUintSlice(t *testing.T) {
	mr := mockGetter{
		"slice":       []uint64{1, 2, 3, 4},
		"casttoslice": "42",
		"intslice":    []int64{1, 2, -3, 4},
		"stringslice": []string{"one", "two", "three"},
		"notaslice":   "bogus",
	}
	patterns := []struct {
		k   string
		x   []uint64
		err error
	}{
		{"slice", []uint64{1, 2, 3, 4}, nil},
		{"casttoslice", []uint64{42}, nil},
		{"intslice", []uint64{1, 2, 0, 4}, cfgconv.TypeError{}},
		{"stringslice", []uint64{0, 0, 0}, &strconv.NumError{}},
		{"notaslice", []uint64{0}, &strconv.NumError{}},
		{"nosuchslice", nil, config.NotFoundError{}},
	}
	c := config.NewConfig(&mr)
	f := func(t *testing.T) {
		for _, p := range patterns {
			v, err := c.GetUintSlice(p.k)
			assert.IsType(t, p.err, err, p.k)
			assert.Equal(t, p.x, v, p.k)
		}
	}
	t.Run("Config", f)
	var err error
	m := config.NewMust(&mr, config.WithErrorHandler(
		func(e error) {
			err = e
		}))
	f = func(t *testing.T) {
		for _, p := range patterns {
			err = nil
			v := m.GetUintSlice(p.k)
			assert.IsType(t, p.err, err)
			assert.Equal(t, p.x, v, p.k)
		}
	}
	t.Run("Must", f)
}

func TestGetConfig(t *testing.T) {
	// Single Getter
	mr := mockGetter{
		"foo.a": "foo.a",
		"foo.b": "foo.b",
		"bar.b": "bar.b",
		"bar.c": "bar.c",
		"baz.a": "baz.a",
		"baz.c": "baz.c",
	}
	a := config.NewAlias()
	a.Append("foo.d", "bar.c") // leaf alias
	a.Append("fuz", "foo")     // node alias

	type testPoint struct {
		k   string
		v   interface{}
		err error
	}
	patterns := []struct {
		name    string
		subtree string
		tp      []testPoint
	}{
		{"root", "", []testPoint{
			{"foo.a", "foo.a", nil},
			{"foo.b", "foo.b", nil},
			{"bar.b", "bar.b", nil},
			{"bar.c", "bar.c", nil},
			{"baz.a", "baz.a", nil},
			{"baz.c", "baz.c", nil},
		}},
		{"foo", "foo", []testPoint{
			{"a", "foo.a", nil},
			{"b", "foo.b", nil},
			{"c", nil, config.NotFoundError{}},
			{"d", "bar.c", nil},
		}},
		{"fuz", "fuz", []testPoint{
			{"a", "foo.a", nil},
			{"b", "foo.b", nil},
			{"c", nil, config.NotFoundError{}},
			{"d", nil, config.NotFoundError{}}, // ignores foo.d -> bar.c alias
		}},
		{"bar", "bar", []testPoint{
			{"a", nil, config.NotFoundError{}},
			{"b", "bar.b", nil},
			{"c", "bar.c", nil},
			{"e", nil, config.NotFoundError{}},
		}},
		{"baz", "baz", []testPoint{
			{"a", "baz.a", nil},
			{"b", nil, config.NotFoundError{}},
			{"c", "baz.c", nil},
			{"e", nil, config.NotFoundError{}},
		}},
		{"blah", "blah", []testPoint{
			{"a", nil, config.NotFoundError{}},
			{"b", nil, config.NotFoundError{}},
			{"c", nil, config.NotFoundError{}},
			{"e", nil, config.NotFoundError{}},
		}},
	}
	type getNode interface {
		GetConfig(node string, options ...config.ConfigOption) *config.Config
		GetMust(node string, options ...config.MustOption) *config.Must
	}
	configs := []struct {
		name string
		c    getNode
	}{{"config", config.NewConfig(config.Decorate(&mr, config.WithAlias(a)))},
		{"must", config.NewMust(config.Decorate(&mr, config.WithAlias(a)))},
	}
	for _, cfg := range configs {
		for _, p := range patterns {
			f := func(t *testing.T) {
				for _, tp := range p.tp {
					subc := cfg.c.GetConfig(p.subtree)
					require.NotNil(t, subc)
					v, err := subc.Get(tp.k)
					assert.IsType(t, tp.err, err, tp.k)
					assert.Equal(t, tp.v, v, tp.k)
				}
			}
			t.Run(cfg.name+"-Config-"+p.name, f)
			f = func(t *testing.T) {
				for _, tp := range p.tp {
					subc := cfg.c.GetMust(p.subtree)
					require.NotNil(t, subc)
					v := subc.Get(tp.k)
					assert.Equal(t, tp.v, v, tp.k)
				}
			}
			t.Run(cfg.name+"-Must-"+p.name, f)
		}
	}
}

func TestGetConfigWithSeparator(t *testing.T) {
	mr := mockGetter{
		"a.b.c_d": true,
	}
	c := config.NewConfig(&mr)
	v, err := c.Get("a.b.c_d")
	assert.Nil(t, err)
	assert.Equal(t, true, v)
	cfg := c.GetConfig("a", config.WithSeparator("_"))
	v, err = cfg.Get("b.c_d")
	assert.Nil(t, err)
	assert.Equal(t, true, v)
	cfg = cfg.GetConfig("b.c")
	v, err = cfg.Get("d")
	assert.Nil(t, err)
	assert.Equal(t, true, v)

	m := config.NewMust(&mr)
	v = m.Get("a.b.c_d")
	assert.Equal(t, true, v)
	cfg = m.GetConfig("a", config.WithSeparator("_"))
	v, err = cfg.Get("b.c_d")
	assert.Nil(t, err)
	assert.Equal(t, true, v)
	cfg = cfg.GetConfig("b.c")
	v, err = cfg.Get("d")
	assert.Nil(t, err)
	assert.Equal(t, true, v)
}

func TestGetMustWithSeparator(t *testing.T) {
	mr := mockGetter{
		"a.b.c_d": true,
	}
	c := config.NewConfig(&mr)
	v, err := c.Get("a.b.c_d")
	assert.Nil(t, err)
	assert.Equal(t, true, v)
	cfg := c.GetMust("a", config.WithSeparator("_"))
	v = cfg.Get("b.c_d")
	assert.Equal(t, true, v)
	cfg = cfg.GetMust("b.c")
	v = cfg.Get("d")
	assert.Nil(t, err)
	assert.Equal(t, true, v)

	m := config.NewMust(&mr)
	v = m.Get("a.b.c_d")
	assert.Equal(t, true, v)
	cfg = m.GetMust("a", config.WithSeparator("_"))
	v = cfg.Get("b.c_d")
	assert.Equal(t, true, v)
	cfg = cfg.GetMust("b.c")
	v = cfg.Get("d")
	assert.Nil(t, err)
	assert.Equal(t, true, v)
}

type fooConfig struct {
	Atagged int `config:"a"`
	B       string
	C       []int `cfg:"ctag"`
	E       string
	F       []innerConfig
	G       [][]int
	Nested  innerConfig `config:"nested"`
	p       string
}

type innerConfig struct {
	A       int
	Btagged string `config:"b_inner"`
	C       []int
	E       string
}

type aliasSetup func(*config.Alias)

func TestUnmarshal(t *testing.T) {
	blah := "blah"
	patterns := []struct {
		name   string
		g      config.Getter
		k      string
		target interface{}
		x      interface{}
		err    error
	}{
		{"non-struct target",
			&mockGetter{
				"foo.a": 42,
			},
			"foo",
			&blah,
			&blah,
			config.ErrInvalidStruct},
		{"non-pointer target",
			&mockGetter{
				"foo.a": 42,
			},
			"foo",
			fooConfig{},
			fooConfig{},
			config.ErrInvalidStruct},
		{"scalars",
			&mockGetter{
				"a": 42,
				"b": "foo.b",
				"d": "ignored",
				"p": "non-exported fields can't be set",
			},
			"",
			&fooConfig{},
			&fooConfig{
				Atagged: 42,
				B:       "foo.b"},
			nil},
		{"maltyped",
			&mockGetter{
				"a": []int{3, 4},
			},
			"",
			&fooConfig{},
			&fooConfig{},
			config.UnmarshalError{}},
		{"array of scalar",
			&mockGetter{
				"c": []int{1, 2, 3, 4},
				"d": "ignored",
			},
			"",
			&fooConfig{},
			&fooConfig{
				C: []int{1, 2, 3, 4}},
			nil},
		{"array of array",
			&mockGetter{
				"g": [][]int{{1, 2}, {3, 4}},
				"d": "ignored",
			},
			"",
			&fooConfig{},
			&fooConfig{
				G: [][]int{{1, 2}, {3, 4}}},
			nil},
		{"array of object",
			&mockGetter{
				"foo.f[]":    2,
				"foo.f[0].a": 1,
				"foo.f[1].a": 2,
				"foo.d":      "ignored",
			},
			"foo",
			&fooConfig{F: []innerConfig{}},
			&fooConfig{
				F: []innerConfig{
					{A: 1},
					{A: 2},
				}},
			nil},
		{"bad array of object",
			&mockGetter{
				"foo.f[]":    "notint",
				"foo.f[0].a": 1,
				"foo.f[1].a": 2,
				"foo.d":      "ignored",
			},
			"foo",
			&fooConfig{F: []innerConfig{}},
			&fooConfig{
				F: []innerConfig{}},
			&strconv.NumError{}},
		{"nested",
			&mockGetter{
				"foo.b":              "foo.b",
				"foo.nested.a":       43,
				"foo.nested.b_inner": "foo.nested.b",
				"foo.nested.c":       []int{5, 6, 7, 8},
			},
			"foo",
			&fooConfig{},
			&fooConfig{
				B: "foo.b",
				Nested: innerConfig{
					A:       43,
					Btagged: "foo.nested.b",
					C:       []int{5, 6, 7, 8}}},
			nil},
		{"nested wrong type",
			&mockGetter{
				"foo.nested.a": []int{6, 7},
			},
			"foo",
			&fooConfig{},
			&fooConfig{},
			config.UnmarshalError{}},
	}
	f := func(t *testing.T) {
		for _, p := range patterns {
			f := func(t *testing.T) {
				c := config.NewConfig(p.g)
				err := c.Unmarshal(p.k, p.target)
				assert.IsType(t, p.err, err)
				assert.Equal(t, p.x, p.target)
			}
			t.Run(p.name, f)
		}
	}
	t.Run("Config", f)
	f = func(t *testing.T) {
		for _, p := range patterns {
			f := func(t *testing.T) {
				var err error
				m := config.NewMust(p.g, config.WithErrorHandler(
					func(e error) {
						err = e
					}))
				m.Unmarshal(p.k, p.target)
				assert.IsType(t, p.err, err)
				assert.Equal(t, p.x, p.target)
			}
			t.Run(p.name, f)
		}
	}
	t.Run("Must", f)
}

func TestUnmarshalWithTag(t *testing.T) {
	patterns := []struct {
		name   string
		g      config.Getter
		k      string
		target interface{}
		x      interface{}
		err    error
	}{
		{"array of scalar",
			&mockGetter{
				"c":    []int{5, 6, 7, 8},
				"ctag": []int{1, 2, 3, 4},
				"d":    "ignored",
			},
			"",
			&fooConfig{},
			&fooConfig{
				C: []int{1, 2, 3, 4}},
			nil},
	}
	f := func(t *testing.T) {
		for _, p := range patterns {
			f := func(t *testing.T) {
				c := config.NewConfig(p.g, config.WithTag("cfg"))
				err := c.Unmarshal(p.k, p.target)
				assert.IsType(t, p.err, err)
				assert.Equal(t, p.x, p.target)
			}
			t.Run(p.name, f)
		}
	}
	t.Run("Config", f)
	f = func(t *testing.T) {
		for _, p := range patterns {
			f := func(t *testing.T) {
				var err error
				m := config.NewMust(p.g,
					config.WithTag("cfg"),
					config.WithErrorHandler(
						func(e error) {
							err = e
						}))
				m.Unmarshal(p.k, p.target)
				assert.IsType(t, p.err, err)
				assert.Equal(t, p.x, p.target)
			}
			t.Run(p.name, f)
		}
	}
	t.Run("Must", f)
}

func TestUnmarshalToMap(t *testing.T) {
	mg := mockGetter{
		"foo.a": 42,
		"foo.b": "foo.b",
		"foo.c": []int{1, 2, 3, 4},
		"foo.d": "ignored",
	}
	patterns := []struct {
		name   string
		g      config.Getter
		target map[string]interface{}
		x      map[string]interface{}
		err    error
	}{
		{"nil types",
			mg,
			map[string]interface{}{"a": nil, "b": nil, "c": nil, "e": nil},
			map[string]interface{}{
				"a": 42,
				"b": "foo.b",
				"c": []int{1, 2, 3, 4},
				"e": nil},
			nil,
		},
		{"typed",
			mg,
			map[string]interface{}{
				"a": int(0),
				"b": "",
				"c": []int{},
				"e": "some useful default"},
			map[string]interface{}{
				"a": 42,
				"b": "foo.b",
				"c": []int{1, 2, 3, 4},
				"e": "some useful default"},
			nil,
		},
		{"maltyped int",
			mg,
			map[string]interface{}{"a": []int{0}},
			map[string]interface{}{"a": []int{0}},
			config.UnmarshalError{},
		},
		{"maltyped string",
			mg,
			map[string]interface{}{"b": 2},
			map[string]interface{}{"b": 2},
			config.UnmarshalError{},
		},
		{"maltyped array",
			mg,
			map[string]interface{}{"c": 3},
			map[string]interface{}{"c": 3},
			config.UnmarshalError{},
		},
		{"raw array of arrays",
			mockGetter{
				"foo.aa": [][]int{{1, 2, 3, 4}, {4, 5, 6, 7}},
			},
			map[string]interface{}{"aa": nil},
			map[string]interface{}{
				"aa": [][]int{{1, 2, 3, 4}, {4, 5, 6, 7}},
			},
			nil,
		},
		{"array of array of int",
			mockGetter{
				"foo.aa": [][]int64{{1, 2, 3, 4}, {4, 5, 6, 7}},
			},
			map[string]interface{}{"aa": [][]int{}},
			map[string]interface{}{
				"aa": [][]int{{1, 2, 3, 4}, {4, 5, 6, 7}},
			},
			nil,
		},

		{"array of interface",
			mockGetter{
				"foo.aa": [][]int{{1, 2, 3, 4}, {4, 5, 6, 7}},
			},
			map[string]interface{}{"aa": []interface{}{}},
			map[string]interface{}{
				"aa": []interface{}{[]int{1, 2, 3, 4}, []int{4, 5, 6, 7}},
			},
			nil,
		},
		{"array of arrays",
			mockGetter{
				"foo.aa": [][]int{{1, 2, 3, 4}, {4, 5, 6, 7}},
			},
			map[string]interface{}{"aa": [][]interface{}{}},
			map[string]interface{}{
				"aa": [][]interface{}{{1, 2, 3, 4}, {4, 5, 6, 7}},
			},
			nil,
		},
		{"array of objects",
			&mockGetter{
				"foo.f[]":    2,
				"foo.f[0].A": 1,
				"foo.f[1].A": 2,
				"foo.d":      "ignored",
			},
			map[string]interface{}{
				"f": []map[string]interface{}{
					{"A": 0},
				},
			},
			map[string]interface{}{
				"f": []map[string]interface{}{
					{"A": 1},
					{"A": 2},
				},
			},
			nil,
		},
		{"bad array of objects",
			&mockGetter{
				"foo.f[]":    "notint",
				"foo.f[0].A": 1,
				"foo.f[1].A": 2,
				"foo.d":      "ignored",
			},
			map[string]interface{}{
				"f": []map[string]interface{}{
					{"A": 0},
				},
			},
			map[string]interface{}{
				"f": []map[string]interface{}{
					{"A": 0},
				},
			},
			&strconv.NumError{},
		},
		{"empty array of objects",
			&mockGetter{
				"foo.f[]":    2,
				"foo.f[0].A": 1,
				"foo.f[1].A": 2,
				"foo.d":      "ignored",
			},
			map[string]interface{}{
				"f": []map[string]interface{}{},
			},
			map[string]interface{}{
				"f": []map[string]interface{}(nil),
			},
			nil,
		},
		{"nil array of objects",
			&mockGetter{
				"foo.f[]":    2,
				"foo.f[0].A": 1,
				"foo.f[1].A": 2,
				"foo.d":      "ignored",
			},
			map[string]interface{}{
				"f": []map[string]interface{}(nil),
			},
			map[string]interface{}{
				"f": []map[string]interface{}(nil),
			},
			nil,
		},
		{"nested",
			mockGetter{
				"foo.a":        42,
				"foo.b":        "foo.b",
				"foo.c":        []int{1, 2, 3, 4},
				"foo.d":        "ignored",
				"foo.nested.a": 43,
				"foo.nested.b": "foo.nested.b",
				"foo.nested.c": []int{1, 2, -3, 4},
			},
			map[string]interface{}{
				"a": 0,
				"nested": map[string]interface{}{
					"a": int(0),
					"b": "",
					"c": []int{}},
			},
			map[string]interface{}{
				"a": 42,
				"nested": map[string]interface{}{
					"a": 43,
					"b": "foo.nested.b",
					"c": []int{1, 2, -3, 4}},
			},
			nil,
		},
		{"nested maltyped",
			mockGetter{
				"foo.a":        42,
				"foo.b":        "foo.b",
				"foo.c":        []int{1, 2, 3, 4},
				"foo.d":        "ignored",
				"foo.nested.a": []int{},
				"foo.nested.b": "foo.nested.b",
				"foo.nested.c": []int{1, 2, -3, 4},
			},
			map[string]interface{}{
				"a": 0,
				"nested": map[string]interface{}{
					"a": int(0),
					"b": "",
					"c": []int{}},
			},
			map[string]interface{}{
				"a": 42,
				"nested": map[string]interface{}{
					"a": 0,
					"b": "foo.nested.b",
					"c": []int{1, 2, -3, 4}},
			},
			config.UnmarshalError{},
		},
	}
	f := func(t *testing.T) {
		for _, p := range patterns {
			f := func(t *testing.T) {
				c := config.NewConfig(p.g)
				target, err := deepcopy(p.target)
				assert.Nil(t, err)
				require.NotNil(t, target)
				err = c.UnmarshalToMap("foo", target)
				assert.IsType(t, p.err, err)
				assert.Equal(t, p.x, target)
			}
			t.Run(p.name, f)
		}
	}
	t.Run("Config", f)
	f = func(t *testing.T) {
		for _, p := range patterns {
			f := func(t *testing.T) {
				var err error
				m := config.NewMust(p.g, config.WithErrorHandler(
					func(e error) {
						err = e
					}))
				target, err := deepcopy(p.target)
				assert.Nil(t, err)
				require.NotNil(t, target)
				m.UnmarshalToMap("foo", target)
				assert.IsType(t, p.err, err)
				assert.Equal(t, p.x, target)
			}
			t.Run(p.name, f)
		}
	}
	t.Run("Must", f)
}

type mockGetter map[string]interface{}

func (m mockGetter) Get(key string) (interface{}, bool) {
	v, ok := m[key]
	return v, ok
}

func init() {
	gob.Register([]interface{}{})
	gob.Register([][]interface{}{})
	gob.Register([][]int{})
	gob.Register(map[string]interface{}{})
	gob.Register([]map[string]interface{}{})
}

// deepcopy performs a deep copy of the given map m.
func deepcopy(m map[string]interface{}) (map[string]interface{}, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	dec := gob.NewDecoder(&buf)
	err := enc.Encode(m)
	if err != nil {
		return nil, err
	}
	var copy map[string]interface{}
	err = dec.Decode(&copy)
	if err != nil {
		return nil, err
	}
	return copy, nil
}
