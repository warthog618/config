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

func TestOverlay(t *testing.T) {
	under := mockGetter{
		"a.b.c": 43,
		"a.b.d": 41,
	}
	over := mockGetter{
		"a.b.d": 42,
	}
	g := config.Overlay(over, under)

	// under
	c, ok := g.Get("a.b.c")
	assert.True(t, ok)
	assert.Equal(t, 43, c)

	// shadowed by over
	c, ok = g.Get("a.b.d")
	assert.True(t, ok)
	assert.Equal(t, 42, c)

	// neither
	c, ok = g.Get("a.b.e")
	assert.False(t, ok)
	assert.Nil(t, c)

	// singular
	g = config.Overlay(over)
	assert.Equal(t, over, g)
}

func TestOverlayWatcherClose(t *testing.T) {
	g1 := mockGetter{
		"a.b.c": 43,
		"a.b.d": 41,
	}
	g2 := mockGetter{
		"a.b.d": 42,
	}
	gw1 := NewGetterWatcher()
	defer gw1.Close()
	gw2 := NewGetterWatcher()
	defer gw2.Close()
	wg1 := watchedGetter{g1, gw1}
	wg2 := watchedGetter{g2, gw2}
	g := config.Overlay(wg1, wg2)
	require.NotNil(t, g)
	assert.False(t, gw1.Closed)
	assert.False(t, gw2.Closed)
	w, ok := g.(config.WatchableGetter).Watcher()
	assert.True(t, ok)
	require.NotNil(t, w)
	w.Close()
	assert.True(t, gw1.Closed)
	assert.True(t, gw2.Closed)
}

func TestOverlayWatcherWatch(t *testing.T) {
	g1 := mockGetter{
		"a.b.c": 43,
		"a.b.d": 41,
	}
	g2 := mockGetter{
		"a.b.d": 42,
	}
	gw1 := NewGetterWatcher()
	defer gw1.Close()
	gw2 := NewGetterWatcher()
	defer gw2.Close()
	wg1 := watchedGetter{g1, gw1}
	wg2 := watchedGetter{g2, gw2}
	g := config.Overlay(wg1, wg2)
	require.NotNil(t, g)
	w, ok := g.(config.WatchableGetter).Watcher()
	assert.True(t, ok)
	require.NotNil(t, w)

	testWatcher(t, w, nil, context.DeadlineExceeded)

	gw1.WatchError(config.WithTemporary(errors.New("watch error")))
	testWatcher(t, w, gw1.Notify, context.DeadlineExceeded)
	gw1.WatchError(nil)

	testWatcher(t, w, gw1.Notify, nil)
	w.CommitUpdate()

	testWatcher(t, w, gw2.Notify, nil)
	w.CommitUpdate()

	testWatcher(t, w, gw1.Notify, nil)
	w.CommitUpdate()

	gw1.WatchError(errors.New("watch error"))
	testWatcher(t, w, gw1.Notify, context.DeadlineExceeded)

	testWatcher(t, w, nil, context.DeadlineExceeded)

	// Close after start
	assert.False(t, gw1.Closed)
	assert.False(t, gw2.Closed)
	w.Close()
	assert.True(t, gw1.Closed)
	assert.True(t, gw2.Closed)
}
