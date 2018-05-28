// Copyright © 2017 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package properties_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/warthog618/config"
	"github.com/warthog618/config/cfgconv"
	"github.com/warthog618/config/properties"
)

var validConfig = []byte(`
bool:true
int:42
float:3.1415
string = this is a string
slice: a:#b
intSlice = 1,2,3,4
stringSlice = one,two,three,four

nested.bool = false
nested.int = 18
nested.float = 3.141
nested.string = this is also a string
nested.intSlice = 1,2,3,4,5,6
nested.stringSlice = one,two,three
`)

var malformedConfig = []byte(`
=malformed
bool: true
`)

// Test that config fields can be read and converted to required types using cfgconv.
func testGetterGet(t *testing.T, g *properties.Getter) {
	bogusKeys := []string{
		"intslice", "stringslice", "bogus",
		"nested", "nested.bogus", "nested.stringslice",
	}
	for _, key := range bogusKeys {
		if v, ok := g.Get(key); ok {
			t.Errorf("could read %s", key)
		} else if v != nil {
			t.Errorf("returned non-nil on failed read for %s, got %v", key, v)
		}
	}
	patterns := []struct {
		k string
		v interface{}
	}{
		{"bool", true},
		{"int", 42},
		{"float", 3.1415},
		{"string", "this is a string"},
		{"intSlice", []interface{}{"1", "2", "3", "4"}},
		{"stringSlice", []interface{}{"one", "two", "three", "four"}},
		{"nested.bool", false},
		{"nested.int", 18},
		{"nested.float", 3.141},
		{"nested.string", "this is also a string"},
		{"nested.intSlice", []interface{}{"1", "2", "3", "4", "5", "6"}},
		{"nested.stringSlice", []interface{}{"one", "two", "three"}},
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			v, ok := g.Get(p.k)
			assert.True(t, ok)
			var cv interface{}
			var err error
			switch p.v.(type) {
			case bool:
				cv, err = cfgconv.Bool(v)
			case int:
				cv, err = cfgconv.Int(v)
			case float64:
				cv, err = cfgconv.Float(v)
			case string:
				cv, err = cfgconv.String(v)
			case []interface{}:
				cv, err = cfgconv.Slice(v)
			default:
				assert.Fail(t, "unsupported value type")
			}
			assert.Nil(t, err)
			assert.EqualValues(t, p.v, cv)
		}
		t.Run(p.k, f)
	}
}

func TestGetterWithListSeparator(t *testing.T) {
	patterns := []struct {
		name     string
		sep      string
		expected interface{}
	}{
		{"default", ":", []string{"a", "#b"}},
		{"multi", ":#", []string{"a", "b"}},
		{"none", "", "a:#b"},
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			r, err := properties.NewBytes(validConfig, properties.WithListSeparator(p.sep))
			assert.Nil(t, err)
			require.NotNil(t, r)
			v, ok := r.Get("slice")
			assert.True(t, ok)
			assert.Equal(t, p.expected, v)
		}
		t.Run(p.name, f)
	}
}

func TestNewBytes(t *testing.T) {
	b, err := properties.NewBytes(malformedConfig)
	assert.NotNil(t, err)
	assert.Nil(t, b)
	b, err = properties.NewBytes(validConfig)
	assert.Nil(t, err)
	require.NotNil(t, b)
	// test provides config.Getter interface.
	cfg := config.New()
	cfg.AppendGetter(b)
}

func TestNewFile(t *testing.T) {
	f, err := properties.NewFile("no_such.properties")
	assert.NotNil(t, err)
	assert.Nil(t, f)
	f, err = properties.NewFile("malformed.properties")
	assert.NotNil(t, err)
	assert.Nil(t, f)
	f, err = properties.NewFile("config.properties")
	assert.Nil(t, err)
	require.NotNil(t, f)
	// test provides config.Getter interface.
	cfg := config.New()
	cfg.AppendGetter(f)
}

func TestStringGetterGet(t *testing.T) {
	g, err := properties.NewBytes(validConfig)
	assert.Nil(t, err)
	require.NotNil(t, g)
	testGetterGet(t, g)
}

func TestFileGetterGet(t *testing.T) {
	g, err := properties.NewFile("config.properties")
	assert.Nil(t, err)
	require.NotNil(t, g)
	testGetterGet(t, g)
}
