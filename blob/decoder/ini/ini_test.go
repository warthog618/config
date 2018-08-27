// Copyright © 2018 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package ini_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/warthog618/config/blob/decoder/ini"
)

func TestNewDecoder(t *testing.T) {
	d := ini.NewDecoder()
	require.NotNil(t, d)
}

func TestDecode(t *testing.T) {
	d := ini.NewDecoder()
	require.NotNil(t, d)
	m := make(map[string]interface{})
	err := d.Decode(malformedConfig, &m)
	assert.NotNil(t, err)
	assert.Equal(t, map[string]interface{}{}, m)
	err = d.Decode(validConfig, &m)
	assert.Nil(t, err)
	assert.Equal(t, parsedConfig, m)
	err = d.Decode(validConfig, 3)
	assert.Equal(t, "Decode only supports map[string]interface{}", err.Error())
}

func TestDecodeWithListSeparator(t *testing.T) {
	patterns := []struct {
		name     string
		sep      string
		expected interface{}
	}{
		{"default", ":", []string{"a", "@b"}},
		{"multi", ":@", []string{"a", "b"}},
		{"none", "", "a:@b"},
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			d := ini.NewDecoder(ini.WithListSeparator(p.sep))
			require.NotNil(t, d)
			v := make(map[string]interface{})
			err := d.Decode(validConfig, &v)
			assert.Nil(t, err)
			val, ok := v["slice"]
			assert.True(t, ok)
			assert.Equal(t, p.expected, val)
		}
		t.Run(p.name, f)
	}
}

var validConfig = []byte(`
bool:true
int:42
float:3.1415
string = this is a string
slice = a:@b
intSlice = 1,2,3,4
stringSlice = one,two,three,four

[nested]
bool = false
int = 18
float = 3.141
string = this is also a string
intSlice = 1,2,3,4,5,6
stringSlice = one,two,three
`)

var malformedConfig = []byte(`
=malformed
bool: true
`)

var parsedConfig = map[string]interface{}{
	"bool":        "true",
	"int":         "42",
	"float":       "3.1415",
	"string":      "this is a string",
	"slice":       "a:@b",
	"intSlice":    []string{"1", "2", "3", "4"},
	"stringSlice": []string{"one", "two", "three", "four"},
	"nested": map[string]interface{}{
		"string":      "this is also a string",
		"intSlice":    []string{"1", "2", "3", "4", "5", "6"},
		"stringSlice": []string{"one", "two", "three"},
		"bool":        "false",
		"int":         "18",
		"float":       "3.141",
	},
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

func BenchmarkDecode(b *testing.B) {
	d := ini.NewDecoder()
	m := make(map[string]interface{})
	for n := 0; n < b.N; n++ {
		d.Decode(benchConfig, &m)
	}
}
