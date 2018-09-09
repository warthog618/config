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
	"github.com/warthog618/config/keys"
)

func TestDecorate(t *testing.T) {
	eg := echoGetter{}
	patterns := []struct {
		name     string
		k        string
		d        []config.Decorator
		expected string
	}{
		{"none", "a.b.c.d", nil, "a.b.c.d"},
		{"one", "A.B.C.D",
			[]config.Decorator{
				config.WithKeyReplacer(keys.LowerCaseReplacer()),
			},
			"a.b.c.d",
		},
		{"two", "C.D",
			[]config.Decorator{
				config.WithKeyReplacer(keys.LowerCaseReplacer()),
				config.WithPrefix("A.B."),
			},
			"A.B.c.d",
		},
		{"two reversed", "C.D",
			[]config.Decorator{
				config.WithPrefix("a.b."),
				config.WithKeyReplacer(keys.LowerCaseReplacer()),
			},
			"a.b.c.d",
		},
		{"three", "foo",
			[]config.Decorator{
				config.WithPrefix("a.B."),
				config.WithKeyReplacer(keys.LowerCaseReplacer()),
				config.WithKeyReplacer(keys.StringReplacer("foo", "C.D")),
			},
			"a.b.C.D",
		},
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			pr := config.Decorate(&eg, p.d...)
			v, ok := pr.Get(p.k)
			assert.Equal(t, true, ok, p.k)
			assert.Equal(t, p.expected, v)
		}
		t.Run(p.name, f)
	}
}

func TestWithDefault(t *testing.T) {
	def := mockGetter{
		"a.b.c": 43,
		"a.b.d": 41,
	}
	nondef := mockGetter{
		"a.b.d": 42,
	}
	g := config.WithDefault(&def)(&nondef)

	// defaulted
	c, ok := g.Get("a.b.c")
	assert.True(t, ok)
	assert.Equal(t, 43, c)

	// non-default
	c, ok = g.Get("a.b.d")
	assert.True(t, ok)
	assert.Equal(t, 42, c)

	// neither
	c, ok = g.Get("a.b.e")
	assert.False(t, ok)
	assert.Nil(t, c)

	// nil default
	g = config.WithDefault(nil)(&nondef)
	assert.Equal(t, &nondef, g)

	// no longer defaulted
	c, ok = g.Get("a.b.c")
	assert.False(t, ok)
	assert.Nil(t, c)

	// non-default
	c, ok = g.Get("a.b.d")
	assert.True(t, ok)
	assert.Equal(t, 42, c)

	testDecoratorWatchable(t, config.WithDefault(&def))
}

func TestWithGraft(t *testing.T) {
	mg := mockGetter{
		"a":     "is a",
		"foo.b": "is foo.b",
	}
	patterns := []struct {
		name string
		k    string
		ok   bool
		v    interface{}
	}{
		{"level 1", "blah.a", true, "is a"},
		{"level 2", "blah.foo.b", true, "is foo.b"},
		{"not level 1", "notblah.a", false, nil},
		{"not level 2", "notblah.foo.a", false, nil},
		{"empty", "", false, nil},
		{"level 0", "a", false, nil},
	}
	pr := config.WithGraft("blah.")(&mg)
	for _, p := range patterns {
		f := func(t *testing.T) {
			v, ok := pr.Get(p.k)
			assert.Equal(t, p.ok, ok)
			assert.Equal(t, p.v, v)
		}
		t.Run(p.name, f)
	}
	testDecoratorWatchable(t, config.WithGraft("blah."))
}
func TestWithKeyReplacer(t *testing.T) {
	patterns := []struct {
		name     string
		k        string
		d        keys.Replacer
		expected string
	}{
		{"nil", "a.b.c.d", nil, "a.b.c.d"},
		{"lower", "A.B.C.D", keys.LowerCaseReplacer(), "a.b.c.d"},
		{"upper", "a.b.C.d", keys.UpperCaseReplacer(), "A.B.C.D"},
		{"string", "a.b.foo", keys.StringReplacer("foo", "C.D"), "a.b.C.D"},
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			pr := config.WithKeyReplacer(p.d)(echoGetter{})
			v, ok := pr.Get(p.k)
			assert.Equal(t, true, ok, p.k)
			assert.Equal(t, p.expected, v)
		}
		t.Run(p.name, f)
	}
	testDecoratorWatchable(t, config.WithKeyReplacer(keys.LowerCaseReplacer()))
}

func TestWithMustGet(t *testing.T) {
	mg := mockGetter{"a": "is a"}
	pr := config.WithMustGet()(&mg)
	v, ok := pr.Get("a")
	assert.True(t, true, ok)
	assert.Equal(t, "is a", v)
	assert.Panics(t, func() {
		pr.Get("nosuch")
	})
	testDecoratorWatchable(t, config.WithMustGet())
}

func TestWithPrefix(t *testing.T) {
	patterns := []struct {
		name   string
		k      string
		prefix string
	}{
		{"none", "key", ""},
		{"one", "b.c", "a"},
		{"two", "c.d", "a.b"},
		{"three", "d", "a.b.c"},
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			pr := config.WithPrefix(p.prefix)(echoGetter{})
			v, ok := pr.Get(p.k)
			assert.Equal(t, true, ok, p.k)
			assert.Equal(t, p.prefix+p.k, v)
		}
		t.Run(p.name, f)
	}
	testDecoratorWatchable(t, config.WithPrefix("any prefix"))
}

func TestWithUpdateHandler(t *testing.T) {
	mg := mockGetter{
		"a":     "a",
		"a.b":   "a.b",
		"a.b.c": "a.b.c",
	}
	patterns := []struct {
		name string
		k    string
		v    string
	}{
		{"one", "a", "a"},
		{"two", "a.b", "a.b"},
		{"three", "a.b.c", "a.b.c"},
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			// unwatchable
			pr := config.WithUpdateHandler(nil)(echoGetter{})
			v, ok := pr.Get(p.k)
			assert.Equal(t, true, ok, p.k)
			assert.Equal(t, p.v, v)
			// watchable
			mgw := watchedGetter{mg, nil}
			pr = config.WithUpdateHandler(nil)(&mgw)
			v, ok = pr.Get(p.k)
			assert.Equal(t, true, ok, p.k)
			assert.Equal(t, p.v, v)
		}
		t.Run(p.name, f)
	}
	passthru := func(done <-chan struct{}, in <-chan config.GetterUpdate, out chan<- config.GetterUpdate) {
		for {
			select {
			case <-done:
				return
			case u, ok := <-in:
				if !ok {
					return
				}
				select {
				case <-done:
					return
				case out <- u:
				}
			}
		}
	}
	testDecoratorWatchable(t, config.WithUpdateHandler(config.UpdateHandler(passthru)))
	dropper := func(done <-chan struct{}, in <-chan config.GetterUpdate, out chan<- config.GetterUpdate) {
		<-done
	}
	testDecoratorNoUpdate(t, config.WithUpdateHandler(config.UpdateHandler(dropper)))

}

func testDecoratorWatchable(t *testing.T, d config.Decorator) {
	t.Helper()
	mg := mockGetter{}
	mgw := watchedGetter{mg, nil}
	g := d(&mgw)
	wg, ok := g.(config.WatchableGetter)
	assert.True(t, ok)
	require.NotNil(t, wg)
	done := make(chan struct{})
	w := wg.NewWatcher(done)
	assert.NotNil(t, w)
	ws := mgw.w
	require.NotNil(t, ws)
	assert.True(t, done == ws.donech)
	go mgw.w.Notify()
	select {
	case <-w.Update():
	case <-time.After(defaultTimeout):
		assert.Fail(t, "failed to propagate update")
	}
}

func testDecoratorNoUpdate(t *testing.T, d config.Decorator) {
	t.Helper()
	mg := mockGetter{}
	mgw := watchedGetter{mg, nil}
	g := d(&mgw)
	wg, ok := g.(config.WatchableGetter)
	assert.True(t, ok)
	require.NotNil(t, wg)
	done := make(chan struct{})
	w := wg.NewWatcher(done)
	assert.NotNil(t, w)
	ws := mgw.w
	require.NotNil(t, ws)
	assert.True(t, done == ws.donech)
	go mgw.w.Notify()
	select {
	case u := <-w.Update():
		assert.Fail(t, "unexpected update", "%#v", u)
	case <-time.After(defaultTimeout):
	}
}

type echoGetter struct{}

func (e echoGetter) Get(key string) (interface{}, bool) {
	return key, true
}
