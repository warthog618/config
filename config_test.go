// Copyright © 2017 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package config

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

type mapReader struct {
	// simple key value map.
	// Note keys must be added as lowercase for config.GetX to work.
	config map[string]interface{}
}

func (mr *mapReader) Contains(key string) bool {
	_, ok := mr.config[key]
	return ok
}

func (mr *mapReader) Read(key string) (interface{}, bool) {
	v, ok := mr.config[key]
	return v, ok
}

func TestNew(t *testing.T) {
	cfg := New()
	// just do something with it...
	if _, err := cfg.Get(""); err == nil {
		t.Errorf("Empty config contains something.")
	}
}

func TestAddAlias(t *testing.T) {
	cfg := New()
	mr := mapReader{map[string]interface{}{}}
	cfg.InsertReader(&mr)
	// alias maps newKey (requested) -> oldKey (in config)
	mr.config["oldthing"] = "an old config string"
	cfg.AddAlias("newthing", "oldthing")
	if v, err := cfg.Get("oldthing"); err != nil {
		t.Errorf("couldn't get oldthing - err '%v'", err)
	} else if oldthing, ok := v.(string); ok {
		if oldthing != mr.config["oldthing"] {
			t.Errorf("oldthing mismatch - expected '%v' but got '%v'", mr.config["oldthing"], oldthing)
		}
	} else {
		t.Errorf("oldthing is not a string")
	}
	if v, err := cfg.Get("newthing"); err != nil {
		t.Errorf("couldn't get newthing - err '%v'", err)
	} else if newthing, ok := v.(string); ok {
		if newthing != mr.config["oldthing"] {
			t.Errorf("newthing mismatch - expected '%v' but got '%v'", mr.config["oldthing"], newthing)
		}
	} else {
		t.Errorf("newthing is not a string")
	}
	// alias ignored if newKey exists
	mr.config["newthing"] = "a new config string"
	if v, err := cfg.Get("oldthing"); err != nil {
		t.Errorf("couldn't get oldthing - err '%v'", err)
	} else if oldthing, ok := v.(string); ok {
		if oldthing != mr.config["oldthing"] {
			t.Errorf("oldthing mismatch - expected '%v' but got '%v'", mr.config["oldthing"], oldthing)
		}
	} else {
		t.Errorf("oldthing is not a string")
	}
	if v, err := cfg.Get("newthing"); err != nil {
		t.Errorf("couldn't get newthing - err '%v'", err)
	} else if newthing, ok := v.(string); ok {
		if newthing != mr.config["newthing"] {
			t.Errorf("newthing mismatch - expected '%v' but got '%v'", mr.config["newthing"], newthing)
		}
	} else {
		t.Errorf("newthing is not a string")
	}
}

func TestAddAliasNested(t *testing.T) {
	cfg := New()
	mr := mapReader{map[string]interface{}{}}
	cfg.InsertReader(&mr)
	mr.config["a"] = "a"
	mr.config["foo.a"] = "foo.a"
	mr.config["foo.b"] = "foo.b"
	mr.config["bar.b"] = "bar.b"
	mr.config["bar.c"] = "bar.c"
	cfg.AddAlias("foo.a", "foo.a")
	assertGet(t, cfg, "foo.a", mr.config["foo.a"].(string), "nested leaf to self (ignored)")
	cfg.AddAlias("c", "foo.b")
	assertGet(t, cfg, "c", mr.config["foo.b"].(string), "root leaf to nested leaf")
	cfg.AddAlias("d", "c")
	refuteGet(t, cfg, "d", "alias to alias")
	cfg.AddAlias("baz.b", "a")
	assertGet(t, cfg, "baz.b", mr.config["a"].(string), "nested leaf to root leaf")
	cfg.AddAlias("node", "")
	assertGet(t, cfg, "node.a", mr.config["a"].(string), "nested node to root node")
	cfg.AddAlias("", "foo")
	assertGet(t, cfg, "b", mr.config["foo.b"].(string), "root node to nested node")
	cfg.AddAlias("baz", "bar")
	assertGet(t, cfg, "baz.c", mr.config["bar.c"].(string), "nested node to nested node")
	assertGet(t, cfg, "baz.b", mr.config["a"].(string), "leaf alias has priority over node alias")
	cfg.AddAlias("blob", "baz")
	refuteGet(t, cfg, "blob.b", "alias to node alias")
	// sub-tree
	if bazCfg, err := cfg.GetConfig("baz"); err == nil {
		assertGet(t, bazCfg, "b", mr.config["a"].(string), "sub-tree node leaf alias")
		assertGet(t, bazCfg, "c", mr.config["bar.c"].(string), "sub-tree node alias")
		bazCfg.AddAlias("d", "b")
		assertGet(t, bazCfg, "b", mr.config["a"].(string), "sub-tree local leaf alias")
	}
	refuteGet(t, cfg, "d", "sub-tree alias locality")
}

func TestAppendReader(t *testing.T) {
	cfg := New()
	mr1 := mapReader{map[string]interface{}{}}
	cfg.AppendReader(nil) // should be ignored
	cfg.InsertReader(&mr1)
	mr1.config["something"] = "a test string"
	if v, err := cfg.Get("something"); err != nil {
		t.Errorf("couldn't get something - err '%v'", err)
	} else if something, ok := v.(string); ok {
		if something != mr1.config["something"] {
			t.Errorf("something mismatch - expected '%v' but got '%v'", mr1.config["something"], something)
		}
	} else {
		t.Errorf("something is not a string")
	}
	// append a second reader
	mr2 := mapReader{map[string]interface{}{}}
	cfg.AppendReader(&mr2)
	mr2.config["something"] = "another test string"
	mr2.config["something else"] = "yet another test string"
	if v, err := cfg.Get("something"); err != nil {
		t.Errorf("couldn't get something - err '%v'", err)
	} else if something, ok := v.(string); ok {
		if something != mr1.config["something"] {
			t.Errorf("something mismatch - expected '%v' but got '%v'", mr1.config["something"], something)
		}
	} else {
		t.Errorf("something is not a string")
	}
	if v, err := cfg.Get("something else"); err != nil {
		t.Errorf("couldn't get something else - err '%v'", err)
	} else if something, ok := v.(string); ok {
		if something != mr2.config["something else"] {
			t.Errorf("something else mismatch - expected '%v' but got '%v'", mr2.config["something else"], something)
		}
	} else {
		t.Errorf("something else is not a string")
	}
}

func TestInsertReader(t *testing.T) {
	cfg := New()
	mr := mapReader{map[string]interface{}{}}
	cfg.InsertReader(nil) // should be ignored
	cfg.InsertReader(&mr)
	mr.config["something"] = "a test string"
	if v, err := cfg.Get("something"); err != nil {
		t.Errorf("couldn't get something - err '%v'", err)
	} else if something, ok := v.(string); ok {
		if something != mr.config["something"] {
			t.Errorf("something mismatch - expected '%v' but got '%v'", mr.config["something"], something)
		}
	} else {
		t.Errorf("something is not a string")
	}
	// overlay a second reader
	mr = mapReader{map[string]interface{}{}}
	cfg.InsertReader(&mr)
	mr.config["something"] = "another test string"
	if v, err := cfg.Get("something"); err != nil {
		t.Errorf("couldn't get something - err '%v'", err)
	} else if something, ok := v.(string); ok {
		if something != mr.config["something"] {
			t.Errorf("something mismatch - expected '%v' but got '%v'", mr.config["something"], something)
		}
	} else {
		t.Errorf("something is not a string")
	}
}

func TestSetSeparator(t *testing.T) {
	cfg := New()
	// separator is internal, so use nasty type assertion to check.
	if cfg.(*config).separator != "." {
		t.Errorf("default separator not set by New")
	}
	cfg.SetSeparator("_")
	if cfg.(*config).separator != "_" {
		t.Errorf("separator not set by SetSeparator")
	}
}

func assertGet(t *testing.T, cfg Config, key string, expected string, comment string) {
	if v, err := cfg.Get(key); err != nil {
		t.Errorf("%s - failed to get '%s'", comment, key)
	} else {
		if vstr, ok := v.(string); ok {
			if vstr != expected {
				t.Errorf("%s - didn't get '%s' - expected '%s', got '%v'", comment, key, expected, v)
			}
		} else {
			t.Errorf("%s - didn't get '%s' - expected '%s', got %v", comment, key, expected, v)
		}
	}
}

func refuteGet(t *testing.T, cfg Config, key string, comment string) {
	if v, err := cfg.Get(key); err == nil {
		t.Errorf("%s - succeeded to get '%s' - got '%v'", comment, key, v)
	} else {
		if nf, ok := err.(NotFoundError); ok {
			nfstr := nf.Error()
			if !strings.Contains(nfstr, key) {
				t.Errorf("not found error does not identify key %s - %v", key, nf)
			}
		} else {
			t.Errorf("get key (non existent) returned error other than NotFound:%v", err)
		}
	}
}

func TestGetOverlayed(t *testing.T) {
	cfg := New()
	// Single Reader
	mr1 := mapReader{map[string]interface{}{}}
	cfg.InsertReader(&mr1)
	mr1.config["a"] = "a - tier 1"
	mr1.config["b"] = "b - tier 1"
	mr1.config["c"] = "c - tier 1"
	assertGet(t, cfg, "a", mr1.config["a"].(string), "one reader get")
	assertGet(t, cfg, "b", mr1.config["b"].(string), "one reader get")
	assertGet(t, cfg, "c", mr1.config["c"].(string), "one reader get")
	refuteGet(t, cfg, "d", "one reader get")
	refuteGet(t, cfg, "e", "one reader get")

	// Two Readers
	mr2 := mapReader{map[string]interface{}{}}
	cfg.InsertReader(&mr2)
	mr2.config["b"] = "b - tier 2"
	mr2.config["d"] = "d - tier 2"
	assertGet(t, cfg, "a", mr1.config["a"].(string), "two reader get")
	assertGet(t, cfg, "b", mr2.config["b"].(string), "two reader get")
	assertGet(t, cfg, "c", mr1.config["c"].(string), "two reader get")
	assertGet(t, cfg, "d", mr2.config["d"].(string), "two reader get")
	refuteGet(t, cfg, "e", "two reader get")

	// Three Readers
	mr3 := mapReader{map[string]interface{}{}}
	cfg.InsertReader(&mr3)
	mr3.config["c"] = "c - tier 3"
	mr3.config["d"] = "d - tier 3"
	assertGet(t, cfg, "a", mr1.config["a"].(string), "three reader get")
	assertGet(t, cfg, "b", mr2.config["b"].(string), "three reader get")
	assertGet(t, cfg, "c", mr3.config["c"].(string), "three reader get")
	assertGet(t, cfg, "d", mr3.config["d"].(string), "three reader get")
	refuteGet(t, cfg, "e", "three reader get")
}

func TestGetBool(t *testing.T) {
	cfg := New()
	mr := mapReader{map[string]interface{}{}}
	cfg.InsertReader(&mr)
	mr.config["bool"] = true
	mr.config["boolString"] = "true"
	mr.config["boolInt"] = 1
	mr.config["notabool"] = "bogus"
	if v, err := cfg.GetBool("bool"); err != nil {
		t.Errorf("couldn't read bool - %v", err)
	} else if v != true {
		t.Errorf("read bool %v, expected true", v)
	}
	if v, err := cfg.GetBool("boolString"); err != nil {
		t.Errorf("couldn't read boolString - %v", err)
	} else if v != true {
		t.Errorf("read boolString %v, expected true", v)
	}
	if v, err := cfg.GetBool("boolInt"); err != nil {
		t.Errorf("couldn't read boolInt - %v", err)
	} else if v != true {
		t.Errorf("read boolInt %v, expected true", v)
	}
	if v, err := cfg.GetBool("notabool"); err == nil {
		t.Errorf("could read notabool -%v", v)
	} else {
		if v != false {
			t.Errorf("didn't return false -%v", v)
		}
	}
	if v, err := cfg.GetBool("nosuchbool"); err == nil {
		t.Errorf("could read nosuchbool -%v", v)
	} else {
		if v != false {
			t.Errorf("didn't return false -%v", v)
		}
	}
}

func TestGetDuration(t *testing.T) {
	cfg := New()
	mr := mapReader{map[string]interface{}{}}
	cfg.InsertReader(&mr)
	mr.config["duration"] = "123ms"
	mr.config["notaduration"] = "bogus"
	if v, err := cfg.GetDuration("duration"); err != nil {
		t.Errorf("couldn't read duration - %v", err)
	} else if v != time.Duration(123000000) {
		t.Errorf("read duration %v, expected 123ms", v)
	}
	if v, err := cfg.GetDuration("notaduration"); err == nil {
		t.Errorf("could read duration - notaduration -%v", v)
	} else {
		if v != 0 {
			t.Errorf("didn't return 0 -%v", v)
		}
	}
	if v, err := cfg.GetDuration("nosuchduration"); err == nil {
		t.Errorf("could read duration - nosuchduration -%v", v)
	} else {
		if v != 0 {
			t.Errorf("didn't return 0 -%v", v)
		}
	}
}

func TestGetFloat(t *testing.T) {
	cfg := New()
	mr := mapReader{map[string]interface{}{}}
	cfg.InsertReader(&mr)
	mr.config["float"] = 3.1415
	mr.config["floatString"] = "3.1415"
	mr.config["floatInt"] = 1
	mr.config["notafloat"] = "bogus"
	if v, err := cfg.GetFloat("float"); err != nil {
		t.Errorf("couldn't read float - %v", err)
	} else if v != 3.1415 {
		t.Errorf("read float %v, expected 3.1415", v)
	}
	if v, err := cfg.GetFloat("floatString"); err != nil {
		t.Errorf("couldn't read floatString - %v", err)
	} else if v != 3.1415 {
		t.Errorf("read floatString %v, expected 3.1415", v)
	}
	if v, err := cfg.GetFloat("floatInt"); err != nil {
		t.Errorf("couldn't read floatInt - %v", err)
	} else if v != 1 {
		t.Errorf("read floatInt %v, expected 1", v)
	}
	if v, err := cfg.GetFloat("notafloat"); err == nil {
		t.Errorf("could read float - notafloat -%v", v)
	} else {
		if v != 0 {
			t.Errorf("didn't return 0 -%v", v)
		}
	}
	if v, err := cfg.GetFloat("nosuchfloat"); err == nil {
		t.Errorf("could read float - nosuchfloat -%v", v)
	} else {
		if v != 0 {
			t.Errorf("didn't return 0 -%v", v)
		}
	}
}

func TestGetInt(t *testing.T) {
	cfg := New()
	mr := mapReader{map[string]interface{}{}}
	cfg.InsertReader(&mr)
	mr.config["int"] = 42
	mr.config["intString"] = "42"
	mr.config["notaint"] = "bogus"
	if v, err := cfg.GetInt("int"); err != nil {
		t.Errorf("couldn't read int - %v", err)
	} else if v != 42 {
		t.Errorf("read int %v, expected 3.1415", v)
	}
	if v, err := cfg.GetInt("intString"); err != nil {
		t.Errorf("couldn't read intString - %v", err)
	} else if v != 42 {
		t.Errorf("read intString %v, expected 3.1415", v)
	}
	if v, err := cfg.GetInt("notaint"); err == nil {
		t.Errorf("could read int - notaint -%v", v)
	} else {
		if v != 0 {
			t.Errorf("didn't return 0 -%v", v)
		}
	}
	if v, err := cfg.GetInt("nosuchint"); err == nil {
		t.Errorf("could read int - nosuchint -%v", v)
	} else {
		if v != 0 {
			t.Errorf("didn't return 0 -%v", v)
		}
	}
}

func TestGetString(t *testing.T) {
	cfg := New()
	mr := mapReader{map[string]interface{}{}}
	cfg.InsertReader(&mr)
	mr.config["string"] = "a string"
	mr.config["stringInt"] = 42
	mr.config["notastring"] = struct{}{}
	if v, err := cfg.GetString("string"); err != nil {
		t.Errorf("couldn't read string - %v", err)
	} else if v != "a string" {
		t.Errorf("read string %v, expected 3.1415", v)
	}
	if v, err := cfg.GetString("stringInt"); err != nil {
		t.Errorf("couldn't read stringInt - %v", err)
	} else if v != "42" {
		t.Errorf("read stringInt %v, expected 3.1415", v)
	}
	if v, err := cfg.GetString("notastring"); err == nil {
		t.Errorf("could read string - notastring -%v", v)
	} else {
		if v != "" {
			t.Errorf("didn't return empty -%v", v)
		}
	}
	if v, err := cfg.GetString("nosuchstring"); err == nil {
		t.Errorf("could read string - nosuchstring -%v", v)
	} else {
		if v != "" {
			t.Errorf("didn't return empty -%v", v)
		}
	}
}

func TestGetTime(t *testing.T) {
	cfg := New()
	mr := mapReader{map[string]interface{}{}}
	cfg.InsertReader(&mr)
	mr.config["time"] = "2017-03-01T01:02:03Z"
	mr.config["notatime"] = "bogus"
	if v, err := cfg.GetTime("time"); err != nil {
		t.Errorf("couldn't read time - %v", err)
	} else if v != time.Date(2017, 3, 1, 1, 2, 3, 0, time.UTC) {
		t.Errorf("read time %v, expected 123ms", v)
	}
	if v, err := cfg.GetTime("notatime"); err == nil {
		t.Errorf("could read time - notatime -%v", v)
	} else {
		if !reflect.DeepEqual(v, time.Time{}) {
			t.Errorf("didn't return 0 -%v", v)
		}
	}
	if v, err := cfg.GetTime("nosuchtime"); err == nil {
		t.Errorf("could read time - nosuchtime -%v", v)
	} else {
		if !reflect.DeepEqual(v, time.Time{}) {
			t.Errorf("didn't return 0 -%v", v)
		}
	}
}

func TestGetUint(t *testing.T) {
	cfg := New()
	mr := mapReader{map[string]interface{}{}}
	cfg.InsertReader(&mr)
	mr.config["uint"] = 42
	mr.config["uintString"] = "42"
	mr.config["notaUint"] = "bogus"
	if v, err := cfg.GetUint("uint"); err != nil {
		t.Errorf("couldn't read uint - %v", err)
	} else if v != 42 {
		t.Errorf("read uint %v, expected 3.1415", v)
	}
	if v, err := cfg.GetUint("uintString"); err != nil {
		t.Errorf("couldn't read uint - %v", err)
	} else if v != 42 {
		t.Errorf("read uint %v, expected 3.1415", v)
	}
	if v, err := cfg.GetUint("notaUint"); err == nil {
		t.Errorf("could read notaUint -%v", v)
	} else {
		if v != 0 {
			t.Errorf("didn't return 0 -%v", v)
		}
	}
	if v, err := cfg.GetUint("nosuchUint"); err == nil {
		t.Errorf("could read nosuchUint -%v", v)
	} else {
		if v != 0 {
			t.Errorf("didn't return 0 -%v", v)
		}
	}
}

func TestGetSlice(t *testing.T) {
	cfg := New()
	mr := mapReader{map[string]interface{}{}}
	cfg.InsertReader(&mr)
	mr.config["slice"] = []interface{}{1, 2, 3, 4}
	mr.config["casttoslice"] = "bogus"
	mr.config["notaslice"] = struct{}{}
	if v, err := cfg.GetSlice("slice"); err != nil {
		t.Errorf("couldn't read slice - %v", err)
	} else if !reflect.DeepEqual(v, []interface{}{1, 2, 3, 4}) {
		t.Errorf("read slice %v, expected %v", v, mr.config["slice"])
	}
	if v, err := cfg.GetSlice("casttoslice"); err != nil {
		t.Errorf("couldn't read slice - %v", err)
	} else if !reflect.DeepEqual(v, []interface{}{"bogus"}) {
		t.Errorf("read slice %v, expected %v", v, mr.config["casttoslice"])
	}
	if v, err := cfg.GetSlice("notaslice"); err == nil {
		t.Errorf("could read slice - notaslice -%v", v)
	} else {
		if len(v) != 0 {
			t.Errorf("didn't return empty -%v", v)
		}
	}
	if v, err := cfg.GetSlice("nosuchslice"); err == nil {
		t.Errorf("could read slice - nosuchslice -%v", v)
	} else {
		if len(v) != 0 {
			t.Errorf("didn't return empty -%v", v)
		}
	}
}

func TestGetIntSlice(t *testing.T) {
	cfg := New()
	mr := mapReader{map[string]interface{}{}}
	cfg.InsertReader(&mr)
	mr.config["slice"] = []int64{1, 2, -3, 4}
	mr.config["casttoslice"] = "42"
	mr.config["stringslice"] = []string{"one", "two", "three"}
	mr.config["notaslice"] = "bogus"
	if v, err := cfg.GetIntSlice("slice"); err != nil {
		t.Errorf("couldn't read slice - %v", err)
	} else if !reflect.DeepEqual(v, mr.config["slice"]) {
		t.Errorf("read slice %v, expected %v", v, mr.config["slice"])
	}
	if v, err := cfg.GetIntSlice("casttoslice"); err != nil {
		t.Errorf("couldn't read slice - %v", err)
	} else if !reflect.DeepEqual(v, []int64{42}) {
		t.Errorf("read slice %v, expected %v", v, []int64{42})
	}
	if v, err := cfg.GetIntSlice("stringslice"); err == nil {
		t.Errorf("could read slice - stringslice -%v", v)
	} else {
		if len(v) != 0 {
			t.Errorf("didn't return empty -%v", v)
		}
	}
	if v, err := cfg.GetIntSlice("notaslice"); err == nil {
		t.Errorf("could read slice - notaslice -%v", v)
	} else {
		if len(v) != 0 {
			t.Errorf("didn't return empty -%v", v)
		}
	}
	if v, err := cfg.GetIntSlice("nosuchslice"); err == nil {
		t.Errorf("could read slice - nosuchslice -%v", v)
	} else {
		if len(v) != 0 {
			t.Errorf("didn't return empty -%v", v)
		}
	}
}

func TestGetStringSlice(t *testing.T) {
	cfg := New()
	mr := mapReader{map[string]interface{}{}}
	cfg.InsertReader(&mr)
	mr.config["intslice"] = []int64{1, 2, -3, 4}
	mr.config["stringslice"] = []string{"one", "two", "three"}
	mr.config["uintslice"] = []uint64{1, 2, 3, 4}
	mr.config["notastringslice"] = []interface{}{1, 2, struct{}{}}
	mr.config["casttoslice"] = "bogus"
	mr.config["notaslice"] = struct{}{}
	expectedIntSlice := []string{"1", "2", "-3", "4"}
	expectedUintSlice := []string{"1", "2", "3", "4"}
	expectedCastToSlice := []string{"bogus"}
	if v, err := cfg.GetStringSlice("stringslice"); err != nil {
		t.Errorf("couldn't read slice - %v", err)
	} else if !reflect.DeepEqual(v, mr.config["stringslice"]) {
		t.Errorf("read slice %v, expected %v", v, mr.config["stringslice"])
	}
	if v, err := cfg.GetStringSlice("casttoslice"); err != nil {
		t.Errorf("couldn't read slice - %v", err)
	} else if !reflect.DeepEqual(v, expectedCastToSlice) {
		t.Errorf("read slice %v, expected %v", v, expectedCastToSlice)
	}
	if v, err := cfg.GetStringSlice("intslice"); err != nil {
		t.Errorf("couldn't read slice - intslice -%v", v)
	} else if !reflect.DeepEqual(v, expectedIntSlice) {
		t.Errorf("read slice %v, expected %v", v, expectedIntSlice)
	}
	if v, err := cfg.GetStringSlice("uintslice"); err != nil {
		t.Errorf("couldn't read slice - uintslice -%v", v)
	} else if !reflect.DeepEqual(v, expectedUintSlice) {
		t.Errorf("read slice %v, expected %v", v, expectedUintSlice)
	}
	if v, err := cfg.GetStringSlice("notastringslice"); err == nil {
		t.Errorf("could read slice - notastringslice -%v", v)
	} else {
		if len(v) != 0 {
			t.Errorf("didn't return empty -%v", v)
		}
	}
	if v, err := cfg.GetStringSlice("notaslice"); err == nil {
		t.Errorf("could read slice - notaslice -%v", v)
	} else {
		if len(v) != 0 {
			t.Errorf("didn't return empty -%v", v)
		}
	}
	if v, err := cfg.GetStringSlice("nosuchslice"); err == nil {
		t.Errorf("could read slice - nosuchslice -%v", v)
	} else {
		if len(v) != 0 {
			t.Errorf("didn't return empty -%v", v)
		}
	}
}

func TestGetUintSlice(t *testing.T) {
	cfg := New()
	mr := mapReader{map[string]interface{}{}}
	cfg.InsertReader(&mr)
	mr.config["slice"] = []uint64{1, 2, 3, 4}
	mr.config["casttoslice"] = "42"
	mr.config["intslice"] = []int64{1, 2, -3, 4}
	mr.config["stringslice"] = []string{"one", "two", "three"}
	mr.config["notaslice"] = "bogus"
	if v, err := cfg.GetUintSlice("slice"); err != nil {
		t.Errorf("couldn't read slice - %v", err)
	} else if !reflect.DeepEqual(v, mr.config["slice"]) {
		t.Errorf("read slice %v, expected %v", v, mr.config["slice"])
	}
	if v, err := cfg.GetUintSlice("casttoslice"); err != nil {
		t.Errorf("couldn't read slice - %v", err)
	} else if !reflect.DeepEqual(v, []uint64{42}) {
		t.Errorf("read slice %v, expected %v", v, []uint64{42})
	}
	if v, err := cfg.GetUintSlice("intslice"); err == nil {
		t.Errorf("could read slice - intslice -%v", v)
	} else {
		if len(v) != 0 {
			t.Errorf("didn't return empty -%v", v)
		}
	}
	if v, err := cfg.GetUintSlice("stringslice"); err == nil {
		t.Errorf("could read slice - stringslice -%v", v)
	} else {
		if len(v) != 0 {
			t.Errorf("didn't return empty -%v", v)
		}
	}
	if v, err := cfg.GetUintSlice("notaslice"); err == nil {
		t.Errorf("could read slice - notaslice -%v", v)
	} else {
		if len(v) != 0 {
			t.Errorf("didn't return empty -%v", v)
		}
	}
	if v, err := cfg.GetUintSlice("nosuchslice"); err == nil {
		t.Errorf("could read slice - nosuchslice -%v", v)
	} else {
		if len(v) != 0 {
			t.Errorf("didn't return empty -%v", v)
		}
	}
}

func TestGetConfig(t *testing.T) {
	cfg := New()
	// Single Reader
	mr := mapReader{map[string]interface{}{}}
	cfg.InsertReader(&mr)
	mr.config["foo.a"] = "foo.a"
	mr.config["foo.b"] = "foo.b"
	mr.config["bar.b"] = "bar.b"
	mr.config["bar.c"] = "bar.c"
	mr.config["baz.a"] = "baz.a"
	mr.config["baz.c"] = "baz.c"
	cfg.AddAlias("foo.d", "bar.c")
	if rootCfg, err := cfg.GetConfig(""); err == nil {
		assertGet(t, rootCfg, "foo.a", mr.config["foo.a"].(string), "root config")
		assertGet(t, rootCfg, "bar.b", mr.config["bar.b"].(string), "root config")
		assertGet(t, rootCfg, "baz.c", mr.config["baz.c"].(string), "root config")
	} else {
		t.Errorf("failed to get root config")
	}
	if fooCfg, err := cfg.GetConfig("foo"); err == nil {
		assertGet(t, fooCfg, "a", mr.config["foo.a"].(string), "foo config")
		assertGet(t, fooCfg, "b", mr.config["foo.b"].(string), "foo config")
		refuteGet(t, fooCfg, "c", "foo config")
		// alias in cfg
		assertGet(t, fooCfg, "d", mr.config["bar.c"].(string), "foo config")
		// alias in fooCfg
		fooCfg.AddAlias("e", "b")
		assertGet(t, fooCfg, "e", mr.config["foo.b"].(string), "foo config")
		refuteGet(t, cfg, "foo.e", "foo config")
	} else {
		t.Errorf("failed to get foo config")
	}
	if barCfg, err := cfg.GetConfig("bar"); err == nil {
		refuteGet(t, barCfg, "a", "bar config")
		assertGet(t, barCfg, "b", mr.config["bar.b"].(string), "bar config")
		assertGet(t, barCfg, "c", mr.config["bar.c"].(string), "bar config")
		refuteGet(t, barCfg, "e", "bar config")
	} else {
		t.Errorf("failed to get bar config")
	}
	if bazCfg, err := cfg.GetConfig("baz"); err == nil {
		assertGet(t, bazCfg, "a", mr.config["baz.a"].(string), "baz config")
		refuteGet(t, bazCfg, "b", "baz config")
		assertGet(t, bazCfg, "c", mr.config["baz.c"].(string), "baz config")
	} else {
		t.Errorf("failed to get bar config")
	}
	if blahCfg, err := cfg.GetConfig("blah"); err == nil {
		refuteGet(t, blahCfg, "a", "blah config")
		refuteGet(t, blahCfg, "b", "blah config")
		refuteGet(t, blahCfg, "c", "blah config")
	} else {
		t.Errorf("failed to get blah config")
	}
}

type fooConfig struct {
	Atagged int `config:"a"`
	B       string
	C       []int
	E       string
}

type innerConfig struct {
	A       int
	Btagged string `config:"b"`
	C       []int
	E       string
}

type nestedConfig struct {
	Atagged int `config:"a"`
	B       string
	C       []int
	Nested  innerConfig `config:"nested"`
}

func TestUnmarshal(t *testing.T) {
	cfg := New()
	// Root Reader
	mr := mapReader{map[string]interface{}{}}
	cfg.InsertReader(&mr)
	mr.config["foo.a"] = 42
	mr.config["foo.b"] = "foo.b"
	mr.config["foo.c"] = []int{1, 2, 3, 4}
	mr.config["foo.d"] = "ignored"
	if err := cfg.Unmarshal("foo", 0); err == nil {
		t.Errorf("failed to reject unmarshal into non-struct")
	}
	foo := fooConfig{}
	foo.E = "some useful default"
	// correctly typed
	if err := cfg.Unmarshal("foo", &foo); err == nil {
		if foo.Atagged != mr.config["foo.a"] {
			t.Errorf("failed to unmarshal 'foo.a', expected %v, got %v", mr.config["foo.a"], foo.Atagged)
		}
		if foo.B != mr.config["foo.b"] {
			t.Errorf("failed to unmarshal 'foo.b', expected %v, got %v", mr.config["foo.b"], foo.B)
		}
		if !reflect.DeepEqual(foo.C, mr.config["foo.c"]) {
			t.Errorf("failed to unmarshal 'foo.c', expected %v, got %v", mr.config["foo.c"], foo.C)
		}
		if foo.E != "some useful default" {
			t.Errorf("unmarshalled 'foo.e', got %v", foo.B)
		}
	} else {
		t.Errorf("failed to unmarshal foo")
	}
	// Mistyped
	mr.config["foo.a"] = []int{1, 2}
	foo = fooConfig{}
	if err := cfg.Unmarshal("foo", &foo); err != nil {
		if !strings.Contains(err.Error(), "foo.a") {
			t.Errorf("unmarshal error doesn't identify key 'foo.a' - %v", err)
		}
		if foo.Atagged != 0 {
			t.Errorf("set mistyped 'foo.a', expected 0, got %v", foo.Atagged)
		}
		if foo.B != mr.config["foo.b"] {
			t.Errorf("failed to unmarshal 'foo.b', expected %v, got %v", mr.config["foo.b"], foo.B)
		}
		if !reflect.DeepEqual(foo.C, mr.config["foo.c"]) {
			t.Errorf("failed to unmarshal 'foo.c', expected %v, got %v", mr.config["foo.c"], foo.C)
		}
	} else {
		t.Errorf("successfully unmarshalled mistyped foo")
	}
	mr.config["foo.a"] = 42
	// Nested
	mr.config["foo.nested.a"] = 43
	mr.config["foo.nested.b"] = "foo.nested.b"
	mr.config["foo.nested.c"] = []int{1, 2, -3, 4}
	nc := nestedConfig{}
	if err := cfg.Unmarshal("foo", &nc); err == nil {
		if nc.Atagged != mr.config["foo.a"] {
			t.Errorf("failed to unmarshal 'foo.a', expected %v, got %v", mr.config["foo.a"], nc.Atagged)
		}
		if nc.B != mr.config["foo.b"] {
			t.Errorf("failed to unmarshal 'foo.b', expected %v, got %v", mr.config["foo.b"], nc.B)
		}
		if !reflect.DeepEqual(nc.C, mr.config["foo.c"]) {
			t.Errorf("failed to unmarshal 'foo.c', expected %v, got %v", mr.config["foo.c"], nc.C)
		}
		if nc.Nested.A != mr.config["foo.nested.a"] {
			t.Errorf("failed to unmarshal 'foo.nested.a', expected %v, got %v", mr.config["foo.nested.a"], nc.Nested.A)
		}
		if nc.Nested.Btagged != mr.config["foo.nested.b"] {
			t.Errorf("failed to unmarshal 'foo.nested.b', expected %v, got %v", mr.config["foo.nested.b"], nc.Nested.Btagged)
		}
		if !reflect.DeepEqual(nc.Nested.C, mr.config["foo.nested.c"]) {
			t.Errorf("failed to unmarshal 'foo.nested.c', expected %v, got %v", mr.config["foo.nested.c"], nc.Nested.C)
		}
	} else {
		t.Errorf("failed to unmarshal foo")
	}
	// Nested mistyped
	mr.config["foo.nested.a"] = []int{}
	nc = nestedConfig{}
	if err := cfg.Unmarshal("foo", &nc); err != nil {
		if !strings.Contains(err.Error(), "foo.nested.a") {
			t.Errorf("unmarshal error doesn't identify key 'foo.nested.a' - %v", err)
		}
		if nc.Atagged != mr.config["foo.a"] {
			t.Errorf("failed to unmarshal 'foo.a', expected %v, got %v", mr.config["foo.a"], nc.Atagged)
		}
		if nc.B != mr.config["foo.b"] {
			t.Errorf("failed to unmarshal 'foo.b', expected %v, got %v", mr.config["foo.b"], nc.B)
		}
		if !reflect.DeepEqual(nc.C, mr.config["foo.c"]) {
			t.Errorf("failed to unmarshal 'foo.c', expected %v, got %v", mr.config["foo.c"], nc.C)
		}
		if nc.Nested.A != 0 {
			t.Errorf("set mistyped 'foo.nested.a', expected 0, got %v", nc.Nested.A)
		}
		if nc.Nested.Btagged != mr.config["foo.nested.b"] {
			t.Errorf("failed to unmarshal 'foo.nested.b', expected %v, got %v", mr.config["foo.nested.b"], nc.Nested.Btagged)
		}
		if !reflect.DeepEqual(nc.Nested.C, mr.config["foo.nested.c"]) {
			t.Errorf("failed to unmarshal 'foo.nested.c', expected %v, got %v", mr.config["foo.nested.c"], nc.Nested.C)
		}
	} else {
		t.Errorf("successfully unmarshalled mistyped foo.nested")
	}
	mr.config["foo.nested.a"] = 43

	// Aliased
	mr.config["foo.b"] = "foo.b"
	cfg.AddAlias("foo.nested.e", "foo.b")
	nc = nestedConfig{}
	if err := cfg.Unmarshal("foo", &nc); err == nil {
		if !reflect.DeepEqual(nc.Nested.E, mr.config["foo.b"]) {
			t.Errorf("failed to unmarshal 'foo.nested.e', expected %v, got %v", mr.config["foo.b"], nc.Nested.E)
		}
	} else {
		t.Errorf("failed to unmarshal foo")
	}
}

func TestUnmarshalToMap(t *testing.T) {
	cfg := New()
	// Root Reader
	mr := mapReader{map[string]interface{}{}}
	cfg.InsertReader(&mr)
	mr.config["foo.a"] = 42
	mr.config["foo.b"] = "foo.b"
	mr.config["foo.c"] = []int{1, 2, 3, 4}
	mr.config["foo.d"] = "ignored"
	// Nil - raw
	obj := map[string]interface{}{"a": nil, "b": nil, "c": nil, "e": nil}
	if err := cfg.UnmarshalToMap("foo", obj); err == nil {
		if obj["a"] != mr.config["foo.a"] {
			t.Errorf("failed to unmarshal 'foo.a', expected %v, got %v", mr.config["foo.a"], obj["a"])
		}
		if obj["b"] != mr.config["foo.b"] {
			t.Errorf("failed to unmarshal 'foo.b', expected %v, got %v", mr.config["foo.b"], obj["b"])
		}
		if !reflect.DeepEqual(obj["c"], mr.config["foo.c"]) {
			t.Errorf("failed to unmarshal 'foo.c', expected %v, got %v", mr.config["foo.c"], obj["c"])
		}
		if v, ok := obj["d"]; ok {
			t.Errorf("unmarshalled unrequested 'd', got %v", v)
		}
		if obj["e"] != nil {
			t.Errorf("unmarshalled unconfigured 'e', expected %v, got %v", nil, obj["e"])
		}
	} else {
		t.Errorf("failed to unmarshal foo")
	}
	// Typed
	obj = map[string]interface{}{"a": int(0), "b": "", "c": []int{}, "e": "some useful default"}
	if err := cfg.UnmarshalToMap("foo", obj); err == nil {
		if obj["a"] != mr.config["foo.a"] {
			t.Errorf("failed to unmarshal 'foo.a', expected %#v, got %#v", mr.config["foo.a"], obj["a"])
		}
		if obj["b"] != mr.config["foo.b"] {
			t.Errorf("failed to unmarshal 'foo.b', expected %v, got %v", mr.config["foo.b"], obj["b"])
		}
		if !reflect.DeepEqual(obj["c"], mr.config["foo.c"]) {
			t.Errorf("failed to unmarshal 'foo.c', expected %v, got %v", mr.config["foo.c"], obj["c"])
		}
		if v, ok := obj["d"]; ok {
			t.Errorf("unmarshalled unrequested 'd', got %v", v)
		}
		if obj["e"] != "some useful default" {
			t.Errorf("unmarshalled unconfigured 'e', expected %v, got %v", "some useful default", obj["e"])
		}
	} else {
		t.Errorf("failed to unmarshal foo")
	}
	// Mistyped
	obj = map[string]interface{}{"a": []int{}}
	if err := cfg.UnmarshalToMap("foo", obj); err == nil {
		t.Errorf("successfully unmarshalled foo.a - %v", obj["a"])
	} else if !strings.Contains(err.Error(), "foo.a") {
		t.Errorf("unmarshal error doesn't identify key 'foo.a' - %v", err)
	}
	obj = map[string]interface{}{"b": 44}
	if err := cfg.UnmarshalToMap("foo", obj); err == nil {
		t.Errorf("successfully unmarshalled foo.b - %v", obj["b"])
	}
	obj = map[string]interface{}{"c": ""}
	if err := cfg.UnmarshalToMap("foo", obj); err == nil {
		t.Errorf("successfully unmarshalled foo.c - %v", obj["c"])
	}
	// Nested
	mr.config["foo.nested.a"] = 43
	mr.config["foo.nested.b"] = "foo.nested.b"
	mr.config["foo.nested.c"] = []int{1, 2, -3, 4}
	obj = map[string]interface{}{"a": nil,
		"nested": map[string]interface{}{"a": int(0), "b": "", "c": []int{}}}
	n1 := obj["nested"].(map[string]interface{})
	if err := cfg.UnmarshalToMap("foo", obj); err == nil {
		if obj["a"] != mr.config["foo.a"] {
			t.Errorf("failed to unmarshal 'foo.a', expected %v, got %v", mr.config["foo.a"], obj["a"])
		}
		if n1["a"] != mr.config["foo.nested.a"] {
			t.Errorf("failed to unmarshal 'foo.nested.a', expected %#v, got %#v", mr.config["foo.a"], n1["a"])
		}
		if n1["b"] != mr.config["foo.nested.b"] {
			t.Errorf("failed to unmarshal 'foo.nested.b', expected %v, got %v", mr.config["foo.b"], n1["b"])
		}
		if !reflect.DeepEqual(n1["c"], mr.config["foo.nested.c"]) {
			t.Errorf("failed to unmarshal 'foo.nested.c', expected %v, got %v", mr.config["foo.c"], n1["c"])
		}
		if v, ok := obj["d"]; ok {
			t.Errorf("unmarshalled unrequested 'd', got %v", v)
		}
	} else {
		t.Errorf("failed to unmarshal foo")
	}
	// nested - mistyped
	mr.config["foo.nested.a"] = []int{}
	mr.config["foo.nested.b"] = "foo.nested.b"
	mr.config["foo.nested.c"] = []int{1, 2, -3, 4}
	obj = map[string]interface{}{"a": nil,
		"nested": map[string]interface{}{"a": int(0), "b": "", "c": []int{}}}
	n1 = obj["nested"].(map[string]interface{})
	if err := cfg.UnmarshalToMap("foo", obj); err != nil {
		if !strings.Contains(err.Error(), "foo.nested.a") {
			t.Errorf("unmarshal error doesn't identify key 'foo.nested.a' - %v", err)
		}
		if obj["a"] != mr.config["foo.a"] {
			t.Errorf("failed to unmarshal 'foo.a', expected %v, got %v", mr.config["foo.a"], obj["a"])
		}
		if n1["a"] != 0 {
			t.Errorf("set mistyped 'foo.nested.a', expected 0, got %v", n1["a"])
		}
		if n1["b"] != mr.config["foo.nested.b"] {
			t.Errorf("failed to unmarshal 'foo.nested.b', expected %v, got %v", mr.config["foo.b"], n1["b"])
		}
		if !reflect.DeepEqual(n1["c"], mr.config["foo.nested.c"]) {
			t.Errorf("failed to unmarshal 'foo.nested.c', expected %v, got %v", mr.config["foo.c"], n1["c"])
		}
		if v, ok := obj["d"]; ok {
			t.Errorf("unmarshalled unrequested 'd', got %v", v)
		}
	} else {
		t.Errorf("failed to unmarshal foo")
	}

	// Aliased
	cfg.AddAlias("foo.nested.e", "foo.b")
	obj = map[string]interface{}{"a": nil,
		"nested": map[string]interface{}{"e": nil}}
	n1 = obj["nested"].(map[string]interface{})
	if err := cfg.UnmarshalToMap("foo", obj); err == nil {
		if obj["a"] != mr.config["foo.a"] {
			t.Errorf("failed to unmarshal 'foo.a', expected %v, got %v", mr.config["foo.a"], obj["a"])
		}
		if n1["e"] != mr.config["foo.b"] {
			t.Errorf("failed to unmarshal 'foo.nested.e', expected %#v, got %#v", mr.config["foo.b"], n1["e"])
		}
	} else {
		t.Errorf("failed to unmarshal foo")
	}
}
