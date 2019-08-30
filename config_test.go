// Copyright © 2017 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package config_test

import (
	"bytes"
	"encoding/gob"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/warthog618/config"
)

var defaultTimeout = 10 * time.Millisecond

func TestNew(t *testing.T) {
	mr := mockGetter{"a.b.c_d": true}
	c := config.New(&mr)
	require.NotNil(t, c)
	v, err := c.Get("")
	assert.IsType(t, config.NotFoundError{}, err)
	assert.Equal(t, nil, v.Value())
	v, err = c.Get("a.b.c_d")
	assert.Nil(t, err)
	assert.Equal(t, true, v.Value())

	// nil getter
	c = config.New(nil)
	v, err = c.Get("")
	assert.IsType(t, config.NotFoundError{}, err)
	assert.Equal(t, nil, v.Value())
	v, err = c.Get("a.b.c_d")
	assert.IsType(t, config.NotFoundError{}, err)
	assert.Equal(t, nil, v.Value())

	// multiple getters
	mr2 := mockGetterAsOption{mockGetter: mockGetter{"a.b.c_d": false, "a.b.c_e": 2}}
	mr3 := mockGetterAsOption{mockGetter: mockGetter{"a.b.c_d": false, "a.b.c_e": 3}}
	c = config.New(&mr, &mr2, &mr3)
	require.NotNil(t, c)
	v, err = c.Get("")
	assert.IsType(t, config.NotFoundError{}, err)
	assert.Equal(t, nil, v.Value())
	v, err = c.Get("a.b.c_d")
	assert.Nil(t, err)
	assert.Equal(t, true, v.Value())
	v, err = c.Get("a.b.c_e")
	assert.Nil(t, err)
	assert.Equal(t, int64(2), v.Int())
}

func TestNewWithErrorHandler(t *testing.T) {
	mr := mockGetter{"a.b.c_d": "this is a.b.c.d"}
	var eherr error
	eh := func(err error) error {
		eherr = err
		return nil
	}
	c := config.New(&mr, config.WithErrorHandler(eh))
	// get fail
	v, err := c.Get("")
	assert.Nil(t, err)
	assert.IsType(t, config.NotFoundError{}, eherr)
	assert.Equal(t, nil, v.Value())
	eherr = nil
	// get ok
	v, err = c.Get("a.b.c_d")
	assert.Nil(t, err)
	assert.Nil(t, eherr)
	assert.Equal(t, "this is a.b.c.d", v.Value())
	// convert fail
	v.Int()
	assert.IsType(t, &strconv.NumError{}, eherr)
}

func TestNewWithGetErrorHandler(t *testing.T) {
	mr := mockGetter{"a.b.c_d": "this is a.b.c.d"}
	var eherr error
	eh := func(err error) error {
		eherr = err
		return nil
	}
	c := config.New(&mr, config.WithGetErrorHandler(eh))
	// get fail
	v, err := c.Get("")
	assert.Nil(t, err)
	assert.IsType(t, config.NotFoundError{}, eherr)
	assert.Equal(t, nil, v.Value())
	eherr = nil
	// get ok
	v, err = c.Get("a.b.c_d")
	assert.Nil(t, err)
	assert.Nil(t, eherr)
	assert.Equal(t, "this is a.b.c.d", v.Value())
	// convert fail
	v.Int()
	assert.Nil(t, eherr)
}

func TestNewWithValueErrorHandler(t *testing.T) {
	mr := mockGetter{"a.b.c_d": "this is a.b.c.d"}
	var eherr error
	eh := func(err error) error {
		eherr = err
		return nil
	}
	c := config.New(&mr, config.WithValueErrorHandler(eh))
	// get fail
	v, err := c.Get("")
	assert.IsType(t, config.NotFoundError{}, err)
	assert.Nil(t, eherr)
	assert.Equal(t, nil, v.Value())
	// get ok
	v, err = c.Get("a.b.c_d")
	assert.Nil(t, err)
	assert.Nil(t, eherr)
	assert.Equal(t, "this is a.b.c.d", v.Value())
	// convert fail
	v.Int()
	assert.IsType(t, &strconv.NumError{}, eherr)
}

func TestNewWithZeroDefaults(t *testing.T) {
	mr := mockGetter{"a.b.c_d": true}
	c := config.New(&mr, config.WithZeroDefaults())
	v, err := c.Get("not.a.b.c_d")
	assert.Nil(t, err)
	assert.Equal(t, nil, v.Value())
	assert.Equal(t, int64(0), v.Int())
	v, err = c.Get("a.b.c_d")
	assert.Nil(t, err)
	assert.Equal(t, true, v.Value())
	assert.NotPanics(t, func() {
		v := c.MustGet("not.a.b.c_d")
		assert.Equal(t, nil, v.Value())
		assert.Equal(t, int64(0), v.Int())
	})
}

func TestNewWithMust(t *testing.T) {
	mr := mockGetter{"a.b.c_d": true}
	c := config.New(&mr, config.WithMust())
	v, err := c.Get("a.b.c_d")
	assert.Nil(t, err)
	assert.Equal(t, true, v.Value())
	assert.Panics(t, func() {
		c.Get("")
	})
}

func TestAppend(t *testing.T) {
	mr1 := mockGetter{
		"foo":   "this is foo",
		"bar.b": "this is bar.b",
		"id":    1,
	}
	mrA := mockGetter{
		"foo":   "this is foo",
		"bar.b": "this is bar.b",
		"id":    2,
	}
	cfg0 := config.New(nil)
	cfg1 := config.New(&mr1)
	patterns := []struct {
		name string
		cfg  *config.Config
		k    string
		v    interface{}
		err  error
	}{
		{"solo", cfg0, "foo", "this is foo", nil},
		{"solo", cfg0, "bar.b", "this is bar.b", nil},
		{"solo", cfg0, "id", 2, nil},
		{"solo", cfg0, "nosuch", nil, config.NotFoundError{}},
		{"underlay", cfg1, "foo", "this is foo", nil},
		{"underlay", cfg1, "bar.b", "this is bar.b", nil},
		{"underlay", cfg1, "id", 1, nil},
		{"underlay", cfg1, "nosuch", nil, config.NotFoundError{}},
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			cfg := p.cfg.GetConfig("")
			cfg.Append(&mrA)
			v, err := cfg.Get(p.k)
			assert.IsType(t, p.err, err)
			assert.Equal(t, p.v, v.Value())
		}
		t.Run(p.name+"-"+p.k, f)
	}
}

func TestAppendNil(t *testing.T) {
	mr := mockGetter{
		"id": 1,
	}
	cfg := config.New(&mr)
	cfg.Append(nil)
	v, err := cfg.Get("id")
	assert.Nil(t, err)
	assert.Equal(t, 1, v.Value())
}

func TestClose(t *testing.T) {
	mr := mockGetter{
		"foo":   "this is foo",
		"bar.b": "this is bar.b",
	}
	cfg := config.New(&mr)
	w := cfg.NewWatcher()
	testNotUpdated(t, w, nil)
	done := make(chan struct{})
	go func() {
		<-time.After(defaultTimeout)
		close(done)
	}()
	cfg.Close()
	err := w.Watch(done)
	assert.Equal(t, config.ErrClosed, err)
	cfg.Close() // can be closed repeatedly
	err = w.Watch(done)
	assert.Equal(t, config.ErrClosed, err)
}

func TestGet(t *testing.T) {
	mr := mockGetter{
		"foo":   "this is foo",
		"bar.b": "this is bar.b",
	}
	patterns := []struct {
		k   string
		v   interface{}
		err error
	}{
		{"foo", "this is foo", nil},
		{"bar.b", "this is bar.b", nil},
		{"nosuch", nil, config.NotFoundError{}},
	}
	cfg := config.New(&mr)
	for _, p := range patterns {
		f := func(t *testing.T) {
			v, err := cfg.Get(p.k)
			assert.IsType(t, p.err, err)
			assert.Equal(t, p.v, v.Value())
		}
		t.Run(p.k, f)
	}
}

func TestGetWithErrorHandler(t *testing.T) {
	mr := mockGetter{
		"foo": "this is foo",
	}
	var eherr error
	eh := func(err error) error {
		eherr = err
		return err
	}
	patterns := []struct {
		k   string
		vo  config.ValueOption
		err error
	}{
		{"nil", config.WithErrorHandler(nil), nil},
		{"eh", config.WithErrorHandler(eh), &strconv.NumError{}},
	}
	cfg := config.New(&mr)
	for _, p := range patterns {
		f := func(t *testing.T) {
			eherr = nil
			v, err := cfg.Get("foo", p.vo)
			assert.Nil(t, err)
			assert.Equal(t, "this is foo", v.Value())
			assert.Nil(t, eherr)
			vi := v.Int()
			assert.Equal(t, int64(0), vi)
			assert.IsType(t, p.err, eherr)
		}
		t.Run(p.k, f)
	}
}

func TestGetWithPanic(t *testing.T) {
	mr := mockGetter{
		"foo": "this is foo",
	}
	cfg := config.New(&mr)
	assert.Panics(t, func() {
		v, _ := cfg.Get("foo", config.WithMust())
		v.Int()
	})
}

func TestGetConfig(t *testing.T) {
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
	cfg := config.New(config.Decorate(&mr, config.WithAlias(a)))
	for _, p := range patterns {
		f := func(t *testing.T) {
			for _, tp := range p.tp {
				subc := cfg.GetConfig(p.subtree)
				require.NotNil(t, subc)
				v, err := subc.Get(tp.k)
				assert.IsType(t, tp.err, err, tp.k)
				assert.Equal(t, tp.v, v.Value(), tp.k)
			}
		}
		t.Run(p.name, f)
	}
}

func TestGetConfigWithSeparator(t *testing.T) {
	mr := mockGetter{
		"a.b.c_d": true,
	}
	c := config.New(&mr)
	v, err := c.Get("a.b.c_d")
	assert.Nil(t, err)
	assert.Equal(t, true, v.Value())
	cfg := c.GetConfig("a", config.WithSeparator("_"))
	v, err = cfg.Get("b.c_d")
	assert.Nil(t, err)
	assert.Equal(t, true, v.Value())
	cfg = cfg.GetConfig("b.c")
	v, err = cfg.Get("d")
	assert.Nil(t, err)
	assert.Equal(t, true, v.Value())
}

func TestInsert(t *testing.T) {
	mr1 := mockGetter{
		"foo":   "this is foo",
		"bar.b": "this is bar.b",
		"id":    1,
	}
	mrI := mockGetter{
		"foo":   "this is foo",
		"bar.b": "this is bar.b",
		"id":    2,
	}
	cfg0 := config.New(nil)
	cfg1 := config.New(&mr1)
	patterns := []struct {
		name string
		cfg  *config.Config
		k    string
		v    interface{}
		err  error
	}{
		{"solo", cfg0, "foo", "this is foo", nil},
		{"solo", cfg0, "bar.b", "this is bar.b", nil},
		{"solo", cfg0, "id", 2, nil},
		{"solo", cfg0, "nosuch", nil, config.NotFoundError{}},
		{"overlay", cfg1, "foo", "this is foo", nil},
		{"overlay", cfg1, "bar.b", "this is bar.b", nil},
		{"overlay", cfg1, "id", 2, nil},
		{"overlay", cfg1, "nosuch", nil, config.NotFoundError{}},
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			cfg := p.cfg.GetConfig("")
			cfg.Insert(&mrI)
			v, err := cfg.Get(p.k)
			assert.IsType(t, p.err, err)
			assert.Equal(t, p.v, v.Value())
		}
		t.Run(p.name+"-"+p.k, f)
	}
}

func TestInsertNil(t *testing.T) {
	mr := mockGetter{
		"id": 1,
	}
	cfg := config.New(&mr)
	cfg.Insert(nil)
	v, err := cfg.Get("id")
	assert.Nil(t, err)
	assert.Equal(t, 1, v.Value())
}

func TestMustGet(t *testing.T) {
	mr := mockGetter{
		"foo":   "this is foo",
		"bar.b": "this is bar.b",
	}
	type testPoint struct {
	}
	patterns := []struct {
		name string
		k    string
		v    interface{}
		eh   config.ErrorHandler
		err  error
	}{
		{"hit", "foo", "this is foo", nil, nil},
		{"hit2", "bar.b", "this is bar.b", nil, nil},
		{"miss", "nosuch", nil, nil, config.NotFoundError{Key: "nosuch"}},
	}
	cfg := config.New(&mr)
	for _, p := range patterns {
		f := func(t *testing.T) {
			var v config.Value
			if p.err == nil {
				assert.NotPanics(t, func() {
					v = cfg.MustGet(p.k)
				})
				assert.Equal(t, p.v, v.String())
			} else {
				assert.PanicsWithValue(t, p.err, func() {
					v = cfg.MustGet(p.k)
				})
			}
		}
		t.Run(p.name, f)
	}
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
	for _, p := range patterns {
		f := func(t *testing.T) {
			c := config.New(p.g)
			err := c.Unmarshal(p.k, p.target)
			assert.IsType(t, p.err, err)
			assert.Equal(t, p.x, p.target)
		}
		t.Run(p.name, f)
	}
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
	for _, p := range patterns {
		f := func(t *testing.T) {
			c := config.New(p.g, config.WithTag("cfg"))
			err := c.Unmarshal(p.k, p.target)
			assert.IsType(t, p.err, err)
			assert.Equal(t, p.x, p.target)
		}
		t.Run(p.name, f)
	}
}

func TestUnmarshalToMap(t *testing.T) {
	mg := &mockGetter{
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
			&mockGetter{
				"foo.aa": [][]int{{1, 2, 3, 4}, {4, 5, 6, 7}},
			},
			map[string]interface{}{"aa": nil},
			map[string]interface{}{
				"aa": [][]int{{1, 2, 3, 4}, {4, 5, 6, 7}},
			},
			nil,
		},
		{"array of array of int",
			&mockGetter{
				"foo.aa": [][]int64{{1, 2, 3, 4}, {4, 5, 6, 7}},
			},
			map[string]interface{}{"aa": [][]int{}},
			map[string]interface{}{
				"aa": [][]int{{1, 2, 3, 4}, {4, 5, 6, 7}},
			},
			nil,
		},

		{"array of interface",
			&mockGetter{
				"foo.aa": [][]int{{1, 2, 3, 4}, {4, 5, 6, 7}},
			},
			map[string]interface{}{"aa": []interface{}{}},
			map[string]interface{}{
				"aa": []interface{}{[]int{1, 2, 3, 4}, []int{4, 5, 6, 7}},
			},
			nil,
		},
		{"array of arrays",
			&mockGetter{
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
			&mockGetter{
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
			&mockGetter{
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
	for _, p := range patterns {
		f := func(t *testing.T) {
			c := config.New(p.g)
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

func TestNewWatcher(t *testing.T) {
	mr := mockGetter{
		"foo":   "this is foo",
		"bar.b": "this is bar.b",
	}

	// Static config
	cfg := config.New(&mr)
	w := cfg.NewWatcher()
	testNotUpdated(t, w, nil)

	// config Closed
	wg := watchedGetter{mr, nil}
	cfg = config.New(&wg)
	done := make(chan struct{})
	defer close(done)
	w = cfg.NewWatcher()
	updated := make(chan error)
	go func() {
		e := w.Watch(done)
		updated <- e
	}()
	cfg.Close()
	time.Sleep(defaultTimeout)
	select {
	case e := <-updated:
		assert.Equal(t, config.ErrClosed, e)
	case <-time.After(time.Second):
		assert.Fail(t, "watch failed to return")
	}

	// Updated
	wg = watchedGetter{mr, nil}
	cfg = config.New(&wg)
	ws := wg.w
	require.NotNil(t, ws)
	w = cfg.NewWatcher()
	require.NotNil(t, w)
	testUpdated(t, w, ws.Notify)

	// Cancelled
	done = make(chan struct{})
	w = cfg.NewWatcher()
	updated = make(chan error)
	go func() {
		e := w.Watch(done)
		updated <- e
	}()
	close(done)
	time.Sleep(defaultTimeout)
	select {
	case e := <-updated:
		assert.Equal(t, config.ErrCanceled, e)
	case <-time.After(time.Second):
		assert.Fail(t, "watch failed to return")
	}

	// updatech closed by getter
	done = make(chan struct{})
	defer close(done)
	w = cfg.NewWatcher()
	updated = make(chan error)
	go func() {
		e := w.Watch(done)
		updated <- e
	}()
	close(wg.w.updatech)
	time.Sleep(defaultTimeout)
	select {
	case e := <-updated:
		assert.Fail(t, "unexpected update", "err: %#v", e)
	case <-time.After(defaultTimeout):
	}
}

type watcher interface {
	Watch(done <-chan struct{}) error
}

func testUpdated(t *testing.T, w watcher, notify func()) {
	t.Helper()
	done := make(chan struct{})
	updated := make(chan error)
	go func() {
		err := w.Watch(done)
		updated <- err
	}()
	if notify != nil {
		notify()
	}
	select {
	case err := <-updated:
		assert.Nil(t, err)
	case <-time.After(time.Second):
		assert.Fail(t, "watch failed to return")
	}
	close(done)
}

func testNotUpdated(t *testing.T, w watcher, notify func()) {
	t.Helper()
	done := make(chan struct{})
	updated := make(chan error)
	go func() {
		err := w.Watch(done)
		updated <- err
	}()
	if notify != nil {
		notify()
	}
	select {
	case err := <-updated:
		assert.Fail(t, "unexpected update", "err: %#v", err)
	case <-time.After(5 * defaultTimeout):
	}
	close(done)
}

type keyWatcher interface {
	Watch(done <-chan struct{}) (config.Value, error)
}

func testKeyUpdated(t *testing.T, w keyWatcher, notify func(), xv string, xerr error) {
	t.Helper()
	done := make(chan struct{})
	updated := make(chan error)
	var v config.Value
	go func() {
		var err error
		v, err = w.Watch(done)
		updated <- err
	}()
	if notify != nil {
		notify()
	}
	select {
	case err := <-updated:
		assert.Equal(t, xerr, err)
		assert.Equal(t, xv, v.String())
	case <-time.After(time.Second):
		assert.Fail(t, "watch failed to return")
	}
	close(done)
}

func testKeyNotUpdated(t *testing.T, w keyWatcher, notify func()) {
	t.Helper()
	done := make(chan struct{})
	updated := make(chan error)
	var v config.Value
	go func() {
		var err error
		v, err = w.Watch(done)
		updated <- err
	}()
	if notify != nil {
		notify()
	}
	select {
	case err := <-updated:
		assert.Fail(t, "unexpected update")
		assert.Nil(t, err)
		assert.Equal(t, "", v.String())
	case <-time.After(5 * defaultTimeout):
	}
	close(done)
	// tidy up as mockGetter is not mt-safe...
	select {
	case <-updated:
		// his watch is done
	case <-time.After(time.Second):
		assert.Fail(t, "watcher failed to exit on close")
	}
}

func TestNewKeyWatcher(t *testing.T) {
	mr := mockGetter{
		"foo":   "this is foo",
		"bar.b": "this is bar.b",
	}
	// Static config
	cfg := config.New(&mr)
	w := cfg.NewKeyWatcher("foo")
	testKeyUpdated(t, w, nil, "this is foo", nil)

	// Updated
	wg := watchedGetter{mr, nil}
	cfg = config.New(&wg)
	ws := wg.w
	require.NotNil(t, ws)
	w = cfg.NewKeyWatcher("foo")
	require.NotNil(t, w)
	testKeyUpdated(t, w, nil, "this is foo", nil)
	testKeyNotUpdated(t, w, nil)

	// unchanged
	testKeyNotUpdated(t, w, ws.Notify)

	// changed
	mr["foo"] = "this is new foo"
	testKeyUpdated(t, w, ws.Notify, "this is new foo", nil)

	// Deleted
	delete(mr, "foo")
	testKeyUpdated(t, w, ws.Notify, "", config.NotFoundError{Key: "foo"})

	// Cancelled
	mr["foo"] = "this is foo too"
	done := make(chan struct{})
	w = cfg.NewKeyWatcher("foo")
	v, err := w.Watch(done)
	assert.Nil(t, err)
	assert.Equal(t, "this is foo too", v.String())
	updated := make(chan error)
	go func() {
		v, err = w.Watch(done)
		updated <- err
	}()
	// unchanged
	select {
	case <-updated:
		assert.Fail(t, "unexpected value update")
	case <-time.After(defaultTimeout):
	}
	close(done)
	time.Sleep(defaultTimeout)
	mr["foo"] = "this is new foo too"
	select {
	case e := <-updated:
		assert.Equal(t, config.ErrCanceled, e)
	case <-time.After(time.Second):
		assert.Fail(t, "didn't cancel")
	}
}

type mockGetter map[string]interface{}

func (m *mockGetter) Get(key string) (interface{}, bool) {
	v, ok := (*m)[key]
	return v, ok
}

type mockGetterAsOption struct {
	config.GetterAsOption
	mockGetter
}

type watchedGetter struct {
	mockGetter
	w *getterWatcher
}

func (w *watchedGetter) NewWatcher(donech <-chan struct{}) config.GetterWatcher {
	if w.w == nil {
		w.w = &getterWatcher{donech: donech, updatech: make(chan config.GetterUpdate)}
	}
	return w.w
}

type getterWatcher struct {
	mu        sync.RWMutex
	donech    <-chan struct{}
	updatech  chan config.GetterUpdate
	Committed bool
}

func (w *getterWatcher) Update() <-chan config.GetterUpdate {
	return w.updatech
}

func (w *getterWatcher) Notify() {
	w.updatech <- update{w: w}
}

type update struct {
	w *getterWatcher
}

func (u update) Commit() {
	u.w.Committed = true
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
