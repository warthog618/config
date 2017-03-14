// Copyright © 2017 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package json

import (
	"reflect"
	"testing"

	"github.com/warthog618/config"
	"github.com/warthog618/config/cfgconv"
)

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

var validKeys = []string{"bool", "int", "float", "string", "intSlice", "stringSlice",
	"nested", "nested.bool", "nested.int", "nested.float", "nested.string",
	"nested.intSlice", "nested.stringSlice"}

var bogusKeys = []string{"bogus", "nested.bogus"}

var intSlice = []interface{}{float64(1), float64(2), float64(3), float64(4)}
var nestedIntSlice = []interface{}{float64(1), float64(2), float64(3), float64(4), float64(5), float64(6)}
var stringSlice = []interface{}{"one", "two", "three", "four"}
var nestedStringSlice = []interface{}{"one", "two", "three"}

func testReaderContains(t *testing.T, reader *Reader) {
	for _, key := range validKeys {
		if ok := reader.Contains(key); !ok {
			t.Errorf("doesn't contain %s", key)
		}
	}
	for _, key := range bogusKeys {
		if reader.Contains(key) {
			t.Errorf("does contain %s", key)
		}
	}
}

// Test that config fields can be read and converted to required types using cfgconv.
func testReaderRead(t *testing.T, reader *Reader) {
	for _, key := range validKeys {
		if _, ok := reader.Read(key); !ok {
			t.Errorf("couldn't read %s", key)
		}
	}
	for _, key := range bogusKeys {
		if _, ok := reader.Read(key); ok {
			t.Errorf("could read %s", key)
		}
	}
	if val, ok := reader.Read("bool"); ok {
		if v, err := cfgconv.Bool(val); err != nil {
			t.Errorf("failed to convert bool")
		} else if v == false {
			t.Errorf("expected bool true, got false")
		}
	}
	if val, ok := reader.Read("int"); ok {
		if v, err := cfgconv.Int(val); err != nil {
			t.Errorf("failed to convert int")
		} else if v != 42 {
			t.Errorf("expected int 42, got %v", v)
		}
	}
	if val, ok := reader.Read("float"); ok {
		if v, err := cfgconv.Float(val); err != nil {
			t.Errorf("failed to convert float")
		} else if v != 3.1415 {
			t.Errorf("expected float 3.1415, got %v", v)
		}
	}
	if val, ok := reader.Read("string"); ok {
		if v, err := cfgconv.String(val); err != nil {
			t.Errorf("failed to convert string")
		} else if v != "this is a string" {
			t.Errorf("expected string 'this is a string', got %v", v)
		}
	}
	if val, ok := reader.Read("intSlice"); ok {
		if v, err := cfgconv.Slice(val); err != nil {
			t.Errorf("failed to convert slice")
		} else if !reflect.DeepEqual(v, intSlice) {
			t.Errorf("expected int slice %v, got %v", intSlice, v)
		}
	}
	if val, ok := reader.Read("stringSlice"); ok {
		if v, err := cfgconv.Slice(val); err != nil {
			t.Errorf("failed to convert slice")
		} else if !reflect.DeepEqual(v, stringSlice) {
			t.Errorf("expected string slice %v, got %v", stringSlice, v)
		}
	}
	if val, ok := reader.Read("nested.bool"); ok {
		if v, err := cfgconv.Bool(val); err != nil {
			t.Errorf("failed to convert bool")
		} else if v == true {
			t.Errorf("expected nested.bool false, got true")
		}
	}
	if val, ok := reader.Read("nested.int"); ok {
		if v, err := cfgconv.Int(val); err != nil {
			t.Errorf("failed to convert int")
		} else if v != 18 {
			t.Errorf("expected nested.int 18, got %v", v)
		}
	}
	if val, ok := reader.Read("nested.float"); ok {
		if v, err := cfgconv.Float(val); err != nil {
			t.Errorf("failed to convert float")
		} else if v != 3.141 {
			t.Errorf("expected nested.float 3.141, got %v", v)
		}
	}
	if val, ok := reader.Read("nested.string"); ok {
		if v, err := cfgconv.String(val); err != nil {
			t.Errorf("failed to convert string")
		} else if v != "this is also a string" {
			t.Errorf("expected nested.string 'this is also a string', got %v", v)
		}
	}
	if val, ok := reader.Read("nested.intSlice"); ok {
		if v, err := cfgconv.Slice(val); err != nil {
			t.Errorf("failed to convert slice")
		} else if !reflect.DeepEqual(v, nestedIntSlice) {
			t.Errorf("expected int slice %v, got %v", nestedIntSlice, v)
		}
	}
	if val, ok := reader.Read("nested.stringSlice"); ok {
		if v, err := cfgconv.Slice(val); err != nil {
			t.Errorf("failed to convert slice")
		} else if !reflect.DeepEqual(v, nestedStringSlice) {
			t.Errorf("expected string slice %v, got %v", nestedStringSlice, v)
		}
	}
}

func TestNewBytes(t *testing.T) {
	if _, err := NewBytes(malformedConfig); err == nil {
		t.Errorf("parsed malformed config")
	}
	if b, err := NewBytes(validConfig); err != nil {
		t.Errorf("failed to parse validConfig")
	} else {
		// test provides config.Reader interface.
		cfg := config.New()
		cfg.AppendReader(b)
	}
}

func TestNewFile(t *testing.T) {
	if _, err := NewFile("no_such.json"); err == nil {
		t.Errorf("parsed no such config")
	}
	if f, err := NewFile("config.json"); err != nil {
		t.Errorf("failed to parse config")
	} else {
		// test provides config.Reader interface.
		cfg := config.New()
		cfg.AppendReader(f)
	}
	if _, err := NewFile("malformed.json"); err == nil {
		t.Errorf("parsed malformed config")
	}
}

func TestBytesReaderContains(t *testing.T) {
	reader, err := NewBytes(validConfig)
	if err != nil {
		t.Fatalf("failed to parse config")
	}
	testReaderContains(t, reader)
}

func TestBytesReaderRead(t *testing.T) {
	reader, err := NewBytes(validConfig)
	if err != nil {
		t.Fatalf("failed to parse config")
	}
	testReaderRead(t, reader)
}

func TestFileReaderContains(t *testing.T) {
	reader, err := NewFile("config.json")
	if err != nil {
		t.Fatalf("failed to parse config")
	}
	testReaderContains(t, reader)
}

func TestFileReaderRead(t *testing.T) {
	reader, err := NewFile("config.json")
	if err != nil {
		t.Fatalf("failed to parse config")
	}
	testReaderRead(t, reader)
}
