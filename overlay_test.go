// Copyright © 2018 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/warthog618/config"
)

func TestOverlay(t *testing.T) {
	under := &mockGetter{
		"a.b.c": 43,
		"a.b.d": 41,
	}
	over := &mockGetter{
		"a.b.d": 42,
	}
	g := config.Overlay(over, under)
	require.NotNil(t, g)
}

func TestOverlayGet(t *testing.T) {
	under := &mockGetter{
		"a.b.c": 43,
		"a.b.d": 41,
	}
	over := &mockGetter{
		"a.b.d": 42,
	}
	g := config.Overlay(over, under)
	require.NotNil(t, g)

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

func TestOverlayNewWatcher(t *testing.T) {
	under := mockGetter{
		"a.b.c": 43,
		"a.b.d": 41,
	}
	over := mockGetter{
		"a.b.d": 42,
	}
	patterns := []struct {
		name       string
		overwatch  bool
		underwatch bool
		watchable  bool
	}{
		{"none", false, false, false},
		{"under", false, true, true},
		{"over", true, false, true},
		{"both", true, true, true},
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			ow := &watchedGetter{over, nil}
			uw := &watchedGetter{under, nil}
			var o, u config.Getter
			o = &over
			if p.overwatch {
				o = ow
			}
			u = &under
			if p.underwatch {
				u = uw
			}
			g := config.Overlay(o, u)
			require.NotNil(t, g)
			wg, ok := g.(config.WatchableGetter)
			assert.True(t, ok)
			require.NotNil(t, wg)
			done := make(chan struct{})
			defer close(done)
			w := wg.NewWatcher(done)
			if !p.watchable {
				assert.Nil(t, w)
				return
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
