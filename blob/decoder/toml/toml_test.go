// Copyright © 2018 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package toml_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/warthog618/config/blob/decoder/toml"
)

func TestNewDecoder(t *testing.T) {
	d := toml.NewDecoder()
	require.NotNil(t, d)
}

func TestDecode(t *testing.T) {
	d := toml.NewDecoder()
	require.NotNil(t, d)
	m := make(map[string]interface{})
	err := d.Decode(malformedConfig, &m)
	assert.NotNil(t, err)
	assert.Equal(t, map[string]interface{}{}, m)
	err = d.Decode(validConfig, &m)
	assert.Nil(t, err)
	assert.Equal(t, parsedConfig, m)
}

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
Name = "Platypus"
Order= "Monotremata"

[[animals]]
Name = "Quoll"
Order = "Dasyuromorphia"

`)

var malformedConfig = []byte(`
malformed
bool: true
`)

var parsedConfig = map[string]interface{}{
	"bool":        true,
	"int":         int64(42),
	"float":       float64(3.1415),
	"string":      "this is a string",
	"intSlice":    []interface{}{int64(1), int64(2), int64(3), int64(4)},
	"stringSlice": []interface{}{"one", "two", "three", "four"},
	"sliceslice": []interface{}{
		[]interface{}{int64(1), int64(2), int64(3), int64(4)},
		[]interface{}{int64(5), int64(6), int64(7), int64(8)}},
	"nested": map[string]interface{}{
		"string":      "this is also a string",
		"intSlice":    []interface{}{int64(1), int64(2), int64(3), int64(4), int64(5), int64(6)},
		"stringSlice": []interface{}{"one", "two", "three"},
		"bool":        false,
		"int":         int64(18),
		"float":       float64(3.141),
	},
	"animals": []map[string]interface{}{
		{"Name": "Platypus", "Order": "Monotremata"},
		{"Name": "Quoll", "Order": "Dasyuromorphia"},
	},
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

func BenchmarkDecode(b *testing.B) {
	d := toml.NewDecoder()
	m := make(map[string]interface{})
	for n := 0; n < b.N; n++ {
		d.Decode(benchConfig, &m)
	}
}
