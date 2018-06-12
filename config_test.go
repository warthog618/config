// Copyright © 2017 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package config_test

import (
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/warthog618/config/keys"

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
	mr := mockGetter{map[string]interface{}{
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

func TestNewWithDefault(t *testing.T) {
	def := mockGetter{map[string]interface{}{
		"a.b.c": 43,
	}}
	cfg := config.New(config.WithDefault(&def))
	c, err := cfg.Get("a.b.c")
	assert.Nil(t, err)
	assert.Equal(t, 43, c)

	// After WithGetters
	mr := mockGetter{map[string]interface{}{
		"a.b.c_d": true,
	}}
	cfg = config.New(
		config.WithGetters([]config.Getter{&mr}),
		config.WithDefault(&def))
	c, err = cfg.Get("a.b.c_d")
	assert.Nil(t, err)
	assert.Equal(t, true, c)
	c, err = cfg.Get("a.b.c")
	assert.Nil(t, err)
	assert.Equal(t, 43, c)

	// After WithGetters AND WithDefault
	def2 := mockGetter{map[string]interface{}{
		"a.b.d": 43,
	}}
	cfg = config.New(
		config.WithGetters([]config.Getter{&mr}),
		config.WithDefault(&def),
		config.WithDefault(&def2),
	)
	c, err = cfg.Get("a.b.c_d")
	assert.Nil(t, err)
	assert.Equal(t, true, c)
	c, err = cfg.Get("a.b.c")
	assert.IsType(t, config.NotFoundError{}, err)
	assert.Nil(t, c)
	c, err = cfg.Get("a.b.d")
	assert.Nil(t, err)
	assert.Equal(t, 43, c)

}

func TestNewWithGetters(t *testing.T) {
	mr := mockGetter{map[string]interface{}{
		"a.b.c_d": true,
	}}
	gg := []config.Getter{&mr}
	cfg := config.New(config.WithGetters(gg))
	c, err := cfg.Get("a.b.c_d")
	assert.Nil(t, err)
	assert.Equal(t, true, c)
	// show getters copied
	gg[0] = mockGetter{}
	c, err = cfg.Get("a.b.c_d")
	assert.Nil(t, err)
	assert.Equal(t, true, c)
	// show getter not copied
	mr.config["a.b.c_d"] = 43
	c, err = cfg.Get("a.b.c_d")
	assert.Nil(t, err)
	assert.Equal(t, 43, c)

	// after WithDefault
	def := mockGetter{map[string]interface{}{
		"def": true,
	}}
	cfg = config.New(
		config.WithDefault(&def),
		config.WithGetters([]config.Getter{&mr}))
	c, err = cfg.Get("a.b.c_d")
	assert.Nil(t, err)
	assert.Equal(t, 43, c)
	c, err = cfg.Get("def")
	assert.Nil(t, err)
	assert.Equal(t, true, c)

}

func TestNewWithKeyReplacer(t *testing.T) {
	mr := mockGetter{map[string]interface{}{
		"a.b.c_d": true,
	}}
	cfg := config.New(config.WithKeyReplacer(keys.LowerCaseReplacer()))
	cfg.AppendGetter(mr)
	c, err := cfg.Get("a.B.c_d")
	assert.Nil(t, err)
	assert.Equal(t, true, c)

	// subtree config
	ab, err := cfg.GetConfig("a.b")
	assert.Nil(t, err)
	require.NotNil(t, ab)
	c, err = ab.Get("C_d")
	assert.Nil(t, err)
	assert.Equal(t, true, c)

	// capped subtree config
	ab, err = cfg.GetConfig("A.b")
	assert.Nil(t, err)
	require.NotNil(t, ab)
	c, err = ab.Get("C_d")
	assert.Nil(t, err)
	assert.Equal(t, true, c)
}

func TestNewWithSeparator(t *testing.T) {
	cfg := config.New(config.WithSeparator("_"))
	c, err := cfg.Get("")
	assert.IsType(t, config.NotFoundError{}, err)
	assert.Equal(t, nil, c)
	// demonstrate nesting separation by "_"
	mr := mockGetter{map[string]interface{}{
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
	mr := mockGetter{map[string]interface{}{}}
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
	mr := mockGetter{map[string]interface{}{
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

	// It is fundamentally the responsibility of the application to manage the config tree
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
	mr1 := mockGetter{map[string]interface{}{}}
	cfg.AppendGetter(nil) // should be ignored
	cfg.InsertGetter(&mr1)
	mr1.config["something"] = "a test string"
	v, err := cfg.Get("something")
	assert.Nil(t, err)
	assert.Exactly(t, mr1.config["something"], v)

	// append a second reader
	mr2 := mockGetter{map[string]interface{}{
		"something":      "another test string",
		"something else": "yet another test string",
	}}
	cfg.AppendGetter(&mr2)
	v, err = cfg.Get("something")
	assert.Nil(t, err)
	assert.Exactly(t, mr1.config["something"], v)
	v, err = cfg.Get("something else")
	assert.Nil(t, err)
	assert.Exactly(t, mr2.config["something else"], v)

	// with a default getter
	def := mockGetter{map[string]interface{}{
		"something":      "a default string",
		"something else": "yet another test string",
	}}
	cfg = config.New(config.WithDefault(def))
	v, err = cfg.Get("something")
	assert.Nil(t, err)
	assert.Exactly(t, def.config["something"], v)
	cfg.InsertGetter(&mr1)
	cfg.AppendGetter(&mr2)
	v, err = cfg.Get("something")
	assert.Nil(t, err)
	assert.Exactly(t, mr1.config["something"], v)
	v, err = cfg.Get("something else")
	assert.Nil(t, err)
	assert.Exactly(t, mr2.config["something else"], v)
}

func TestInsertGetter(t *testing.T) {
	cfg := config.New()
	mr1 := mockGetter{map[string]interface{}{
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
	mr2 := mockGetter{map[string]interface{}{}}
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
	mr1 := mockGetter{map[string]interface{}{
		"a": "a - tier 1",
		"b": "b - tier 1",
		"c": "c - tier 1",
	}}
	mr2 := mockGetter{map[string]interface{}{
		"b": "b - tier 2",
		"d": "d - tier 2",
	}}
	mr3 := mockGetter{map[string]interface{}{
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
		readers  []mockGetter
		expected []kv
	}{
		{"one", []mockGetter{mr1}, []kv{
			{"a", "a - tier 1", nil},
			{"b", "b - tier 1", nil},
			{"c", "c - tier 1", nil},
			{"d", nil, config.NotFoundError{Key: "d"}},
			{"e", nil, config.NotFoundError{Key: "e"}},
		}},
		{"two", []mockGetter{mr1, mr2}, []kv{
			{"a", "a - tier 1", nil},
			{"b", "b - tier 2", nil},
			{"c", "c - tier 1", nil},
			{"d", "d - tier 2", nil},
			{"e", nil, config.NotFoundError{Key: "e"}},
		}},
		{"three", []mockGetter{mr1, mr2, mr3}, []kv{
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
	mr := mockGetter{map[string]interface{}{
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
	mr := mockGetter{map[string]interface{}{
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
	mr := mockGetter{map[string]interface{}{
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
	mr := mockGetter{map[string]interface{}{
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
	mr := mockGetter{map[string]interface{}{
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
	mr := mockGetter{map[string]interface{}{
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
	mr := mockGetter{map[string]interface{}{
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
	mr := mockGetter{map[string]interface{}{
		"slice": []interface{}{1, 2, 3, 4},
		"animals": []interface{}{
			map[string]interface{}{"Name": "Platypus", "Order": "Monotremata"},
			map[string]interface{}{"Order": "Dasyuromorphia", "Name": "Quoll"},
		},
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
		{"animals", []interface{}{
			map[string]interface{}{"Name": "Platypus", "Order": "Monotremata"},
			map[string]interface{}{"Order": "Dasyuromorphia", "Name": "Quoll"},
		}, nil},
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
	mr := mockGetter{map[string]interface{}{
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
	mr := mockGetter{map[string]interface{}{
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
	mr := mockGetter{map[string]interface{}{
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
	mr := mockGetter{map[string]interface{}{
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
	F       []innerConfig
	G       [][]int
	Nested  innerConfig `config:"nested"`
	p       string
}

type innerConfig struct {
	A       int
	Btagged string `config:"b_inner"`
	C       []int
	E       string
}

type configSetup func(*config.Config)

func TestUnmarshal(t *testing.T) {
	blah := "blah"
	patterns := []struct {
		name     string
		g        config.Getter
		s        configSetup
		k        string
		target   interface{}
		expected interface{}
		err      error
	}{
		{"non-struct target",
			&mockGetter{map[string]interface{}{
				"foo.a": 42,
			}},
			nil,
			"foo",
			&blah,
			&blah,
			config.ErrInvalidStruct},
		{"non-pointer target",
			&mockGetter{map[string]interface{}{
				"foo.a": 42,
			}},
			nil,
			"foo",
			fooConfig{},
			fooConfig{},
			config.ErrInvalidStruct},
		{"scalars",
			&mockGetter{map[string]interface{}{
				"a": 42,
				"b": "foo.b",
				"d": "ignored",
				"p": "non-exported fields can't be set",
			}},
			nil,
			"",
			&fooConfig{},
			&fooConfig{
				Atagged: 42,
				B:       "foo.b"},
			nil},
		{"maltyped",
			&mockGetter{map[string]interface{}{
				"a": []int{3, 4},
			}},
			nil,
			"",
			&fooConfig{},
			&fooConfig{},
			config.UnmarshalError{}},
		{"array of scalar",
			&mockGetter{map[string]interface{}{
				"c": []int{1, 2, 3, 4},
				"d": "ignored",
			}},
			nil,
			"",
			&fooConfig{},
			&fooConfig{
				C: []int{1, 2, 3, 4}},
			nil},
		{"array of array",
			&mockGetter{map[string]interface{}{
				"g": [][]int{{1, 2}, {3, 4}},
				"d": "ignored",
			}},
			nil,
			"",
			&fooConfig{},
			&fooConfig{
				G: [][]int{{1, 2}, {3, 4}}},
			nil},
		{"array of object",
			&mockGetter{map[string]interface{}{
				"foo.f": []map[string]interface{}{
					{"a": 1},
					{"a": 2},
				},
				"foo.d": "ignored",
			}},
			nil,
			"foo",
			&fooConfig{F: []innerConfig{{}, {}}},
			&fooConfig{
				F: []innerConfig{
					{A: 1},
					{A: 2},
				}},
			nil},
		{"nested",
			&mockGetter{map[string]interface{}{
				"foo.b":              "foo.b",
				"foo.nested.a":       43,
				"foo.nested.b_inner": "foo.nested.b",
				"foo.nested.c":       []int{5, 6, 7, 8},
			}},
			nil,
			"foo",
			&fooConfig{},
			&fooConfig{
				B: "foo.b",
				Nested: innerConfig{
					A:       43,
					Btagged: "foo.nested.b",
					C:       []int{5, 6, 7, 8}}},
			nil},
		{"nested wrong type",
			&mockGetter{map[string]interface{}{
				"foo.nested.a": []int{6, 7}},
			},
			nil,
			"foo",
			&fooConfig{},
			&fooConfig{},
			config.UnmarshalError{}},
		{"aliased",
			&mockGetter{map[string]interface{}{
				"foo.b": "foo.b",
			}},
			func(c *config.Config) {
				c.AddAlias("foo.e", "foo.b")
			},
			"foo",
			&fooConfig{},
			&fooConfig{
				B: "foo.b",
				E: "foo.b"},
			nil},
		{"nested aliased",
			&mockGetter{map[string]interface{}{
				"foo.b":              "foo.b",
				"foo.nested.b_inner": "foo.nested.b",
			}},
			func(c *config.Config) {
				c.AddAlias("foo.nested.e", "foo.b")
				// Alias to tagged name
				c.AddAlias("foo.e", "foo.nested.b_inner")
				// aliased alias ignored
				c.AddAlias("foo.nested.bTagged", "foo.nested.e")
			},
			"foo",
			&fooConfig{},
			&fooConfig{
				B: "foo.b",        // from value
				E: "foo.nested.b", // from alias
				Nested: innerConfig{
					Btagged: "foo.nested.b", // from value
					E:       "foo.b"}},      // from alias
			nil},
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			cfg := config.New()
			cfg.InsertGetter(p.g)
			if p.s != nil {
				p.s(cfg)
			}
			err := cfg.Unmarshal(p.k, p.target)
			assert.IsType(t, p.err, err)
			assert.Equal(t, p.expected, p.target)
		}
		t.Run(p.name, f)
	}
}

func TestUnmarshalToMap(t *testing.T) {
	mg := mockGetter{map[string]interface{}{
		"foo.a": 42,
		"foo.b": "foo.b",
		"foo.c": []int{1, 2, 3, 4},
		"foo.d": "ignored",
	}}
	patterns := []struct {
		name     string
		g        config.Getter
		s        configSetup
		target   map[string]interface{}
		expected map[string]interface{}
		err      error
	}{
		{"nil types",
			mg, nil,
			map[string]interface{}{"a": nil, "b": nil, "c": nil, "e": nil},
			map[string]interface{}{
				"a": 42,
				"b": "foo.b",
				"c": []int{1, 2, 3, 4},
				"e": nil},
			nil,
		},
		{"typed",
			mg, nil,
			map[string]interface{}{
				"a": int(0),
				"b": "",
				"c": []int{},
				"e": "some useful default"},
			map[string]interface{}{
				"a": 42,
				"b": "foo.b",
				"c": []int{1, 2, 3, 4},
				"e": "some useful default"},
			nil,
		},
		{"aliased",
			mg,
			func(c *config.Config) {
				c.AddAlias("foo.e", "foo.d")
			},
			map[string]interface{}{"e": nil},
			map[string]interface{}{"e": "ignored"},
			nil,
		},
		{"maltyped int",
			mg, nil,
			map[string]interface{}{"a": []int{0}},
			map[string]interface{}{"a": []int{0}},
			config.UnmarshalError{},
		},
		{"maltyped string",
			mg, nil,
			map[string]interface{}{"b": 2},
			map[string]interface{}{"b": 2},
			config.UnmarshalError{},
		},
		{"maltyped array",
			mg, nil,
			map[string]interface{}{"c": 3},
			map[string]interface{}{"c": 3},
			config.UnmarshalError{},
		},
		{"array of arrays",
			mockGetter{map[string]interface{}{
				"foo.aa": [][]int{{1, 2, 3, 4}, {4, 5, 6, 7}},
			}}, nil,
			map[string]interface{}{"aa": nil},
			map[string]interface{}{
				"aa": [][]int{{1, 2, 3, 4}, {4, 5, 6, 7}},
			},
			nil,
		},
		{"array of objects",
			&mockGetter{map[string]interface{}{
				"foo.f": []map[string]interface{}{
					{"A": 1},
					{"A": 2},
				},
				"foo.d": "ignored",
			}},
			nil,
			map[string]interface{}{"f": nil},
			map[string]interface{}{
				"f": []map[string]interface{}{
					{"A": 1},
					{"A": 2},
				},
			},
			nil,
		},
		{"nested",
			mockGetter{map[string]interface{}{
				"foo.a":        42,
				"foo.b":        "foo.b",
				"foo.c":        []int{1, 2, 3, 4},
				"foo.d":        "ignored",
				"foo.nested.a": 43,
				"foo.nested.b": "foo.nested.b",
				"foo.nested.c": []int{1, 2, -3, 4},
			}},
			nil,
			map[string]interface{}{
				"a": 0,
				"nested": map[string]interface{}{
					"a": int(0),
					"b": "",
					"c": []int{}},
			},
			map[string]interface{}{
				"a": 42,
				"nested": map[string]interface{}{
					"a": 43,
					"b": "foo.nested.b",
					"c": []int{1, 2, -3, 4}},
			},
			nil,
		},
		{"nested maltyped",
			mockGetter{map[string]interface{}{
				"foo.a":        42,
				"foo.b":        "foo.b",
				"foo.c":        []int{1, 2, 3, 4},
				"foo.d":        "ignored",
				"foo.nested.a": []int{},
				"foo.nested.b": "foo.nested.b",
				"foo.nested.c": []int{1, 2, -3, 4},
			}},
			nil,
			map[string]interface{}{
				"a": 0,
				"nested": map[string]interface{}{
					"a": int(0),
					"b": "",
					"c": []int{}},
			},
			map[string]interface{}{
				"a": 42,
				"nested": map[string]interface{}{
					"a": 0,
					"b": "foo.nested.b",
					"c": []int{1, 2, -3, 4}},
			},
			config.UnmarshalError{},
		},
		{"nested alias",
			mockGetter{map[string]interface{}{
				"foo.b": "foo.b",
				"foo.d": "ignored"},
			},
			func(c *config.Config) {
				c.AddAlias("foo.nested.e", "foo.d")
			},
			map[string]interface{}{
				"b": nil,
				"nested": map[string]interface{}{
					"e": nil},
			},
			map[string]interface{}{
				"b": "foo.b",
				"nested": map[string]interface{}{
					"e": "ignored"},
			},
			nil,
		},
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			cfg := config.New()
			cfg.InsertGetter(p.g)
			if p.s != nil {
				p.s(cfg)
			}
			err := cfg.UnmarshalToMap("foo", p.target)
			assert.IsType(t, p.err, err)
			assert.Equal(t, p.expected, p.target)
		}
		t.Run(p.name, f)
	}
}

// A simple mock Getter wrapping a map.

type mockGetter struct {
	config map[string]interface{}
}

func (m mockGetter) Get(key string) (interface{}, bool) {
	v, ok := m.config[key]
	return v, ok
}
