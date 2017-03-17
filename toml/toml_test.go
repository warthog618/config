// Copyright © 2017 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package toml

import (
	"reflect"
	"testing"

	"github.com/warthog618/config"
	"github.com/warthog618/config/cfgconv"
)

var validConfig = []byte(`
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
string = "this is also a string"
intSlice = [1,2,3,4,5,6]
stringSlice = ["one","two","three"]
`)

var malformedConfig = []byte(`
malformed
bool: true
`)

var validKeys = []string{"bool", "int", "float", "string", "intSlice", "stringSlice",
	"nested.bool", "nested.int", "nested.float", "nested.string",
	"nested.intSlice", "nested.stringSlice"}

var bogusKeys = []string{"bogus", "nested", "nested.bogus"}

var intSlice = []interface{}{int64(1), int64(2), int64(3), int64(4)}
var nestedIntSlice = []interface{}{int64(1), int64(2), int64(3), int64(4), int64(5), int64(6)}
var stringSlice = []interface{}{"one", "two", "three", "four"}
var nestedStringSlice = []interface{}{"one", "two", "three"}

// Test that config fields can be read and converted to required types using cfgconv.
func testReaderRead(t *testing.T, reader *Reader) {
	for _, key := range validKeys {
		if _, ok := reader.Read(key); !ok {
			t.Errorf("couldn't read %s", key)
		}
	}
	for _, key := range bogusKeys {
		if v, ok := reader.Read(key); ok {
			t.Errorf("could read %s", key)
		} else if v != nil {
			t.Errorf("returned non-nil on failed read for %s, got %v", key, v)
		}
	}
	if v, ok := reader.Read("bool"); ok {
		if cv, err := cfgconv.Bool(v); err != nil {
			t.Errorf("failed to convert bool")
		} else if cv == false {
			t.Errorf("expected bool true, got false")
		}
	}
	if v, ok := reader.Read("int"); ok {
		if cv, err := cfgconv.Int(v); err != nil {
			t.Errorf("failed to convert int")
		} else if cv != 42 {
			t.Errorf("expected int 42, got %v", cv)
		}
	}
	if v, ok := reader.Read("float"); ok {
		if cv, err := cfgconv.Float(v); err != nil {
			t.Errorf("failed to convert float")
		} else if cv != 3.1415 {
			t.Errorf("expected float 3.1415, got %v", cv)
		}
	}
	if v, ok := reader.Read("string"); ok {
		if cv, err := cfgconv.String(v); err != nil {
			t.Errorf("failed to convert string")
		} else if cv != "this is a string" {
			t.Errorf("expected string 'this is a string', got %v", cv)
		}
	}
	if v, ok := reader.Read("intSlice"); ok {
		if cv, err := cfgconv.Slice(v); err != nil {
			t.Errorf("failed to convert slice")
		} else if !reflect.DeepEqual(cv, intSlice) {
			t.Errorf("expected int slice %v, got %v", intSlice, cv)
		}
	}
	if v, ok := reader.Read("stringSlice"); ok {
		if cv, err := cfgconv.Slice(v); err != nil {
			t.Errorf("failed to convert slice")
		} else if !reflect.DeepEqual(cv, stringSlice) {
			t.Errorf("expected string slice %v, got %v", stringSlice, cv)
		}
	}
	if v, ok := reader.Read("nested.bool"); ok {
		if cv, err := cfgconv.Bool(v); err != nil {
			t.Errorf("failed to convert bool")
		} else if cv == true {
			t.Errorf("expected nested.bool false, got true")
		}
	}
	if v, ok := reader.Read("nested.int"); ok {
		if cv, err := cfgconv.Int(v); err != nil {
			t.Errorf("failed to convert int")
		} else if cv != 18 {
			t.Errorf("expected nested.int 18, got %v", cv)
		}
	}
	if v, ok := reader.Read("nested.float"); ok {
		if cv, err := cfgconv.Float(v); err != nil {
			t.Errorf("failed to convert float")
		} else if cv != 3.141 {
			t.Errorf("expected nested.float 3.141, got %v", cv)
		}
	}
	if v, ok := reader.Read("nested.string"); ok {
		if cv, err := cfgconv.String(v); err != nil {
			t.Errorf("failed to convert string")
		} else if cv != "this is also a string" {
			t.Errorf("expected nested.string 'this is also a string', got %v", cv)
		}
	}
	if v, ok := reader.Read("nested.intSlice"); ok {
		if cv, err := cfgconv.Slice(v); err != nil {
			t.Errorf("failed to convert slice")
		} else if !reflect.DeepEqual(cv, nestedIntSlice) {
			t.Errorf("expected int slice %v, got %v %T", nestedIntSlice, cv, cv)
		}
	}
	if v, ok := reader.Read("nested.stringSlice"); ok {
		if cv, err := cfgconv.Slice(v); err != nil {
			t.Errorf("failed to convert slice")
		} else if !reflect.DeepEqual(cv, nestedStringSlice) {
			t.Errorf("expected string slice %v, got %v", nestedStringSlice, cv)
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
	if _, err := NewFile("no_such.toml"); err == nil {
		t.Errorf("parsed no such config")
	}
	if f, err := NewFile("config.toml"); err != nil {
		t.Errorf("failed to parse config")
	} else {
		// test provides config.Reader interface.
		cfg := config.New()
		cfg.AppendReader(f)
	}
	if _, err := NewFile("malformed.toml"); err == nil {
		t.Errorf("parsed malformed config")
	}
}

func TestStringReaderRead(t *testing.T) {
	reader, err := NewBytes(validConfig)
	if err != nil {
		t.Fatalf("failed to parse config")
	}
	testReaderRead(t, reader)
}

func TestFileReaderRead(t *testing.T) {
	reader, err := NewFile("config.toml")
	if err != nil {
		t.Fatalf("failed to parse config")
	}
	testReaderRead(t, reader)
}
