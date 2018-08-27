// Copyright © 2018 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package json_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/warthog618/config/blob/decoder/json"
)

func TestNewDecoder(t *testing.T) {
	d := json.NewDecoder()
	require.NotNil(t, d)
}

func TestDecode(t *testing.T) {
	d := json.NewDecoder()
	require.NotNil(t, d)
	m := make(map[string]interface{})
	err := d.Decode(malformedConfig, &m)
	assert.NotNil(t, err)
	assert.Equal(t, map[string]interface{}{}, m)
	err = d.Decode(validConfig, &m)
	assert.Nil(t, err)
	assert.Equal(t, parsedConfig, m)
}

var malformedConfig = []byte(`malformed{
	"bool": true,
	"int": 42,
	"float": 3.1415
  }`)

var validConfig = []byte(`{
	"bool": true,
	"int": 42,
	"float": 3.1415,
	"string": "this is a string",
	"intSlice": [1,2,3,4],
	"stringSlice": ["one","two","three","four"],
	"sliceslice": [[1,2,3,4],[5,6,7,8]],
	"nested": {
	  "bool": false,
	  "int": 18,
	  "float": 3.141,
	  "string": "this is also a string",
	  "intSlice": [1,2,3,4,5,6],
	  "stringSlice": ["one","two","three"]
	},
	"animals":[
	  {"Name": "Platypus", "Order": "Monotremata"},
	  {"Name": "Quoll",    "Order": "Dasyuromorphia"}
	]
  }`)

var parsedConfig = map[string]interface{}{
	"bool":        true,
	"int":         float64(42),
	"float":       float64(3.1415),
	"string":      "this is a string",
	"intSlice":    []interface{}{float64(1), float64(2), float64(3), float64(4)},
	"stringSlice": []interface{}{"one", "two", "three", "four"},
	"sliceslice": []interface{}{
		[]interface{}{float64(1), float64(2), float64(3), float64(4)},
		[]interface{}{float64(5), float64(6), float64(7), float64(8)}},
	"nested": map[string]interface{}{
		"string":      "this is also a string",
		"intSlice":    []interface{}{float64(1), float64(2), float64(3), float64(4), float64(5), float64(6)},
		"stringSlice": []interface{}{"one", "two", "three"},
		"bool":        false,
		"int":         float64(18),
		"float":       float64(3.141),
	},
	"animals": []interface{}{
		map[string]interface{}{"Name": "Platypus", "Order": "Monotremata"},
		map[string]interface{}{"Name": "Quoll", "Order": "Dasyuromorphia"},
	},
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

func BenchmarkDecode(b *testing.B) {
	d := json.NewDecoder()
	m := make(map[string]interface{})
	for n := 0; n < b.N; n++ {
		d.Decode(benchConfig, &m)
	}
}
