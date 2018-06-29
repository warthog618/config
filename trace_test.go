// Copyright © 2018 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/warthog618/config"
)

func TestWithTrace(t *testing.T) {
	mr := mockGetter{
		"foo.b": "foo.b",
	}
	type alias struct {
		new string
		old string
	}
	patterns := []struct {
		name string
		tp   string
	}{
		{"ok", "foo.b"},
		{"missing", "b"},
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			tf := func(key string, v interface{}, ok bool) {
				assert.Equal(t, p.tp, key)
				tv, tok := mr[key]
				assert.Equal(t, tv, v)
				assert.Equal(t, tok, ok)
			}
			c := config.WithTrace(tf)(mr)
			v, ok := c.Get(p.tp)
			tv, tok := mr[p.tp]
			assert.Equal(t, tv, v)
			assert.Equal(t, tok, ok)
		}
		t.Run(p.name, f)
	}
}
