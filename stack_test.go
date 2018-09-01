// Copyright © 2018 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package config_test

import (
	"context"
	"errors"
	"testing"

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
	mr1 := mockGetter{
		"a": "a - tier 1",
		"b": "b - tier 1",
		"c": "c - tier 1",
	}
	mr2 := mockGetter{
		"b": "b - tier 2",
		"d": "d - tier 2",
	}
	mr3 := mockGetter{
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

func TestStackWatcher(t *testing.T) {
	g1 := mockGetter{
		"a.b.c": 43,
		"a.b.d": 41,
	}
	g2 := mockGetter{
		"a.b.d": 42,
	}
	gw1 := &getterWatcher{n: make(chan struct{})}
	gw2 := &getterWatcher{n: make(chan struct{})}
	wg1 := watchedGetter{g1, gw1}
	wg2 := watchedGetter{g2, gw2}
	g := config.NewStack(wg1, wg2)
	require.NotNil(t, g)
	w1, ok := g.Watcher()
	assert.True(t, ok)
	assert.NotNil(t, w1)

	w2, ok := g.Watcher()
	assert.True(t, ok)
	assert.Equal(t, w1, w2)
}

func TestStackWatcherWatch(t *testing.T) {
	g1 := mockGetter{
		"a.b.c": 43,
		"a.b.d": 41,
	}
	g2 := mockGetter{
		"a.b.d": 42,
	}
	gw1 := &getterWatcher{n: make(chan struct{})}
	gw2 := &getterWatcher{n: make(chan struct{})}
	wg1 := watchedGetter{g1, gw1}
	wg2 := watchedGetter{g2, gw2}
	g := config.NewStack(wg1, wg2)
	require.NotNil(t, g)
	w, ok := g.Watcher()
	assert.True(t, ok)
	require.NotNil(t, w)

	testWatcher(t, w, context.DeadlineExceeded)

	gw1.WatchError = config.WithTemporary(errors.New("watch error"))
	gw1.Notify()
	testWatcher(t, w, context.DeadlineExceeded)
	gw1.Reset()
	gw1.WatchError = nil

	gw1.Notify()
	testWatcher(t, w, nil)
	w.CommitUpdate()

	gw2.Notify()
	testWatcher(t, w, nil)
	w.CommitUpdate()

	gw1.Notify()
	testWatcher(t, w, nil)
	w.CommitUpdate()

	gw1.WatchError = errors.New("watch error")
	gw1.Notify()
	testWatcher(t, w, context.DeadlineExceeded)

	testWatcher(t, w, context.DeadlineExceeded)

	// Close after start
	assert.False(t, gw1.Closed)
	assert.False(t, gw2.Closed)
	w.Close()
	assert.True(t, gw1.Closed)
	assert.True(t, gw2.Closed)
}

func TestStackAppendWatch(t *testing.T) {
	g1 := mockGetter{
		"a.b.c": 43,
		"a.b.d": 41,
	}
	g2 := mockGetter{
		"a.b.d": 42,
	}
	gw1 := &getterWatcher{n: make(chan struct{})}
	gw2 := &getterWatcher{n: make(chan struct{})}
	wg1 := watchedGetter{g1, gw1}
	wg2 := watchedGetter{g2, gw2}
	g := config.NewStack(wg1)
	require.NotNil(t, g)
	w, ok := g.Watcher()
	assert.True(t, ok)
	require.NotNil(t, w)

	testWatcher(t, w, context.DeadlineExceeded)

	g.Append(wg2)

	wg3 := watchedGetter{g2, nil}
	g.Append(wg3)

	gw1.WatchError = config.WithTemporary(errors.New("watch error"))
	gw1.Notify()
	testWatcher(t, w, context.DeadlineExceeded)
	gw1.Reset()
	gw1.WatchError = nil

	gw1.Notify()
	testWatcher(t, w, nil)
	w.CommitUpdate()

	gw2.Notify()
	testWatcher(t, w, nil)
	w.CommitUpdate()

	gw1.Notify()
	testWatcher(t, w, nil)
	w.CommitUpdate()

	gw1.WatchError = errors.New("watch error")
	gw1.Notify()
	testWatcher(t, w, context.DeadlineExceeded)

	testWatcher(t, w, context.DeadlineExceeded)

	// Close after start
	assert.False(t, gw1.Closed)
	assert.False(t, gw2.Closed)
	w.Close()
	assert.True(t, gw1.Closed)
	assert.True(t, gw2.Closed)
}

func TestStackInsertWatch(t *testing.T) {
	g1 := mockGetter{
		"a.b.c": 43,
		"a.b.d": 41,
	}
	g2 := mockGetter{
		"a.b.d": 42,
	}
	gw1 := &getterWatcher{n: make(chan struct{})}
	gw2 := &getterWatcher{n: make(chan struct{})}
	wg1 := watchedGetter{g1, gw1}
	wg2 := watchedGetter{g2, gw2}
	g := config.NewStack(wg1)
	require.NotNil(t, g)
	w, ok := g.Watcher()
	assert.True(t, ok)
	require.NotNil(t, w)
	g.Insert(wg2)

	testWatcher(t, w, context.DeadlineExceeded)

	gw1.WatchError = config.WithTemporary(errors.New("watch error"))
	gw1.Notify()
	testWatcher(t, w, context.DeadlineExceeded)
	gw1.Reset()
	gw1.WatchError = nil

	gw1.Notify()
	testWatcher(t, w, nil)
	w.CommitUpdate()

	gw2.Notify()
	testWatcher(t, w, nil)
	w.CommitUpdate()

	gw1.Notify()
	testWatcher(t, w, nil)
	w.CommitUpdate()

	gw1.WatchError = errors.New("watch error")
	gw1.Notify()
	testWatcher(t, w, context.DeadlineExceeded)

	testWatcher(t, w, context.DeadlineExceeded)

	// Close after start
	assert.False(t, gw1.Closed)
	assert.False(t, gw2.Closed)
	w.Close()
	assert.True(t, gw1.Closed)
	assert.True(t, gw2.Closed)
}
