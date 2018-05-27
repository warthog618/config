// Copyright © 2017 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package yaml_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/warthog618/config"
	"github.com/warthog618/config/cfgconv"
	"github.com/warthog618/config/yaml"
)

var validConfig = []byte(`
bool: true
int: 42
float: 3.1415
string: this is a string
intSlice: [1,2,3,4]
stringSlice: [one,two,three,four]
nested:
  bool: false
  int: 18
  float: 3.141
  string: this is also a string
  intSlice: [1,2,3,4,5,6]
  stringSlice: [one,two,three]
`)

var malformedConfig = []byte(`
malformed
bool: true
int: 42
float: 3.1415
`)

// Test that config fields can be read and converted to required types using cfgconv.
func testReaderRead(t *testing.T, reader *yaml.Reader) {
	bogusKeys := []string{
		"intslice", "stringslice", "bogus",
		"nested", "nested.bogus", "nested.stringslice",
	}
	for _, key := range bogusKeys {
		if v, ok := reader.Read(key); ok {
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
		{"intSlice", []interface{}{int(1), int(2), int(3), int(4)}},
		{"stringSlice", []interface{}{"one", "two", "three", "four"}},
		{"nested.bool", false},
		{"nested.int", 18},
		{"nested.float", 3.141},
		{"nested.string", "this is also a string"},
		{"nested.intSlice", []interface{}{int(1), int(2), int(3), int(4), int(5), int(6)}},
		{"nested.stringSlice", []interface{}{"one", "two", "three"}},
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			v, ok := reader.Read(p.k)
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

func TestNewBytes(t *testing.T) {
	b, err := yaml.NewBytes(malformedConfig)
	assert.NotNil(t, err)
	assert.Nil(t, b)
	b, err = yaml.NewBytes(validConfig)
	assert.Nil(t, err)
	require.NotNil(t, b)
	// test provides config.Reader interface.
	cfg := config.New()
	cfg.AppendReader(b)
}

func TestNewFile(t *testing.T) {
	f, err := yaml.NewFile("no_such.yaml")
	assert.NotNil(t, err)
	assert.Nil(t, f)
	f, err = yaml.NewFile("malformed.yaml")
	assert.NotNil(t, err)
	assert.Nil(t, f)
	f, err = yaml.NewFile("config.yaml")
	assert.Nil(t, err)
	require.NotNil(t, f)
	cfg := config.New()
	cfg.AppendReader(f)
}

func TestBytesReaderRead(t *testing.T) {
	reader, err := yaml.NewBytes(validConfig)
	assert.Nil(t, err)
	require.NotNil(t, reader)
	testReaderRead(t, reader)
}

func TestFileReaderRead(t *testing.T) {
	reader, err := yaml.NewFile("config.yaml")
	assert.Nil(t, err)
	require.NotNil(t, reader)
	testReaderRead(t, reader)
}
