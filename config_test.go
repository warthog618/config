// Copyright © 2017 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package config_test

import (
	"bytes"
	"context"
	"encoding/gob"
	"strconv"
	"testing"
	"time"

	"github.com/pkg/errors"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/warthog618/config"
)

var defaultTimeout = 10 * time.Millisecond

func TestNewConfig(t *testing.T) {
	mr := mockGetter{"a.b.c_d": true}
	c := config.NewConfig(&mr)
	require.NotNil(t, c)
	v, err := c.Get("")
	assert.IsType(t, config.NotFoundError{}, err)
	assert.Equal(t, nil, v.Value())
	v, err = c.Get("a.b.c_d")
	assert.Nil(t, err)
	assert.Equal(t, true, v.Value())
}

func TestNewConfigWithErrorHandler(t *testing.T) {
	mr := mockGetter{"a.b.c_d": "this is a.b.c.d"}
	var eherr error
	eh := func(err error) error {
		eherr = err
		return nil
	}
	c := config.NewConfig(&mr, config.WithErrorHandler(eh))
	v, err := c.Get("a.b.c_d")
	assert.Nil(t, err)
	assert.Nil(t, eherr)
	assert.Equal(t, "this is a.b.c.d", v.Value())
	v.Int()
	assert.IsType(t, &strconv.NumError{}, eherr)
	eherr = nil
	v, err = c.Get("")
	assert.Nil(t, err)
	assert.IsType(t, config.NotFoundError{}, eherr)
	assert.Equal(t, nil, v.Value())
	eherr = nil
}

func TestNewConfigWithZeroDefaults(t *testing.T) {
	mr := mockGetter{"a.b.c_d": true}
	c := config.NewConfig(&mr, config.WithZeroDefaults())
	v, err := c.Get("not.a.b.c_d")
	assert.Nil(t, err)
	assert.Equal(t, nil, v.Value())
	assert.Equal(t, int64(0), v.Int())
	v, err = c.Get("a.b.c_d")
	assert.Nil(t, err)
	assert.Equal(t, true, v.Value())
}

func TestNewConfigWithMust(t *testing.T) {
	mr := mockGetter{"a.b.c_d": true}
	c := config.NewConfig(&mr, config.WithMust())
	v, err := c.Get("a.b.c_d")
	assert.Nil(t, err)
	assert.Equal(t, true, v.Value())
	assert.Panics(t, func() {
		c.Get("")
	})
}

func TestAddWatchedGetter(t *testing.T) {
	mr := mockGetter{
		"foo":   "this is foo",
		"bar.b": "this is bar.b",
	}
	ws := &watchedGetter{n: make(chan struct{})}
	cfg := config.NewConfig(&mr)
	w := cfg.NewWatcher()
	cfg.AddWatchedGetter(ws)

	// Updated
	ws.Notify()
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	err := w.Watch(ctx)
	assert.Nil(t, err)
	assert.True(t, ws.Committed)
	cancel()

	// Temporary error
	ws.Committed = false
	ws.WatchError = config.WithTemporary(errors.New("temp watch error"))
	ws.Notify()
	ctx, cancel = context.WithTimeout(context.Background(), defaultTimeout)
	err = w.Watch(ctx)
	assert.Equal(t, context.DeadlineExceeded, err)
	assert.False(t, ws.Committed)
	cancel()

	// exit
	ws = &watchedGetter{n: make(chan struct{})}
	cfg.AddWatchedGetter(ws)
	ws.WatchError = errors.New("watch error")
	ws.Notify()
	ctx, cancel = context.WithTimeout(context.Background(), defaultTimeout)
	err = w.Watch(ctx)
	assert.Equal(t, context.DeadlineExceeded, err)
	assert.False(t, ws.Committed)
	cancel()
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
	cfg := config.NewConfig(&mr)
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
	cfg := config.NewConfig(&mr)
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
	cfg := config.NewConfig(&mr)
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
	cfg := config.NewConfig(config.Decorate(&mr, config.WithAlias(a)))
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
	c := config.NewConfig(&mr)
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
	cfg := config.NewConfig(&mr)
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
			c := config.NewConfig(p.g)
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
			c := config.NewConfig(p.g, config.WithTag("cfg"))
			err := c.Unmarshal(p.k, p.target)
			assert.IsType(t, p.err, err)
			assert.Equal(t, p.x, p.target)
		}
		t.Run(p.name, f)
	}
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

func TestNewWatcher(t *testing.T) {
	mr := mockGetter{
		"foo":   "this is foo",
		"bar.b": "this is bar.b",
	}
	// Static config
	cfg := config.NewConfig(&mr)
	ctx := context.Background()
	w := cfg.NewWatcher()
	testWatcher(t, w, context.DeadlineExceeded)

	// Updated
	ws := &watchedGetter{n: make(chan struct{})}
	cfg = config.NewConfig(&mr, config.WithWatchedGetter(ws))
	w = cfg.NewWatcher()
	ws.Notify()
	testWatcher(t, w, nil)

	// Cancelled
	ctx, cancel := context.WithCancel(context.Background())
	w = cfg.NewWatcher()
	updated := make(chan error)
	go func() {
		e := w.Watch(ctx)
		updated <- e
	}()
	cancel()
	time.Sleep(defaultTimeout)
	select {
	case e := <-updated:
		assert.Equal(t, context.Canceled, e)
	case <-time.After(defaultTimeout):
	}
}

type watcher interface {
	Watch(context.Context) error
}

func testWatcher(t *testing.T, w watcher, xerr error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	updated := make(chan error)
	go func() {
		err := w.Watch(ctx)
		updated <- err
	}()
	select {
	case err := <-updated:
		assert.Equal(t, xerr, errors.Cause(err))
	case <-time.After(time.Second):
		assert.Fail(t, "watch failed to return")
	}
	cancel()
}

type keyWatcher interface {
	Watch(context.Context) (config.Value, error)
}

func testKeyWatcher(t *testing.T, w keyWatcher, xv string, xerr error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	updated := make(chan error)
	var v config.Value
	go func() {
		var err error
		v, err = w.Watch(ctx)
		updated <- err
	}()
	select {
	case err := <-updated:
		assert.Equal(t, xerr, err)
		assert.Equal(t, xv, v.String())
	case <-time.After(time.Second):
		assert.Fail(t, "watch failed to return")
	}
	cancel()
}

func TestNewKeyWatcher(t *testing.T) {
	mr := mockGetter{
		"foo":   "this is foo",
		"bar.b": "this is bar.b",
	}
	// Static config
	cfg := config.NewConfig(&mr)
	w := cfg.NewKeyWatcher("foo")
	testKeyWatcher(t, w, "this is foo", nil)

	// Updated
	ws := &watchedGetter{n: make(chan struct{})}
	cfg = config.NewConfig(&mr, config.WithWatchedGetter(ws))
	w = cfg.NewKeyWatcher("foo")
	testKeyWatcher(t, w, "this is foo", nil)
	testKeyWatcher(t, w, "", context.DeadlineExceeded)

	// unchanged
	ws.Notify()
	testKeyWatcher(t, w, "", context.DeadlineExceeded)

	// changed
	mr["foo"] = "this is new foo"
	ws.Notify()
	testKeyWatcher(t, w, "this is new foo", nil)

	// Deleted
	delete(mr, "foo")
	ws.Notify()
	testKeyWatcher(t, w, "", config.NotFoundError{Key: "foo"})

	// Cancelled
	mr["foo"] = "this is foo too"
	ctx, cancel := context.WithCancel(context.Background())
	w = cfg.NewKeyWatcher("foo")
	v, err := w.Watch(ctx)
	assert.Nil(t, err)
	assert.Equal(t, "this is foo too", v.String())
	updated := make(chan error)
	go func() {
		v, err = w.Watch(ctx)
		updated <- err
	}()
	// unchanged
	select {
	case <-updated:
		assert.Fail(t, "unexpected value update")
	case <-time.After(defaultTimeout):
	}
	cancel()
	time.Sleep(defaultTimeout)
	mr["foo"] = "this is new foo too"
	select {
	case e := <-updated:
		assert.Equal(t, context.Canceled, e)
	case <-time.After(defaultTimeout):
		assert.Fail(t, "didn't cancel")
	}
}

type mockGetter map[string]interface{}

func (m mockGetter) Get(key string) (interface{}, bool) {
	v, ok := m[key]
	return v, ok
}

type watchedGetter struct {
	n          chan struct{}
	WatchError error
	Committed  bool
}

func (w *watchedGetter) Watch(ctx context.Context) error {
	<-w.n
	return w.WatchError
}

func (w *watchedGetter) CommitUpdate() {
	w.n = make(chan struct{})
	w.Committed = true
}

func (w *watchedGetter) Notify() {
	close(w.n)
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
