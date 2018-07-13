// Copyright © 2017 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package toml_test

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
	"github.com/warthog618/config/toml"
)

var validConfig = []byte(`
bool = true
int = 42
float=  3.1415
string = "this is a string"
intSlice = [1,2,3,4]
stringSlice = ["one","two","three","four"]
sliceslice = [[1,2,3,4],[5,6,7,8]]

[nested]
bool = false
int = 18
float = 3.141
string = "this is also a string"
intSlice = [1,2,3,4,5,6]
stringSlice = ["one","two","three"]

[[animals]]
name = "Platypus"
order= "Monotremata"

[[animals]]
name = "Quoll"
order = "Dasyuromorphia"

`)

var malformedConfig = []byte(`
malformed
bool: true
`)

// Test that config fields can be read and converted to required types using cfgconv.
func testGetterGet(t *testing.T, g *toml.Getter) {
	t.Helper()
	bogusKeys := []string{
		"intslice", "stringslice", "bogus",
		"nested", "nested.bogus", "nested.stringslice",
		"animals[0]",
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
		{"intSlice", []interface{}{int64(1), int64(2), int64(3), int64(4)}},
		{"intSlice[]", 4},
		{"intSlice[2]", float64(3)},
		{"stringSlice", []interface{}{"one", "two", "three", "four"}},
		{"stringSlice[]", 4},
		{"stringSlice[2]", "three"},
		{"sliceslice", []interface{}{
			[]interface{}{int64(1), int64(2), int64(3), int64(4)},
			[]interface{}{int64(5), int64(6), int64(7), int64(8)}}},
		{"sliceslice[]", 2},
		{"sliceslice[0][]", 4},
		{"sliceslice[1]", []interface{}{int64(5), int64(6), int64(7), int64(8)}},
		{"sliceslice[1][2]", 7},
		{"nested.bool", false},
		{"nested.int", 18},
		{"nested.float", 3.141},
		{"nested.string", "this is also a string"},
		{"nested.intSlice", []interface{}{int64(1), int64(2), int64(3), int64(4), int64(5), int64(6)}},
		{"nested.stringSlice", []interface{}{"one", "two", "three"}},
		{"nested.stringSlice[]", 3},
		{"nested.stringSlice[0]", "one"},
		{"animals", []interface{}{nil, nil}},
		{"animals[]", 2},
		{"animals[0].name", "Platypus"},
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

func TestNew(t *testing.T) {
	b, err := toml.New()
	assert.Nil(t, err)
	require.NotNil(t, b)
	v, ok := b.Get("bogus")
	assert.False(t, ok)
	assert.Nil(t, v)
	assert.Implements(t, (*config.Getter)(nil), b)
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
			b, err := toml.New(toml.FromBytes(p.in))
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
		{"no such", "no_such.toml", &os.PathError{}},
		{"malformed", "malformed.toml", errors.New("")},
		{"valid", "config.toml", nil},
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			b, err := toml.New(toml.FromFile(p.in))
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
			b, err := toml.New(toml.FromReader(p.in))
			assert.IsType(t, p.errType, err)
			if err == nil {
				require.NotNil(t, b)
			}
		}
		t.Run(p.name, f)
	}
}

type failReader int

func (r failReader) Read(b []byte) (n int, err error) {
	return 0, errors.New("read failed")
}

func TestBytesGetterGet(t *testing.T) {
	g, err := toml.New(toml.FromBytes(validConfig))
	assert.Nil(t, err)
	require.NotNil(t, g)
	testGetterGet(t, g)
}

func TestFileGetterGet(t *testing.T) {
	g, err := toml.New(toml.FromFile("config.toml"))
	assert.Nil(t, err)
	require.NotNil(t, g)
	testGetterGet(t, g)
}

var benchConfig = []byte(`
bool = true
int = 42
float=  3.1415
string = "this is a string"
intSlice = [1,2,3,4]
stringSlice = ["one","two","three","four"]
[nested]
bool = false
int = 18
float = 3.141
leaf = 44
string = "this is also a string"
intSlice = [1,2,3,4,5,6]
stringSlice = ["one","two","three"]
	`)

func BenchmarkNew(b *testing.B) {
	for n := 0; n < b.N; n++ {
		toml.New(toml.FromBytes(benchConfig))
	}
}

func BenchmarkGet(b *testing.B) {
	b.StopTimer()
	g, _ := toml.New(toml.FromBytes(benchConfig))
	b.StartTimer()
	for n := 0; n < b.N; n++ {
		g.Get("string")
	}
}

func BenchmarkGetNested(b *testing.B) {
	b.StopTimer()
	g, _ := toml.New(toml.FromBytes(benchConfig))
	b.StartTimer()
	for n := 0; n < b.N; n++ {
		g.Get("nested.leaf")
	}
}

func BenchmarkGetArray(b *testing.B) {
	b.StopTimer()
	g, _ := toml.New(toml.FromBytes(benchConfig))
	b.StartTimer()
	for n := 0; n < b.N; n++ {
		g.Get("intSlice")
	}
}

func BenchmarkGetArrayLen(b *testing.B) {
	b.StopTimer()
	g, _ := toml.New(toml.FromBytes(benchConfig))
	b.StartTimer()
	for n := 0; n < b.N; n++ {
		g.Get("intSlice[]")
	}
}

func BenchmarkGetArrayElement(b *testing.B) {
	b.StopTimer()
	g, _ := toml.New(toml.FromBytes(benchConfig))
	b.StartTimer()
	for n := 0; n < b.N; n++ {
		g.Get("intSlice[2]")
	}
}
