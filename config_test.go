// Copyright © 2017 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package config_test

import (
	"errors"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/warthog618/config"
	"github.com/warthog618/config/cfgconv"
)

func TestNew(t *testing.T) {
	cfg := config.New()
	c, err := cfg.Get("")
	assert.IsType(t, config.NotFoundError{}, err)
	assert.Equal(t, nil, c)
	// demonstrate nesting separation by "."
	mr := mapGetter{map[string]interface{}{
		"a.b.c_d": true,
	}}
	cfg.InsertGetter(&mr)
	c, err = cfg.Get("a.b.c_d")
	assert.Nil(t, err)
	assert.Equal(t, true, c)
	cfg.AddAlias("e", "a.b") // node alias uses "." nesting separator
	c, err = cfg.Get("e.c_d")
	assert.Nil(t, err)
	assert.Equal(t, true, c)
}

func TestNewWithSeparator(t *testing.T) {
	cfg := config.New(config.WithSeparator("_"))
	c, err := cfg.Get("")
	assert.IsType(t, config.NotFoundError{}, err)
	assert.Equal(t, nil, c)
	// demonstrate nesting separation by "_"
	mr := mapGetter{map[string]interface{}{
		"a.b.c_d": true,
	}}
	cfg.InsertGetter(&mr)
	c, err = cfg.Get("a.b.c_d")
	assert.Nil(t, err)
	assert.Equal(t, true, c)
	cfg.AddAlias("e", "a.b.c") // node alias uses "_" nesting separator
	c, err = cfg.Get("e_d")
	assert.Nil(t, err)
	assert.Equal(t, true, c)
}

func TestAddAlias(t *testing.T) {
	cfg := config.New()
	mr := mapGetter{map[string]interface{}{}}
	cfg.InsertGetter(&mr)

	// alias maps newKey (requested) -> oldKey (in config)
	mr.config["oldthing"] = "an old config string"
	cfg.AddAlias("newthing", "oldthing")
	v, err := cfg.Get("oldthing")
	assert.Nil(t, err)
	assert.Exactly(t, mr.config["oldthing"], v)
	v, err = cfg.Get("newthing")
	assert.Nil(t, err)
	assert.Exactly(t, mr.config["oldthing"], v)

	// alias ignored if newKey exists
	mr.config["newthing"] = "a new config string"
	v, err = cfg.Get("oldthing")
	assert.Nil(t, err)
	assert.Exactly(t, mr.config["oldthing"], v)
	v, err = cfg.Get("newthing")
	assert.Nil(t, err)
	assert.Exactly(t, mr.config["newthing"], v)
}

func TestAddAliasNested(t *testing.T) {
	mr := mapGetter{map[string]interface{}{
		"a":     "a",
		"foo.a": "foo.a",
		"foo.b": "foo.b",
		"bar.b": "bar.b",
		"bar.c": "bar.c",
	}}
	type alias struct {
		new string
		old string
	}
	aliases := []struct {
		name     string
		aa       []alias
		tp       string
		expected interface{}
		err      error
	}{
		{"alias to alias", []alias{{"c", "foo.b"}, {"d", "c"}}, "d", nil, config.NotFoundError{}},
		{"alias to node alias", []alias{{"baz", "bar"}, {"blob", "baz"}}, "blob.b", nil, config.NotFoundError{}},
		{"leaf alias has priority over node alias", []alias{{"baz", "bar"}, {"baz.b", "a"}}, "baz.b", "a", nil},
		{"leaf has priority over alias", []alias{{"a", "foo.a"}}, "a", "a", nil},
		{"nested leaf to root leaf", []alias{{"baz.b", "a"}}, "baz.b", "a", nil},
		{"nested leaf to self (ignored)", []alias{{"foo.a", "foo.a"}}, "foo.a", "foo.a", nil},
		{"nested node to nested node", []alias{{"baz", "bar"}}, "baz.b", "bar.b", nil},
		{"nested node to root node", []alias{{"node.a", "a"}}, "node.a", "a", nil},
		{"root leaf to nested leaf", []alias{{"c", "foo.b"}}, "c", "foo.b", nil},
		{"root node to nested node", []alias{{"", "foo"}}, "b", "foo.b", nil},
	}
	for _, a := range aliases {
		f := func(t *testing.T) {
			cfg := config.New()
			cfg.InsertGetter(&mr)
			for _, al := range a.aa {
				cfg.AddAlias(al.new, al.old)
			}
			v, err := cfg.Get(a.tp)
			assert.IsType(t, a.err, err)
			assert.Equal(t, a.expected, v)
		}
		t.Run(a.name, f)
	}
	// sub-tree config
	cfg := config.New()
	cfg.InsertGetter(&mr)
	barCfg, err := cfg.GetConfig("bar")
	assertGet(t, barCfg, "b", "bar.b", "sub-tree leaf")
	assertGet(t, barCfg, "c", "bar.c", "sub-tree leaf")
	barCfg.AddAlias("d", "c")
	assertGet(t, barCfg, "d", "bar.c", "sub-tree local leaf alias")

	barCfg.AddAlias("e", "b")
	refuteGet(t, cfg, "e", "sub-tree alias locality")
	refuteGet(t, cfg, "bar.e", "sub-tree alias locality")

	// aliased sub-tree config
	cfg.AddAlias("baz", "bar")
	cfg.AddAlias("baz.b", "a")
	bazCfg, err := cfg.GetConfig("baz")
	assert.Nil(t, err)
	require.NotNil(t, bazCfg)
	assertGet(t, bazCfg, "b", "a", "sub-tree node leaf alias")
	assertGet(t, bazCfg, "c", "bar.c", "sub-tree leaf")
	bazCfg.AddAlias("d", "c") // gets turned into baz.d -> baz.c which will fail as it is an alias to an alias
	refuteGet(t, bazCfg, "d", "aliased sub-tree local leaf alias")
	bazCfg.AddAlias("e", "b")

	refuteGet(t, cfg, "e", "sub-tree alias locality")
	refuteGet(t, cfg, "baz.e", "sub-tree alias locality")

	// Fundamentally is the responsibility of the application to manage the config tree
	// and setup the aliases for any included modules.
	// In the case above they need to add the baz.d -> bar.c mapping like this...
	cfg.AddAlias("baz.d", "bar.c")
	assertGet(t, cfg, "baz.d", "bar.c", "sub-tree leaf")
	bazCfg, err = cfg.GetConfig("baz")
	assert.Nil(t, err)
	require.NotNil(t, bazCfg)
	bazCfg.AddAlias("d", "c") // this is pointless as noted above, but the baz.g -> bar.c alias should still work...
	assertGet(t, bazCfg, "d", "bar.c", "sub-tree leaf alias")
}

func TestAppendGetter(t *testing.T) {
	cfg := config.New()
	mr1 := mapGetter{map[string]interface{}{}}
	cfg.AppendGetter(nil) // should be ignored
	cfg.InsertGetter(&mr1)
	mr1.config["something"] = "a test string"
	v, err := cfg.Get("something")
	assert.Nil(t, err)
	assert.Exactly(t, mr1.config["something"], v)

	// append a second reader
	mr2 := mapGetter{map[string]interface{}{}}
	cfg.AppendGetter(&mr2)
	mr2.config["something"] = "another test string"
	mr2.config["something else"] = "yet another test string"
	v, err = cfg.Get("something")
	assert.Nil(t, err)
	assert.Exactly(t, mr1.config["something"], v)
	v, err = cfg.Get("something else")
	assert.Nil(t, err)
	assert.Exactly(t, mr2.config["something else"], v)
}

func TestInsertGetter(t *testing.T) {
	cfg := config.New()
	mr1 := mapGetter{map[string]interface{}{
		"something":      "a test string",
		"something else": "yet another test string",
	}}
	cfg.InsertGetter(nil) // should be ignored
	cfg.InsertGetter(&mr1)
	v, err := cfg.Get("something")
	assert.Nil(t, err)
	assert.Exactly(t, mr1.config["something"], v)
	v, err = cfg.Get("something else")
	assert.Nil(t, err)
	assert.Exactly(t, mr1.config["something else"], v)

	// insert a second reader
	mr2 := mapGetter{map[string]interface{}{}}
	cfg.InsertGetter(&mr2)
	v, err = cfg.Get("something")
	assert.Nil(t, err)
	assert.Exactly(t, mr1.config["something"], v)
	mr2.config["something"] = "another test string"
	v, err = cfg.Get("something")
	assert.Nil(t, err)
	assert.Exactly(t, mr2.config["something"], v)
	v, err = cfg.Get("something else")
	assert.Nil(t, err)
	assert.Exactly(t, mr1.config["something else"], v)
}

func assertGet(t *testing.T, cfg *config.Config, key string, expected interface{}, comment string) {
	v, err := cfg.Get(key)
	assert.Nil(t, err)
	assert.Equal(t, expected, v, comment)
}

func refuteGet(t *testing.T, cfg *config.Config, key string, comment string) {
	v, err := cfg.Get(key)
	assert.IsType(t, config.NotFoundError{}, err, comment)
	assert.Equal(t, nil, v, comment)
}

func TestGetOverlayed(t *testing.T) {
	mr1 := mapGetter{map[string]interface{}{
		"a": "a - tier 1",
		"b": "b - tier 1",
		"c": "c - tier 1",
	}}
	mr2 := mapGetter{map[string]interface{}{
		"b": "b - tier 2",
		"d": "d - tier 2",
	}}
	mr3 := mapGetter{map[string]interface{}{
		"c": "c - tier 3",
		"d": "d - tier 3",
	}}
	type kv struct {
		k   string
		v   interface{}
		err error
	}
	patterns := []struct {
		name     string
		readers  []*mapGetter // !!! breaks test if value instead of pointer - why???
		expected []kv
	}{
		{"one", []*mapGetter{&mr1}, []kv{
			{"a", "a - tier 1", nil},
			{"b", "b - tier 1", nil},
			{"c", "c - tier 1", nil},
			{"d", nil, config.NotFoundError{Key: "d"}},
			{"e", nil, config.NotFoundError{Key: "e"}},
		}},
		{"two", []*mapGetter{&mr1, &mr2}, []kv{
			{"a", "a - tier 1", nil},
			{"b", "b - tier 2", nil},
			{"c", "c - tier 1", nil},
			{"d", "d - tier 2", nil},
			{"e", nil, config.NotFoundError{Key: "e"}},
		}},
		{"three", []*mapGetter{&mr1, &mr2, &mr3}, []kv{
			{"a", "a - tier 1", nil},
			{"b", "b - tier 2", nil},
			{"c", "c - tier 3", nil},
			{"d", "d - tier 3", nil},
			{"e", nil, config.NotFoundError{Key: "e"}},
		}},
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			cfg := config.New()
			for _, r := range p.readers {
				cfg.InsertGetter(r)
			}
			for _, x := range p.expected {
				v, err := cfg.Get(x.k)
				assert.Equal(t, x.err, err, x.k)
				assert.Equal(t, x.v, v, x.k)
			}
		}
		t.Run(p.name, f)
	}
}

func TestGetBool(t *testing.T) {
	cfg := config.New()
	mr := mapGetter{map[string]interface{}{
		"bool":       true,
		"boolString": "true",
		"boolInt":    1,
		"notabool":   "bogus",
	}}
	cfg.InsertGetter(&mr)
	patterns := []struct {
		k   string
		v   bool
		err error
	}{
		{"bool", true, nil},
		{"boolString", true, nil},
		{"boolInt", true, nil},
		{"notabool", false, &strconv.NumError{}},
		{"notsuchbool", false, config.NotFoundError{}},
	}
	for _, p := range patterns {
		v, err := cfg.GetBool(p.k)
		assert.IsType(t, p.err, err, p.k)
		assert.Equal(t, p.v, v, p.k)
	}
}

func TestGetDuration(t *testing.T) {
	cfg := config.New()
	mr := mapGetter{map[string]interface{}{
		"duration":     "123ms",
		"notaduration": "bogus",
	}}
	cfg.InsertGetter(&mr)
	patterns := []struct {
		k   string
		v   time.Duration
		err error
	}{
		{"duration", time.Duration(123000000), nil},
		{"notaduration", time.Duration(0), errors.New("")},
		{"nosuchduration", time.Duration(0), config.NotFoundError{}},
	}
	for _, p := range patterns {
		v, err := cfg.GetDuration(p.k)
		assert.IsType(t, p.err, err, p.k)
		assert.Equal(t, p.v, v, p.k)
	}
}

func TestGetFloat(t *testing.T) {
	cfg := config.New()
	mr := mapGetter{map[string]interface{}{
		"float":        3.1415,
		"floatString":  "3.1415",
		"floatInt":     1,
		"notafloatInt": "bogus",
	}}
	cfg.InsertGetter(&mr)
	patterns := []struct {
		k   string
		v   float64
		err error
	}{
		{"float", 3.1415, nil},
		{"floatString", 3.1415, nil},
		{"floatInt", 1, nil},
		{"notafloatInt", 0, &strconv.NumError{}},
		{"nosuchfloat", 0, config.NotFoundError{}},
	}
	for _, p := range patterns {
		v, err := cfg.GetFloat(p.k)
		assert.IsType(t, p.err, err, p.k)
		assert.Equal(t, p.v, v, p.k)
	}
}

func TestGetInt(t *testing.T) {
	cfg := config.New()
	mr := mapGetter{map[string]interface{}{
		"int":       42,
		"intString": "43",
		"notaint":   "bogus",
	}}
	cfg.InsertGetter(&mr)
	patterns := []struct {
		k   string
		v   int64
		err error
	}{
		{"int", 42, nil},
		{"intString", 43, nil},
		{"notaint", 0, &strconv.NumError{}},
		{"nosuchint", 0, config.NotFoundError{}},
	}
	for _, p := range patterns {
		v, err := cfg.GetInt(p.k)
		assert.IsType(t, p.err, err, p.k)
		assert.Equal(t, p.v, v, p.k)
	}
}

func TestGetString(t *testing.T) {
	cfg := config.New()
	mr := mapGetter{map[string]interface{}{
		"string":     "a string",
		"stringInt":  42,
		"notastring": struct{}{},
	}}
	cfg.InsertGetter(&mr)
	patterns := []struct {
		k   string
		v   string
		err error
	}{
		{"string", "a string", nil},
		{"stringInt", "42", nil},
		{"notastring", "", cfgconv.TypeError{}},
		{"nosuchstring", "", config.NotFoundError{}},
	}
	for _, p := range patterns {
		v, err := cfg.GetString(p.k)
		assert.IsType(t, p.err, err, p.k)
		assert.Equal(t, p.v, v, p.k)
	}
}

func TestGetTime(t *testing.T) {
	cfg := config.New()
	mr := mapGetter{map[string]interface{}{
		"time":     "2017-03-01T01:02:03Z",
		"notatime": "bogus",
	}}
	cfg.InsertGetter(&mr)
	patterns := []struct {
		k   string
		v   time.Time
		err error
	}{
		{"time", time.Date(2017, 3, 1, 1, 2, 3, 0, time.UTC), nil},
		{"notatime", time.Time{}, &time.ParseError{}},
		{"nosuchtime", time.Time{}, config.NotFoundError{}},
	}
	for _, p := range patterns {
		v, err := cfg.GetTime(p.k)
		assert.IsType(t, p.err, err, p.k)
		assert.Equal(t, p.v, v, p.k)
	}
}

func TestGetUint(t *testing.T) {
	cfg := config.New()
	mr := mapGetter{map[string]interface{}{
		"uint":       42,
		"uintString": "43",
		"notaUint":   "bogus",
	}}
	cfg.InsertGetter(&mr)
	patterns := []struct {
		k   string
		v   uint64
		err error
	}{
		{"uint", 42, nil},
		{"uintString", 43, nil},
		{"notaUint", 0, &strconv.NumError{}},
		{"nosuchUint", 0, config.NotFoundError{}},
	}
	for _, p := range patterns {
		v, err := cfg.GetUint(p.k)
		assert.IsType(t, p.err, err, p.k)
		assert.Equal(t, p.v, v, p.k)
	}
}

func TestGetSlice(t *testing.T) {
	cfg := config.New()
	mr := mapGetter{map[string]interface{}{
		"slice":       []interface{}{1, 2, 3, 4},
		"casttoslice": "bogus",
		"notaslice":   struct{}{},
	}}
	cfg.InsertGetter(&mr)
	patterns := []struct {
		k   string
		v   []interface{}
		err error
	}{
		{"slice", []interface{}{1, 2, 3, 4}, nil},
		{"casttoslice", []interface{}{"bogus"}, nil},
		{"notaslice", nil, cfgconv.TypeError{}},
		{"nosuchslice", nil, config.NotFoundError{}},
	}
	for _, p := range patterns {
		v, err := cfg.GetSlice(p.k)
		assert.IsType(t, p.err, err, p.k)
		assert.Equal(t, p.v, v, p.k)
	}
}

func TestGetIntSlice(t *testing.T) {
	cfg := config.New()
	mr := mapGetter{map[string]interface{}{
		"slice":       []int64{1, 2, -3, 4},
		"casttoslice": "42",
		"stringslice": []string{"one", "two", "three"},
		"notaslice":   "bogus",
	}}
	cfg.InsertGetter(&mr)
	patterns := []struct {
		k   string
		v   []int64
		err error
	}{
		{"slice", []int64{1, 2, -3, 4}, nil},
		{"casttoslice", []int64{42}, nil},
		{"stringslice", nil, &strconv.NumError{}},
		{"notaslice", nil, &strconv.NumError{}},
		{"nosuchslice", nil, config.NotFoundError{}},
	}
	for _, p := range patterns {
		v, err := cfg.GetIntSlice(p.k)
		assert.IsType(t, p.err, err, p.k)
		assert.Equal(t, p.v, v, p.k)
	}
}

func TestGetStringSlice(t *testing.T) {
	cfg := config.New()
	mr := mapGetter{map[string]interface{}{
		"intslice":        []int64{1, 2, -3, 4},
		"stringslice":     []string{"one", "two", "three"},
		"uintslice":       []uint64{1, 2, 3, 4},
		"notastringslice": []interface{}{1, 2, struct{}{}},
		"casttoslice":     "bogus",
		"notaslice":       struct{}{},
	}}
	cfg.InsertGetter(&mr)
	patterns := []struct {
		k   string
		v   []string
		err error
	}{
		{"intslice", []string{"1", "2", "-3", "4"}, nil},
		{"uintslice", []string{"1", "2", "3", "4"}, nil},
		{"stringslice", []string{"one", "two", "three"}, nil},
		{"casttoslice", []string{"bogus"}, nil},
		{"notastringslice", nil, cfgconv.TypeError{}},
		{"notaslice", nil, cfgconv.TypeError{}},
		{"nosuchslice", nil, config.NotFoundError{}},
	}
	for _, p := range patterns {
		v, err := cfg.GetStringSlice(p.k)
		assert.IsType(t, p.err, err, p.k)
		assert.Equal(t, p.v, v, p.k)
	}
}

func TestGetUintSlice(t *testing.T) {
	cfg := config.New()
	mr := mapGetter{map[string]interface{}{
		"slice":       []uint64{1, 2, 3, 4},
		"casttoslice": "42",
		"intslice":    []int64{1, 2, -3, 4},
		"stringslice": []string{"one", "two", "three"},
		"notaslice":   "bogus",
	}}
	cfg.InsertGetter(&mr)
	patterns := []struct {
		k   string
		v   []uint64
		err error
	}{
		{"slice", []uint64{1, 2, 3, 4}, nil},
		{"casttoslice", []uint64{42}, nil},
		{"intslice", nil, cfgconv.TypeError{}},
		{"stringslice", nil, &strconv.NumError{}},
		{"notaslice", nil, &strconv.NumError{}},
		{"nosuchslice", nil, config.NotFoundError{}},
	}
	for _, p := range patterns {
		v, err := cfg.GetUintSlice(p.k)
		assert.IsType(t, p.err, err, p.k)
		assert.Equal(t, p.v, v, p.k)
	}
}

func TestGetConfig(t *testing.T) {
	cfg := config.New()
	// Single Getter
	mr := mapGetter{map[string]interface{}{
		"foo.a": "foo.a",
		"foo.b": "foo.b",
		"bar.b": "bar.b",
		"bar.c": "bar.c",
		"baz.a": "baz.a",
		"baz.c": "baz.c",
	}}
	cfg.InsertGetter(&mr)
	cfg.AddAlias("foo.d", "bar.c") // leaf alias
	cfg.AddAlias("fuz", "foo")     // node alias

	type testPoint struct {
		k   string
		v   interface{}
		err error
	}
	type alias struct {
		new string
		old string
	}
	patterns := []struct {
		name    string
		subtree string
		tp      []testPoint
	}{
		{"root", "", []testPoint{
			{"foo.a", "foo.a", nil},
			{"foo.b", "foo.b", nil},
			{"bar.b", "bar.b", nil},
			{"bar.c", "bar.c", nil},
			{"baz.a", "baz.a", nil},
			{"baz.c", "baz.c", nil},
		}},
		{"foo", "foo", []testPoint{
			{"a", "foo.a", nil},
			{"b", "foo.b", nil},
			{"c", nil, config.NotFoundError{}},
			{"d", "bar.c", nil},
		}},
		{"fuz", "fuz", []testPoint{
			{"a", "foo.a", nil},
			{"b", "foo.b", nil},
			{"c", nil, config.NotFoundError{}},
			{"d", nil, config.NotFoundError{}}, // ignores foo.d -> bar.c alias
		}},
		{"bar", "bar", []testPoint{
			{"a", nil, config.NotFoundError{}},
			{"b", "bar.b", nil},
			{"c", "bar.c", nil},
			{"e", nil, config.NotFoundError{}},
		}},
		{"baz", "baz", []testPoint{
			{"a", "baz.a", nil},
			{"b", nil, config.NotFoundError{}},
			{"c", "baz.c", nil},
			{"e", nil, config.NotFoundError{}},
		}},
		{"blah", "blah", []testPoint{
			{"a", nil, config.NotFoundError{}},
			{"b", nil, config.NotFoundError{}},
			{"c", nil, config.NotFoundError{}},
			{"e", nil, config.NotFoundError{}},
		}},
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			for _, tp := range p.tp {
				c, err := cfg.GetConfig(p.subtree)
				assert.Nil(t, err)
				require.NotNil(t, c)
				v, err := c.Get(tp.k)
				assert.IsType(t, tp.err, err, tp.k)
				assert.Equal(t, tp.v, v, tp.k)
			}
		}
		t.Run(p.name, f)
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
	cfg := config.New()
	// Root Getter
	mr := mapGetter{map[string]interface{}{
		"foo.a": 42,
		"foo.b": "foo.b",
		"foo.c": []int{1, 2, 3, 4},
		"foo.d": "ignored",
	}}
	cfg.InsertGetter(&mr)
	if err := cfg.Unmarshal("foo", 0); err == nil {
		t.Errorf("failed to reject unmarshal into non-struct")
	}
	foo := fooConfig{}
	foo.E = "some useful default"
	// correctly typed
	if err := cfg.Unmarshal("foo", &foo); err == nil {
		assert.Equal(t, mr.config["foo.a"], foo.Atagged)
		assert.Equal(t, mr.config["foo.b"], foo.B)
		assert.Equal(t, mr.config["foo.c"], foo.C)
		assert.Equal(t, "some useful default", foo.E)
	} else {
		t.Errorf("failed to unmarshal foo")
	}
	// Incorrectly typed
	mr.config["foo.a"] = []int{1, 2}
	foo = fooConfig{}
	if err := cfg.Unmarshal("foo", &foo); err != nil {
		assert.IsType(t, config.UnmarshalError{}, err)
		assert.Contains(t, err.Error(), "foo.a")
		assert.Equal(t, 0, foo.Atagged)
		assert.Equal(t, mr.config["foo.b"], foo.B)
		assert.Equal(t, mr.config["foo.c"], foo.C)
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
		assert.Equal(t, mr.config["foo.a"], nc.Atagged)
		assert.Equal(t, mr.config["foo.b"], nc.B)
		assert.Equal(t, mr.config["foo.c"], nc.C)
		assert.Equal(t, mr.config["foo.nested.a"], nc.Nested.A)
		assert.Equal(t, mr.config["foo.nested.b"], nc.Nested.Btagged)
		assert.Equal(t, mr.config["foo.nested.c"], nc.Nested.C)
	} else {
		t.Errorf("failed to unmarshal foo")
	}

	// Nested incorrectly typed
	mr.config["foo.nested.a"] = []int{}
	nc = nestedConfig{}
	if err := cfg.Unmarshal("foo", &nc); err != nil {
		assert.IsType(t, config.UnmarshalError{}, err)
		assert.Contains(t, err.Error(), "foo.nested.a")
		assert.Equal(t, mr.config["foo.a"], nc.Atagged)
		assert.Equal(t, mr.config["foo.b"], nc.B)
		assert.Equal(t, mr.config["foo.c"], nc.C)
		assert.Equal(t, 0, nc.Nested.A)
		assert.Equal(t, mr.config["foo.nested.b"], nc.Nested.Btagged)
		assert.Equal(t, mr.config["foo.nested.c"], nc.Nested.C)
	} else {
		t.Errorf("successfully unmarshalled mistyped foo.nested")
	}
	mr.config["foo.nested.a"] = 43

	// Aliased
	mr.config["foo.b"] = "foo.b"
	cfg.AddAlias("foo.nested.e", "foo.b")
	nc = nestedConfig{}
	if err := cfg.Unmarshal("foo", &nc); err == nil {
		assert.Equal(t, mr.config["foo.b"], nc.Nested.E)
	} else {
		t.Errorf("failed to unmarshal foo")
	}
}

func TestUnmarshalToMap(t *testing.T) {
	cfg := config.New()
	// Root Getter
	mr := mapGetter{map[string]interface{}{
		"foo.a": 42,
		"foo.b": "foo.b",
		"foo.c": []int{1, 2, 3, 4},
		"foo.d": "ignored",
	}}
	cfg.InsertGetter(&mr)
	// Nil - raw
	obj := map[string]interface{}{"a": nil, "b": nil, "c": nil, "e": nil}
	if err := cfg.UnmarshalToMap("foo", obj); err == nil {
		assert.Equal(t, mr.config["foo.a"], obj["a"])
		assert.Equal(t, mr.config["foo.b"], obj["b"])
		assert.Equal(t, mr.config["foo.c"], obj["c"])
		v, ok := obj["d"]
		assert.False(t, ok)
		assert.Nil(t, v)
		assert.Nil(t, obj["e"])
	} else {
		t.Errorf("failed to unmarshal foo")
	}
	// Typed
	obj = map[string]interface{}{"a": int(0), "b": "", "c": []int{}, "e": "some useful default"}
	if err := cfg.UnmarshalToMap("foo", obj); err == nil {
		assert.Equal(t, mr.config["foo.a"], obj["a"])
		assert.Equal(t, mr.config["foo.b"], obj["b"])
		assert.Equal(t, mr.config["foo.c"], obj["c"])
		v, ok := obj["d"]
		assert.False(t, ok)
		assert.Nil(t, v)
		assert.Equal(t, "some useful default", obj["e"])
	} else {
		t.Errorf("failed to unmarshal foo")
	}
	// Mistyped
	obj = map[string]interface{}{"a": []int{}}
	err := cfg.UnmarshalToMap("foo", obj)
	assert.IsType(t, config.UnmarshalError{}, err)
	assert.Contains(t, err.Error(), "foo.a")

	obj = map[string]interface{}{"b": 44}
	err = cfg.UnmarshalToMap("foo", obj)
	assert.IsType(t, config.UnmarshalError{}, err)
	assert.Contains(t, err.Error(), "foo.b")

	obj = map[string]interface{}{"c": ""}
	err = cfg.UnmarshalToMap("foo", obj)
	assert.IsType(t, config.UnmarshalError{}, err)
	assert.Contains(t, err.Error(), "foo.c")

	// Nested
	mr.config["foo.nested.a"] = 43
	mr.config["foo.nested.b"] = "foo.nested.b"
	mr.config["foo.nested.c"] = []int{1, 2, -3, 4}
	obj = map[string]interface{}{"a": nil,
		"nested": map[string]interface{}{"a": int(0), "b": "", "c": []int{}}}
	n1 := obj["nested"].(map[string]interface{})
	if err := cfg.UnmarshalToMap("foo", obj); err == nil {
		assert.Equal(t, mr.config["foo.a"], obj["a"])
		assert.Equal(t, mr.config["foo.nested.a"], n1["a"])
		assert.Equal(t, mr.config["foo.nested.b"], n1["b"])
		assert.Equal(t, mr.config["foo.nested.c"], n1["c"])
		v, ok := n1["d"]
		assert.False(t, ok)
		assert.Nil(t, v)
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
		assert.IsType(t, config.UnmarshalError{}, err)
		assert.Contains(t, err.Error(), "foo.nested.a")
		assert.Equal(t, mr.config["foo.a"], obj["a"])
		assert.Equal(t, 0, n1["a"])
		assert.Equal(t, mr.config["foo.nested.b"], n1["b"])
		assert.Equal(t, mr.config["foo.nested.c"], n1["c"])
		v, ok := n1["d"]
		assert.False(t, ok)
		assert.Nil(t, v)
	} else {
		t.Errorf("failed to unmarshal foo")
	}

	// Aliased
	cfg.AddAlias("foo.nested.e", "foo.b")
	obj = map[string]interface{}{"a": nil,
		"nested": map[string]interface{}{"e": nil}}
	n1 = obj["nested"].(map[string]interface{})
	if err := cfg.UnmarshalToMap("foo", obj); err == nil {
		assert.Equal(t, mr.config["foo.a"], obj["a"])
		assert.Equal(t, mr.config["foo.b"], n1["e"])
	} else {
		t.Errorf("failed to unmarshal foo")
	}
}

type mapGetter struct {
	// simple key value map.
	// Note keys must be added as lowercase for config.GetX to work.
	config map[string]interface{}
}

func (mr *mapGetter) Get(key string) (interface{}, bool) {
	v, ok := mr.config[key]
	return v, ok
}

func TestNotFoundError(t *testing.T) {
	patterns := []string{"one", "two", "three"}
	for _, p := range patterns {
		f := func(t *testing.T) {
			e := config.NotFoundError{Key: p}
			expected := "config: key '" + e.Key + "' not found"
			assert.Equal(t, expected, e.Error())
		}
		t.Run(fmt.Sprintf("%x", p), f)
	}
}

func TestUnmarshalError(t *testing.T) {
	patterns := []struct {
		k   string
		err error
	}{
		{"one", errors.New("two")},
		{"three", errors.New("four")},
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			e := config.UnmarshalError{Key: p.k, Err: p.err}
			expected := "config: cannot unmarshal " + e.Key + " - " + e.Err.Error()
			assert.Equal(t, expected, e.Error())
		}
		t.Run(p.k, f)
	}
}
