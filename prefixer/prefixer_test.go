// Copyright © 2017 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package prefixer_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/warthog618/config"
	"github.com/warthog618/config/prefixer"
)

func TestNew(t *testing.T) {
	m := mockGetter{map[string]string{}}
	p := prefixer.New("blah.", &m)
	if p == nil {
		t.Fatalf("new returned nil")
	}
	// test provides config.Getter interface.
	cfg := config.New()
	cfg.AppendGetter(p)
}

func TestGet(t *testing.T) {
	m := mockGetter{map[string]string{"a": "is a", "foo.b": "is foo.b"}}
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
	pr := prefixer.New("blah.", &m)
	for _, p := range patterns {
		f := func(t *testing.T) {
			v, ok := pr.Get(p.k)
			assert.Equal(t, p.ok, ok)
			assert.Equal(t, p.v, v)
		}
		t.Run(p.k, f)
	}
}

// A simple mock Getter wrapping an accessible map.

type mockGetter struct {
	Config map[string]string
}

func (m *mockGetter) Get(key string) (interface{}, bool) {
	v, ok := m.Config[key]
	return v, ok
}
