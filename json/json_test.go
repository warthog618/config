// Copyright © 2017 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package json_test

import (
	gojson "encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/warthog618/config"
	"github.com/warthog618/config/cfgconv"
	"github.com/warthog618/config/json"
)

func TestNewBytes(t *testing.T) {
	patterns := []struct {
		name    string
		in      []byte
		errType interface{}
	}{
		{"malformed", malformedConfig, &gojson.SyntaxError{Offset: 0}},
		{"valid", validConfig, nil},
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			b, err := json.NewBytes(p.in)
			assert.IsType(t, p.errType, err)
			if err == nil {
				require.NotNil(t, b)
				// test b provides config.Getter interface.
				cfg := config.New()
				cfg.AppendGetter(b)
			}
		}
		t.Run(p.name, f)
	}
}

func TestNewFile(t *testing.T) {
	patterns := []struct {
		name    string
		in      string
		errType interface{}
	}{
		{"no such", "no_such.json", &os.PathError{}},
		{"malformed", "malformed.json", &gojson.SyntaxError{Offset: 0}},
		{"valid", "config.json", nil},
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			b, err := json.NewFile(p.in)
			assert.IsType(t, p.errType, err)
			if err == nil {
				require.NotNil(t, b)
				// test b provides config.Getter interface.
				cfg := config.New()
				cfg.AppendGetter(b)
			}
		}
		t.Run(p.name, f)
	}
}

func TestBytesGetterGet(t *testing.T) {
	g, err := json.NewBytes(validConfig)
	if err != nil {
		t.Fatalf("failed to parse config")
	}
	testGetterGet(t, g)
}

func TestFileGetterGet(t *testing.T) {
	g, err := json.NewFile("config.json")
	if err != nil {
		t.Fatalf("failed to parse config")
	}
	testGetterGet(t, g)
}

var validConfig = []byte(`{
	"bool": true,
	"int": 42,
	"float": 3.1415,
	"string": "this is a string",
	"intSlice": [1,2,3,4],
	"stringSlice": ["one","two","three","four"],
	"nested": {
	  "bool": false,
	  "int": 18,
	  "float": 3.141,
	  "string": "this is also a string",
	  "intSlice": [1,2,3,4,5,6],
	  "stringSlice": ["one","two","three"]
	}
  }`)

var malformedConfig = []byte(`malformed{
	"bool": true,
	"int": 42,
	"float": 3.1415
  }`)

// Test that config fields can be read and converted to required types using cfgconv.
func testGetterGet(t *testing.T, g *json.Getter) {
	bogusKeys := []string{
		"intslice", "stringslice", "bogus",
		"nested", "nested.bogus", "nested.stringslice",
	}
	for _, key := range bogusKeys {
		v, ok := g.Get(key)
		assert.False(t, ok)
		assert.Nil(t, v)
	}
	patterns := []struct {
		k string
		v interface{}
	}{
		{"bool", true},
		{"int", 42},
		{"float", 3.1415},
		{"string", "this is a string"},
		{"intSlice", []interface{}{float64(1), float64(2), float64(3), float64(4)}},
		{"stringSlice", []interface{}{"one", "two", "three", "four"}},
		{"nested.bool", false},
		{"nested.int", 18},
		{"nested.float", 3.141},
		{"nested.string", "this is also a string"},
		{"nested.intSlice", []interface{}{float64(1), float64(2), float64(3), float64(4), float64(5), float64(6)}},
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

var benchConfig = []byte(`{
	"bool": true,
	"int": 42,
	"float": 3.1415,
	"string": "this is a string",
	"intSlice": [1,2,3,4],
	"stringSlice": ["one","two","three","four"],
	"nested": {
		"bool": false,
		"int": 18,
		"float": 3.141,
		"leaf":44,
		"string": "this is also a string",
		"intSlice": [1,2,3,4,5,6],
		"stringSlice": ["one","two","three"]
	}
  }`)

func BenchmarkGet(b *testing.B) {
	g, _ := json.NewBytes(benchConfig)
	for n := 0; n < b.N; n++ {
		g.Get("nested.leaf")
	}
}
