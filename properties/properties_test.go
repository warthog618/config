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
func testReaderRead(t *testing.T, reader *properties.Reader) {
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

func TestReaderSetListSeparator(t *testing.T) {
	r, err := properties.NewBytes(validConfig)
	assert.Nil(t, err)
	require.NotNil(t, r)
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
			r.SetListSeparator(p.sep)
			v, ok := r.Read("slice")
			assert.True(t, ok)
			assert.Equal(t, p.expected, v)
		}
		t.Run(p.name, f)
	}

	if err != nil {
		t.Fatalf("failed to parse validConfig")
	}
}

func TestNewBytes(t *testing.T) {
	b, err := properties.NewBytes(malformedConfig)
	assert.NotNil(t, err)
	assert.Nil(t, b)
	b, err = properties.NewBytes(validConfig)
	assert.Nil(t, err)
	require.NotNil(t, b)
	// test provides config.Reader interface.
	cfg := config.New()
	cfg.AppendReader(b)
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
	// test provides config.Reader interface.
	cfg := config.New()
	cfg.AppendReader(f)
}

func TestStringReaderRead(t *testing.T) {
	reader, err := properties.NewBytes(validConfig)
	assert.Nil(t, err)
	require.NotNil(t, reader)
	testReaderRead(t, reader)
}

func TestFileReaderRead(t *testing.T) {
	reader, err := properties.NewFile("config.properties")
	assert.Nil(t, err)
	require.NotNil(t, reader)
	testReaderRead(t, reader)
}
