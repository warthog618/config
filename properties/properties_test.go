// Copyright © 2017 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package properties_test

import (
	"bytes"
	"errors"
	"io"
	"os"
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
	t.Helper()
	bogusKeys := []string{
		"intslice", "stringslice", "stringSlice[4]", "bogus",
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
		{"intSlice[]", 4},
		{"intSlice[2]", float64(3)},
		{"stringSlice", []interface{}{"one", "two", "three", "four"}},
		{"stringSlice[]", 4},
		{"stringSlice[2]", "three"},
		{"nested.bool", false},
		{"nested.int", 18},
		{"nested.float", 3.141},
		{"nested.string", "this is also a string"},
		{"nested.intSlice", []interface{}{"1", "2", "3", "4", "5", "6"}},
		{"nested.intSlice[]", 6},
		{"nested.intSlice[3]", float64(4)},
		{"nested.stringSlice", []interface{}{"one", "two", "three"}},
		{"nested.stringSlice[]", 3},
		{"nested.stringSlice[0]", "one"},
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
			r, err := properties.New(
				properties.FromBytes(validConfig),
				properties.WithListSeparator(p.sep))
			assert.Nil(t, err)
			require.NotNil(t, r)
			v, ok := r.Get("slice")
			assert.True(t, ok)
			assert.Equal(t, p.expected, v)
		}
		t.Run(p.name, f)
	}
}

func TestNewFromBytes(t *testing.T) {
	patterns := []struct {
		name    string
		in      []byte
		errType interface{}
	}{
		{"malformed", malformedConfig, errors.New("")},
		{"valid", validConfig, nil},
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			b, err := properties.New(properties.FromBytes(p.in))
			assert.IsType(t, p.errType, err)
			if err == nil {
				require.NotNil(t, b)
				assert.Implements(t, (*config.Getter)(nil), b)
			}
		}
		t.Run(p.name, f)
	}
}

func TestNewFromFile(t *testing.T) {
	patterns := []struct {
		name    string
		in      string
		errType interface{}
	}{
		{"no such", "no_such.properties", &os.PathError{}},
		{"malformed", "malformed.properties", errors.New("")},
		{"valid", "config.properties", nil},
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			b, err := properties.New(properties.FromFile(p.in))
			assert.IsType(t, p.errType, err)
			if err == nil {
				require.NotNil(t, b)
				assert.Implements(t, (*config.Getter)(nil), b)
			}
		}
		t.Run(p.name, f)
	}
}

func TestNewFromReader(t *testing.T) {
	patterns := []struct {
		name    string
		in      io.Reader
		errType interface{}
	}{
		{"failed", failReader(0), errors.New("")},
		{"malformed", bytes.NewReader(malformedConfig), errors.New("")},
		{"valid", bytes.NewReader(validConfig), nil},
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			b, err := properties.New(properties.FromReader(p.in))
			assert.IsType(t, p.errType, err)
			if err == nil {
				require.NotNil(t, b)
				assert.Implements(t, (*config.Getter)(nil), b)
			}
		}
		t.Run(p.name, f)
	}
}

type failReader int

func (r failReader) Read(b []byte) (n int, err error) {
	return 0, errors.New("read failed")
}

func TestStringGetterGet(t *testing.T) {
	g, err := properties.New(properties.FromBytes(validConfig))
	assert.Nil(t, err)
	require.NotNil(t, g)
	testGetterGet(t, g)
}

func TestFileGetterGet(t *testing.T) {
	g, err := properties.New(properties.FromFile("config.properties"))
	assert.Nil(t, err)
	require.NotNil(t, g)
	testGetterGet(t, g)
}

var benchConfig = []byte(`
bool:true
int:42
float:3.1415
string = this is a string
slice: a:#b
intSlice = 1,2,3,4
stringSlice = one,two,three,four

nested.leaf: 44

nested.bool = false
nested.int = 18
nested.float = 3.141
nested.string = this is also a string
nested.intSlice = 1,2,3,4,5,6
nested.stringSlice = one,two,three
`)

func BenchmarkGet(b *testing.B) {
	g, _ := properties.New(properties.FromBytes(benchConfig))
	for n := 0; n < b.N; n++ {
		g.Get("int")
	}
}

func BenchmarkGetNested(b *testing.B) {
	g, _ := properties.New(properties.FromBytes(benchConfig))
	for n := 0; n < b.N; n++ {
		g.Get("nested.leaf")
	}
}

func BenchmarkGetArray(b *testing.B) {
	g, _ := properties.New(properties.FromBytes(benchConfig))
	for n := 0; n < b.N; n++ {
		g.Get("slice")
	}
}

func BenchmarkGetArrayLen(b *testing.B) {
	g, _ := properties.New(properties.FromBytes(benchConfig))
	for n := 0; n < b.N; n++ {
		g.Get("slice[]")
	}
}

func BenchmarkGetArrayElement(b *testing.B) {
	g, _ := properties.New(properties.FromBytes(benchConfig))
	for n := 0; n < b.N; n++ {
		g.Get("slice[1]")
	}
}
