// Copyright © 2018 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package config_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/warthog618/config"
)

func TestNewStack(t *testing.T) {
}

func TestStackAppend(t *testing.T) {
	mr1 := mockGetter{}
	s := config.NewStack(&mr1)
	require.NotNil(t, s)
	s.Append(nil) // should be ignored
	mr1["something"] = "a test string"
	v, ok := s.Get("something")
	assert.True(t, ok)
	assert.Exactly(t, mr1["something"], v)

	// append a second reader
	mr2 := mockGetter{
		"something":      "another test string",
		"something else": "yet another test string",
	}
	s.Append(&mr2)
	v, ok = s.Get("something")
	assert.True(t, ok)
	assert.Exactly(t, mr1["something"], v)
	v, ok = s.Get("something else")
	assert.True(t, ok)
	assert.Exactly(t, mr2["something else"], v)

	// append a third reader
	mr3 := mockGetter{
		"something":       "yet another test string",
		"something else":  "yet another test string",
		"something three": "third test string",
	}
	s.Append(&mr3)
	v, ok = s.Get("something")
	assert.True(t, ok)
	assert.Exactly(t, mr1["something"], v)
	v, ok = s.Get("something else")
	assert.True(t, ok)
	assert.Exactly(t, mr2["something else"], v)

	v, ok = s.Get("something three")
	assert.True(t, ok)
	assert.Exactly(t, mr3["something three"], v)
}

func TestStackGet(t *testing.T) {
	mr1 := &mockGetter{
		"a": "a - tier 1",
		"b": "b - tier 1",
		"c": "c - tier 1",
	}
	mr2 := &mockGetter{
		"b": "b - tier 2",
		"d": "d - tier 2",
	}
	mr3 := &mockGetter{
		"c": "c - tier 3",
		"d": "d - tier 3",
	}
	type kv struct {
		k   string
		v   interface{}
		err error
	}
	patterns := []struct {
		name     string
		gg       []config.Getter
		expected []kv
	}{
		{"one", []config.Getter{mr1}, []kv{
			{"a", "a - tier 1", nil},
			{"b", "b - tier 1", nil},
			{"c", "c - tier 1", nil},
			{"d", nil, config.NotFoundError{Key: "d"}},
			{"e", nil, config.NotFoundError{Key: "e"}},
		}},
		{"two", []config.Getter{mr2, mr1}, []kv{
			{"a", "a - tier 1", nil},
			{"b", "b - tier 2", nil},
			{"c", "c - tier 1", nil},
			{"d", "d - tier 2", nil},
			{"e", nil, config.NotFoundError{Key: "e"}},
		}},
		{"three", []config.Getter{mr3, mr2, mr1}, []kv{
			{"a", "a - tier 1", nil},
			{"b", "b - tier 2", nil},
			{"c", "c - tier 3", nil},
			{"d", "d - tier 3", nil},
			{"e", nil, config.NotFoundError{Key: "e"}},
		}},
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			s := config.NewStack(p.gg...)
			c := config.NewConfig(s)
			for _, x := range p.expected {
				v, err := c.Get(x.k)
				assert.Equal(t, x.err, err, x.k)
				assert.Equal(t, x.v, v.Value(), x.k)
			}
		}
		t.Run(p.name, f)
	}
}

func TestStackInsert(t *testing.T) {
	mr1 := mockGetter{
		"something":        "yet another test string",
		"something else":   "yet another test string",
		"something bottom": "bottom test string",
	}
	mr2 := mockGetter{
		"something":      "another test string",
		"something else": "yet another test string",
	}
	mr3 := mockGetter{
		"something": "a test string",
	}
	s := config.NewStack(&mr1)
	require.NotNil(t, s)
	s.Insert(nil) // should be ignored
	v, ok := s.Get("something")
	assert.True(t, ok)
	assert.Exactly(t, mr1["something"], v)

	// insert a second reader
	s.Insert(&mr2)
	v, ok = s.Get("something")
	assert.True(t, ok)
	assert.Exactly(t, mr2["something"], v)
	v, ok = s.Get("something else")
	assert.True(t, ok)
	assert.Exactly(t, mr2["something else"], v)

	// append a third reader
	s.Insert(&mr3)
	v, ok = s.Get("something")
	assert.True(t, ok)
	assert.Exactly(t, mr3["something"], v)
	v, ok = s.Get("something else")
	assert.True(t, ok)
	assert.Exactly(t, mr2["something else"], v)
	v, ok = s.Get("something bottom")
	assert.True(t, ok)
	assert.Exactly(t, mr1["something bottom"], v)
}

func TestStackNewWatcher(t *testing.T) {
	g1 := mockGetter{
		"a.b.c": 43,
		"a.b.d": 41,
	}
	g2 := mockGetter{
		"a.b.d": 42,
	}
	wg1 := &watchedGetter{g1, nil}
	wg2 := &watchedGetter{g2, nil}
	g := config.NewStack(wg1, wg2)
	require.NotNil(t, g)

	done := make(chan struct{})
	w := g.NewWatcher(done)
	assert.NotNil(t, w)
	ws := wg1.w
	require.NotNil(t, ws)
	assert.True(t, done == ws.donech)

	ws = wg2.w
	require.NotNil(t, ws)
	assert.True(t, done == ws.donech)
}

func TestStackAppendWatch(t *testing.T) {
	g1 := mockGetter{
		"a.b.c": 43,
		"a.b.d": 41,
	}
	g2 := mockGetter{
		"a.b.d": 42,
	}
	patterns := []struct {
		name       string
		overwatch  bool
		underwatch bool
		watchFirst bool
	}{
		{"none wf", false, false, true},
		{"over wf", false, true, true},
		{"under wf", true, false, true},
		{"both wf", true, true, true},
		{"none", false, false, false},
		{"over", false, true, false},
		{"under", true, false, false},
		{"both", true, true, false},
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			ow := &watchedGetter{g1, nil}
			uw := &watchedGetter{g2, nil}
			var o, u config.Getter
			o = &g1
			if p.overwatch {
				o = ow
			}
			u = &g2
			if p.underwatch {
				u = uw
			}
			g := config.NewStack(o)
			require.NotNil(t, g)
			done := make(chan struct{})
			defer close(done)
			var w config.GetterWatcher
			if p.watchFirst {
				w = g.NewWatcher(done)
			}
			g.Append(u)
			if !p.watchFirst {
				w = g.NewWatcher(done)
			}
			require.NotNil(t, w)
			if p.overwatch {
				testDonePropagation(t, o, done)
				testUpdatePropagation(t, w, ow.w)
			}
			if p.underwatch {
				testDonePropagation(t, u, done)
				testUpdatePropagation(t, w, uw.w)
			}
		}
		t.Run(p.name, f)
	}
}

func TestStackInsertWatch(t *testing.T) {
	g1 := mockGetter{
		"a.b.c": 43,
		"a.b.d": 41,
	}
	g2 := mockGetter{
		"a.b.d": 42,
	}
	patterns := []struct {
		name       string
		overwatch  bool
		underwatch bool
		watchFirst bool
	}{
		{"none wf", false, false, true},
		{"over wf", false, true, true},
		{"under wf", true, false, true},
		{"both wf", true, true, true},
		{"none", false, false, false},
		{"over", false, true, false},
		{"under", true, false, false},
		{"both", true, true, false},
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			ow := &watchedGetter{g1, nil}
			uw := &watchedGetter{g2, nil}
			var o, u config.Getter
			o = &g1
			if p.overwatch {
				o = ow
			}
			u = &g2
			if p.underwatch {
				u = uw
			}
			g := config.NewStack(u)
			require.NotNil(t, g)
			done := make(chan struct{})
			defer close(done)
			var w config.GetterWatcher
			if p.watchFirst {
				w = g.NewWatcher(done)
			}
			g.Insert(o)
			if !p.watchFirst {
				w = g.NewWatcher(done)
			}
			require.NotNil(t, w)
			if p.overwatch {
				testDonePropagation(t, o, done)
				testUpdatePropagation(t, w, ow.w)
			}
			if p.underwatch {
				testDonePropagation(t, u, done)
				testUpdatePropagation(t, w, uw.w)
			}
		}
		t.Run(p.name, f)
	}
}

func TestStackWatcherClosedDuringUpdate(t *testing.T) {
	g1 := mockGetter{
		"a.b.c": 43,
		"a.b.d": 41,
	}
	ow := &watchedGetter{g1, nil}
	g := config.NewStack(ow)
	require.NotNil(t, g)
	done := make(chan struct{})
	w := g.NewWatcher(done)
	ow.w.Notify()
	time.Sleep(defaultTimeout)
	close(done)
	testWatcherNotUpdated(t, w)
}

func testDonePropagation(t *testing.T, g config.Getter, done <-chan struct{}) {
	t.Helper()
	wg, ok := g.(*watchedGetter)
	assert.True(t, ok)
	require.NotNil(t, wg)
	require.NotNil(t, wg.w)
	assert.True(t, done == wg.w.donech)
}

func testUpdatePropagation(t *testing.T, w config.GetterWatcher, gw *getterWatcher) {
	t.Helper()
	testWatcherNotUpdated(t, w)
	go gw.Notify()
	ud := testWatcherUpdated(t, w)
	require.NotNil(t, ud)
	gu, ok := ud.(update)
	assert.True(t, ok)
	require.NotNil(t, gu)
	assert.Equal(t, gw, gu.w)
}

func testWatcherUpdated(t *testing.T, w config.GetterWatcher) config.GetterUpdate {
	t.Helper()
	select {
	case update := <-w.Update():
		return update
	case <-time.After(time.Second):
		assert.Fail(t, "watch failed to return")
	}
	return nil
}

func testWatcherNotUpdated(t *testing.T, w config.GetterWatcher) {
	t.Helper()
	select {
	case update := <-w.Update():
		assert.Fail(t, "unexpected update", update)
	case <-time.After(defaultTimeout):
	}
}
