// Copyright © 2017 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package cfgconv_test

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/warthog618/config/cfgconv"
)

func TestBool(t *testing.T) {
	patterns := []struct {
		name string
		in   interface{}
		v    bool
		err  error
	}{
		{"bool false", false, false, nil},
		{"bool true", true, true, nil},
		{"empty string", "", false, &strconv.NumError{}},
		{"float32 0", float32(0), false, cfgconv.TypeError{}},
		{"float32 1", float32(1), false, cfgconv.TypeError{}},
		{"float64 0", float64(0), false, cfgconv.TypeError{}},
		{"float64 1", float64(1), false, cfgconv.TypeError{}},
		{"int 0", int(0), false, nil},
		{"int 1", int(1), true, nil},
		{"int negative", int(-42), true, nil},
		{"int positive", int(42), true, nil},
		{"int16 0", int16(0), false, nil},
		{"int16 1", int16(1), true, nil},
		{"int32 0", int32(0), false, nil},
		{"int32 1", int32(1), true, nil},
		{"int64 0", int64(0), false, nil},
		{"int64 1", int64(1), true, nil},
		{"int8 0", int8(0), false, nil},
		{"int8 1", int8(1), true, nil},
		{"nil", nil, false, nil},
		{"string 0", "0", false, nil},
		{"string 1", "1", true, nil},
		{"string false", "false", false, nil},
		{"string junk", "junk", false, &strconv.NumError{}},
		{"string true", "true", true, nil},
		{"uint 0", uint(0), false, nil},
		{"uint 1", uint(1), true, nil},
		{"uint positive", uint(42), true, nil},
		{"uint16 0", uint16(0), false, nil},
		{"uint16 1", uint16(1), true, nil},
		{"uint32 0", uint32(0), false, nil},
		{"uint32 1", uint32(1), true, nil},
		{"uint64 0", uint64(0), false, nil},
		{"uint64 1", uint64(1), true, nil},
		{"uint8 0", uint8(0), false, nil},
		{"uint8 1", uint8(1), true, nil},
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			v, err := cfgconv.Bool(p.in)
			assert.IsType(t, p.err, err)
			assert.Equal(t, p.v, v)
		}
		t.Run(p.name, f)
	}
}

func TestConvert(t *testing.T) {
	type aStruct struct {
		A int
	}
	patterns := []struct {
		name string
		t    interface{}
		in   interface{}
		v    interface{}
		err  error
	}{
		{"bool bad type", true, []int{}, false, cfgconv.TypeError{}},
		{"bool good", true, 42, true, nil},
		{"bool parse error", true, "glob", false, &strconv.NumError{}},
		{"duration bad type", time.Duration(0), []string{"1", "2", "3"}, time.Duration(0), cfgconv.TypeError{}},
		{"duration good", time.Duration(0), "250ms", time.Duration(250000000), nil},
		{"duration parse error", time.Duration(0), "glob", time.Duration(0), errors.New("")},
		{"float32 bad", float32(0), []int{}, float32(0), cfgconv.TypeError{}},
		{"float32 good", float32(0), "42", float32(42), nil},
		{"float32 overflow", float32(0), float64(340282356779733642748073463979561713664), float32(0), cfgconv.OverflowError{}},
		{"float32 parse error", float32(0), "glob", float32(0), &strconv.NumError{}},
		{"float64 bad", float64(0), []int{}, float64(0), cfgconv.TypeError{}},
		{"float64 good", float64(0), "42", float64(42), nil},
		{"float64 parse error", float64(0), "glob", float64(0), &strconv.NumError{}},
		{"int bad", 0, []int{}, 0, cfgconv.TypeError{}},
		{"int good", 0, "42", 42, nil},
		{"int overflow", 0, "123456789123456789123456789", 0, &strconv.NumError{}},
		{"int parse error", 0, "glob", 0, &strconv.NumError{}},
		{"int16 bad", int16(0), []int{}, int16(0), cfgconv.TypeError{}},
		{"int16 good", int16(0), "42", int16(42), nil},
		{"int16 overflow", int16(0), 32768, int16(0), cfgconv.OverflowError{}},
		{"int16 parse error", int16(0), "glob", int16(0), &strconv.NumError{}},
		{"int32 bad", int32(0), []int{}, int32(0), cfgconv.TypeError{}},
		{"int32 good", int32(0), "42", int32(42), nil},
		{"int32 overflow", int32(0), int64(2147483648), int32(0), cfgconv.OverflowError{}},
		{"int32 parse error", int32(0), "glob", int32(0), &strconv.NumError{}},
		{"int64 bad", int64(0), []int{}, int64(0), cfgconv.TypeError{}},
		{"int64 good", int64(0), "42", int64(42), nil},
		{"int64 overflow", int64(0), "123456789123456789123456789", int64(0), &strconv.NumError{}},
		{"int64 parse error", int64(0), "glob", int64(0), &strconv.NumError{}},
		{"int8 bad", int8(0), []int{}, int8(0), cfgconv.TypeError{}},
		{"int8 good", int8(0), "42", int8(42), nil},
		{"int8 overflow", int8(0), 137, int8(0), cfgconv.OverflowError{}},
		{"int8 parse error", int8(0), "glob", int8(0), &strconv.NumError{}},
		{"slice bad type", []int{}, 42, []int(nil), cfgconv.TypeError{}},
		{"slice parse error string", []int{}, "glob", []int(nil), &strconv.NumError{}},
		{"slice parse error", []int{}, []string{"1", "2", "3", "glob"}, []int(nil), &strconv.NumError{}},
		{"slice slice", []int{}, []string{"1", "2", "3"}, []int{1, 2, 3}, nil},
		{"slice if slice", []interface{}{}, []int{1, 2, 3}, []interface{}{1, 2, 3}, nil},
		{"if good", interface{}(nil), 1, 1, nil},
		{"slice string", []int{}, "42", []int{42}, nil},
		{"string bad", "", []int{}, "", cfgconv.TypeError{}},
		{"string good", "", 42, "42", nil},
		{"time bad type", time.Time{}, []string{"1", "2", "3"}, time.Time{}, cfgconv.TypeError{}},
		{"time good", time.Time{}, "2017-03-01T01:02:03Z", time.Date(2017, 3, 1, 1, 2, 3, 0, time.UTC), nil},
		{"time parse error", time.Time{}, "2017", time.Time{}, &time.ParseError{}},
		{"uint bad", uint(0), []int{}, uint(0), cfgconv.TypeError{}},
		{"uint good", uint(0), "42", uint(42), nil},
		{"uint negative", uint(0), -1, uint(0), cfgconv.TypeError{}},
		{"uint overflow", uint(0), "123456789123456789123456789", uint(0), &strconv.NumError{}},
		{"uint parse error", uint(0), "glob", uint(0), &strconv.NumError{}},
		{"uint16 bad", uint16(0), []int{}, uint16(0), cfgconv.TypeError{}},
		{"uint16 good", uint16(0), "42", uint16(42), nil},
		{"uint16 overflow", uint16(0), 65537, uint16(0), cfgconv.OverflowError{}},
		{"uint16 parse error", uint16(0), "glob", uint16(0), &strconv.NumError{}},
		{"uint32 bad", uint32(0), []int{}, uint32(0), cfgconv.TypeError{}},
		{"uint32 good", uint32(0), "42", uint32(42), nil},
		{"uint32 overflow", uint32(0), int64(4294967296), uint32(0), cfgconv.OverflowError{}},
		{"uint32 parse error", uint32(0), "glob", uint32(0), &strconv.NumError{}},
		{"uint64 bad", uint64(0), []int{}, uint64(0), cfgconv.TypeError{}},
		{"uint64 good", uint64(0), "42", uint64(42), nil},
		{"uint64 overflow", uint64(0), "123456789123456789123456789", uint64(0), &strconv.NumError{}},
		{"uint64 parse error", uint64(0), "glob", uint64(0), &strconv.NumError{}},
		{"uint8 bad", uint8(0), []int{}, uint8(0), cfgconv.TypeError{}},
		{"uint8 good", uint8(0), "42", uint8(42), nil},
		{"uint8 overflow", uint8(0), 257, uint8(0), cfgconv.OverflowError{}},
		{"uint8 parse error", uint8(0), "glob", uint8(0), &strconv.NumError{}},
		{"struct lower", aStruct{}, map[string]interface{}{"a": 1}, aStruct{1}, nil},
		{"struct upper", aStruct{}, map[string]interface{}{"A": 1}, aStruct{}, nil},
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			v, err := cfgconv.Convert(p.in, reflect.TypeOf(p.t))
			assert.IsType(t, p.err, err)
			assert.Equal(t, p.v, v)
		}
		t.Run(p.name, f)
	}
}

func TestDuration(t *testing.T) {
	patterns := []struct {
		name string
		in   interface{}
		v    time.Duration
		err  error
	}{
		{"12ms bytes", []byte("12ms"), time.Duration(12000000), nil},
		{"250 no units", "250", 0, errors.New("")},
		{"250ms string", "250ms", time.Duration(250000000), nil},
		{"bad format bytes", []byte("glob"), 0, errors.New("")},
		{"bad format string", "foo", 0, errors.New("")},
		{"bad type int", 34, 0, cfgconv.TypeError{}},
		{"empty bytes", []byte{}, 0, errors.New("")},
		{"empty string", "", 0, errors.New("")},
		{"no unit bytes", []byte("250"), 0, errors.New("")},
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			v, err := cfgconv.Duration(p.in)
			assert.IsType(t, p.err, err)
			assert.Equal(t, p.v, v)
		}
		t.Run(p.name, f)
	}
}

func TestFloat(t *testing.T) {
	pi := float64(3.1415)
	pi32 := float32(3.1415)
	patterns := []struct {
		name string
		in   interface{}
		v    float64
		err  error
	}{
		{"float64 pi", pi, pi, nil},
		{"float32 pi", pi32, float64(pi32), nil},
		{"bool false", false, 0, nil},
		{"bool true", true, 1, nil},
		{"string pi", "3.1415", pi, nil},
		{"string int", "42", 42, nil},
		{"string int negative", "-42", -42, nil},
		{"int", int(42), 42, nil},
		{"int negative", int(-42), -42, nil},
		{"uint", uint(42), 42, nil},
		{"int8", int8(42), 42, nil},
		{"int8 negative", int8(-42), -42, nil},
		{"uint8", uint8(42), 42, nil},
		{"int16", int16(42), 42, nil},
		{"int16 negative", int16(-42), -42, nil},
		{"uint16", uint16(42), 42, nil},
		{"int32", int32(42), 42, nil},
		{"int32 negative", int32(-42), -42, nil},
		{"uint32", uint32(42), 42, nil},
		{"int64", int64(42), 42, nil},
		{"int64 negative", int64(-42), -42, nil},
		{"uint64", uint64(42), 42, nil},
		{"nil", nil, 0, nil},
		{"string junk", "junk", 0, &strconv.NumError{}},
		{"empty", "", 0, &strconv.NumError{}},
		{"slice", []int{42}, 0, cfgconv.TypeError{}},
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			v, err := cfgconv.Float(p.in)
			assert.IsType(t, p.err, err)
			assert.Equal(t, p.v, v)
		}
		t.Run(p.name, f)
	}
}

func TestInt(t *testing.T) {
	patterns := []struct {
		name string
		in   interface{}
		v    int64
		err  error
	}{
		{"bool false", false, 0, nil},
		{"bool true", true, 1, nil},
		{"string int", "42", 42, nil},
		{"string int negative", "-42", -42, nil},
		{"int", int(42), 42, nil},
		{"int negative", int(-42), -42, nil},
		{"uint", uint(42), 42, nil},
		{"int8", int8(42), 42, nil},
		{"int8 negative", int8(-42), -42, nil},
		{"uint8", uint8(42), 42, nil},
		{"int16", int16(42), 42, nil},
		{"int16 negative", int16(-42), -42, nil},
		{"uint16", uint16(42), 42, nil},
		{"int32", int32(42), 42, nil},
		{"int32 negative", int32(-42), -42, nil},
		{"uint32", uint32(42), 42, nil},
		{"int64", int64(42), 42, nil},
		{"int64 negative", int64(-42), -42, nil},
		{"uint64", uint64(42), 42, nil},
		{"float64", float64(42), 42, nil},
		{"float64 zero", float64(0), 0, nil},
		{"float64 negative", float64(-42), -42, nil},
		{"float64 truncate", float64(42.6), 42, nil},
		{"float64 truncate negative", float64(-42.6), -42, nil},
		{"float32", float32(42), 42, nil},
		{"float32 negative", float32(-42), -42, nil},
		{"float32 truncate", float32(42.6), 42, nil},
		{"float32 truncate negative", float32(-42.6), -42, nil},
		{"nil", nil, 0, nil},
		{"string float", "42.5", 0, &strconv.NumError{}},
		{"empty string", "", 0, &strconv.NumError{}},
		{"string junk", "junk", 0, &strconv.NumError{}},
		{"slice", []int{42}, 0, cfgconv.TypeError{}},
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			v, err := cfgconv.Int(p.in)
			assert.IsType(t, p.err, err)
			assert.Equal(t, p.v, v)
		}
		t.Run(p.name, f)
	}
}

func TestIntSlice(t *testing.T) {
	slice := []interface{}{[]int{1, 2, 3}}
	intSlice := []int{1, 2, -3}
	stringSlice := []string{"one", "two"}
	uintSlice := []int{1, 2, 3}
	patterns := []struct {
		name string
		in   interface{}
		v    []int64
		err  error
	}{
		{"slice", slice, []int64{0}, cfgconv.TypeError{}},
		{"intSlice", intSlice, []int64{1, 2, -3}, nil},
		{"stringSlice", stringSlice, []int64{0, 0}, &strconv.NumError{}},
		{"uintSlice", uintSlice, []int64{1, 2, 3}, nil},
		{"string int", "42", []int64{42}, nil},
		{"bool true", true, nil, cfgconv.TypeError{}},
		{"bool false", false, nil, cfgconv.TypeError{}},
		{"int", int(42), nil, cfgconv.TypeError{}},
		{"int negative", int(-42), nil, cfgconv.TypeError{}},
		{"uint", uint(42), nil, cfgconv.TypeError{}},
		{"int8", int8(42), nil, cfgconv.TypeError{}},
		{"int8 negative", int8(-42), nil, cfgconv.TypeError{}},
		{"uint8", uint8(42), nil, cfgconv.TypeError{}},
		{"int16", int16(42), nil, cfgconv.TypeError{}},
		{"int16 negative", int16(-42), nil, cfgconv.TypeError{}},
		{"uint16", uint16(42), nil, cfgconv.TypeError{}},
		{"int32", int32(42), nil, cfgconv.TypeError{}},
		{"int32 negative", int32(-42), nil, cfgconv.TypeError{}},
		{"uint32", uint32(42), nil, cfgconv.TypeError{}},
		{"int64", int64(42), nil, cfgconv.TypeError{}},
		{"int64 negative", int64(-42), nil, cfgconv.TypeError{}},
		{"uint64", uint64(42), nil, cfgconv.TypeError{}},
		{"float64", float64(42), nil, cfgconv.TypeError{}},
		{"float64 zero", float64(0), nil, cfgconv.TypeError{}},
		{"float64 negative", float64(-42), nil, cfgconv.TypeError{}},
		{"float64 truncate", float64(42.6), nil, cfgconv.TypeError{}},
		{"float64 truncate negative", float64(-42.6), nil, cfgconv.TypeError{}},
		{"float32", float32(42), nil, cfgconv.TypeError{}},
		{"float32 negative", float32(-42), nil, cfgconv.TypeError{}},
		{"float32 truncate", float32(42.6), nil, cfgconv.TypeError{}},
		{"float32 truncate negative", float32(-42.6), nil, cfgconv.TypeError{}},
		{"empty string", "", nil, cfgconv.TypeError{}},
		{"nil", nil, nil, cfgconv.TypeError{}},
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			v, err := cfgconv.IntSlice(p.in)
			assert.IsType(t, p.err, err)
			assert.Equal(t, p.v, v)
		}
		t.Run(p.name, f)
	}
}

func TestSlice(t *testing.T) {
	slice := []interface{}{[]int{1, 2, 3}}
	intSlice := []int{1, 2, -3}
	stringSlice := []string{"one", "two"}
	uintSlice := []int{1, 2, 3}
	patterns := []struct {
		name string
		in   interface{}
		v    []interface{}
		err  error
	}{
		{"slice", slice, slice, nil},
		{"intSlice", intSlice, []interface{}{1, 2, -3}, nil},
		{"stringSlice", stringSlice, []interface{}{"one", "two"}, nil},
		{"uintSlice", uintSlice, []interface{}{1, 2, 3}, nil},
		{"string int", "42", []interface{}{"42"}, nil},
		{"bool true", true, nil, cfgconv.TypeError{}},
		{"bool false", false, nil, cfgconv.TypeError{}},
		{"int", int(42), nil, cfgconv.TypeError{}},
		{"int negative", int(-42), nil, cfgconv.TypeError{}},
		{"uint", uint(42), nil, cfgconv.TypeError{}},
		{"int8", int8(42), nil, cfgconv.TypeError{}},
		{"int8 negative", int8(-42), nil, cfgconv.TypeError{}},
		{"uint8", uint8(42), nil, cfgconv.TypeError{}},
		{"int16", int16(42), nil, cfgconv.TypeError{}},
		{"int16 negative", int16(-42), nil, cfgconv.TypeError{}},
		{"uint16", uint16(42), nil, cfgconv.TypeError{}},
		{"int32", int32(42), nil, cfgconv.TypeError{}},
		{"int32 negative", int32(-42), nil, cfgconv.TypeError{}},
		{"uint32", uint32(42), nil, cfgconv.TypeError{}},
		{"int64", int64(42), nil, cfgconv.TypeError{}},
		{"int64 negative", int64(-42), nil, cfgconv.TypeError{}},
		{"uint64", uint64(42), nil, cfgconv.TypeError{}},
		{"float64", float64(42), nil, cfgconv.TypeError{}},
		{"float64 zero", float64(0), nil, cfgconv.TypeError{}},
		{"float64 negative", float64(-42), nil, cfgconv.TypeError{}},
		{"float64 truncate", float64(42.6), nil, cfgconv.TypeError{}},
		{"float64 truncate negative", float64(-42.6), nil, cfgconv.TypeError{}},
		{"float32", float32(42), nil, cfgconv.TypeError{}},
		{"float32 negative", float32(-42), nil, cfgconv.TypeError{}},
		{"float32 truncate", float32(42.6), nil, cfgconv.TypeError{}},
		{"float32 truncate negative", float32(-42.6), nil, cfgconv.TypeError{}},
		{"empty string", "", nil, cfgconv.TypeError{}},
		{"nil", nil, nil, cfgconv.TypeError{}},
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			v, err := cfgconv.Slice(p.in)
			assert.IsType(t, p.err, err)
			assert.Equal(t, p.v, v)
		}
		t.Run(p.name, f)
	}
}

func TestString(t *testing.T) {
	patterns := []struct {
		name string
		in   interface{}
		v    string
		err  error
	}{
		{"bool false", false, "false", nil},
		{"bool true", true, "true", nil},
		{"string junk", "junk", "junk", nil},
		{"empty string", "", "", nil},
		{"string int", "42", "42", nil},
		{"string int negative", "-42", "-42", nil},
		{"string float", "42.5", "42.5", nil},
		{"byte slice", []byte("1234"), "1234", nil},
		{"int", int(42), "42", nil},
		{"int negative", int(-42), "-42", nil},
		{"uint", uint(42), "42", nil},
		{"int8", int8(42), "42", nil},
		{"int8 negative", int8(-42), "-42", nil},
		{"uint8", uint8(42), "42", nil},
		{"int16", int16(42), "42", nil},
		{"int16 negative", int16(-42), "-42", nil},
		{"uint16", uint16(42), "42", nil},
		{"int32", int32(42), "42", nil},
		{"int32 negative", int32(-42), "-42", nil},
		{"uint32", uint32(42), "42", nil},
		{"int64", int64(42), "42", nil},
		{"int64 negative", int64(-42), "-42", nil},
		{"uint64", uint64(42), "42", nil},
		{"float64", float64(42), "42", nil},
		{"float64 negative", float64(-42), "-42", nil},
		{"float64 zero", float64(0), "0", nil},
		{"float64", float64(42.6), "42.6", nil},
		{"float32", float32(42), "42", nil},
		{"float32 negative", float32(-42), "-42", nil},
		{"float32", float32(42.6), "42.6", nil},
		{"string slice", []string{"a", "b"}, "a,b", nil},
		{"nil", nil, "", nil},
		{"slice", []int{42}, "", cfgconv.TypeError{}},
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			v, err := cfgconv.String(p.in)
			assert.IsType(t, p.err, err)
			assert.Equal(t, p.v, v)
		}
		t.Run(p.name, f)
	}
}

func TestStringSlice(t *testing.T) {
	slice := []interface{}{[]int{1, 2, 3}}
	intSlice := []int{1, 2, -3}
	stringSlice := []string{"one", "two"}
	uintSlice := []int{1, 2, 3}
	patterns := []struct {
		name string
		in   interface{}
		v    []string
		err  error
	}{
		{"slice", slice, []string{""}, cfgconv.TypeError{}},
		{"intSlice", intSlice, []string{"1", "2", "-3"}, nil},
		{"stringSlice", stringSlice, []string{"one", "two"}, nil},
		{"uintSlice", uintSlice, []string{"1", "2", "3"}, nil},
		{"string int", "42", []string{"42"}, nil},
		{"bool true", true, nil, cfgconv.TypeError{}},
		{"bool false", false, nil, cfgconv.TypeError{}},
		{"int", int(42), nil, cfgconv.TypeError{}},
		{"int negative", int(-42), nil, cfgconv.TypeError{}},
		{"uint", uint(42), nil, cfgconv.TypeError{}},
		{"int8", int8(42), nil, cfgconv.TypeError{}},
		{"int8 negative", int8(-42), nil, cfgconv.TypeError{}},
		{"uint8", uint8(42), nil, cfgconv.TypeError{}},
		{"int16", int16(42), nil, cfgconv.TypeError{}},
		{"int16 negative", int16(-42), nil, cfgconv.TypeError{}},
		{"uint16", uint16(42), nil, cfgconv.TypeError{}},
		{"int32", int32(42), nil, cfgconv.TypeError{}},
		{"int32 negative", int32(-42), nil, cfgconv.TypeError{}},
		{"uint32", uint32(42), nil, cfgconv.TypeError{}},
		{"int64", int64(42), nil, cfgconv.TypeError{}},
		{"int64 negative", int64(-42), nil, cfgconv.TypeError{}},
		{"uint64", uint64(42), nil, cfgconv.TypeError{}},
		{"float64", float64(42), nil, cfgconv.TypeError{}},
		{"float64 zero", float64(0), nil, cfgconv.TypeError{}},
		{"float64 negative", float64(-42), nil, cfgconv.TypeError{}},
		{"float64 truncate", float64(42.6), nil, cfgconv.TypeError{}},
		{"float64 truncate negative", float64(-42.6), nil, cfgconv.TypeError{}},
		{"float32", float32(42), nil, cfgconv.TypeError{}},
		{"float32 negative", float32(-42), nil, cfgconv.TypeError{}},
		{"float32 truncate", float32(42.6), nil, cfgconv.TypeError{}},
		{"float32 truncate negative", float32(-42.6), nil, cfgconv.TypeError{}},
		{"empty string", "", nil, cfgconv.TypeError{}},
		{"nil", nil, nil, cfgconv.TypeError{}},
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			v, err := cfgconv.StringSlice(p.in)
			assert.IsType(t, p.err, err)
			assert.Equal(t, p.v, v)
		}
		t.Run(p.name, f)
	}
}

func TestStruct(t *testing.T) {
	type testStruct struct {
		A int
		B string
	}
	patterns := []struct {
		name string
		in   interface{}
		v    interface{}
		x    interface{}
		err  error
	}{
		{"bad type", "blah", &testStruct{}, &testStruct{}, cfgconv.TypeError{}},
		{"map",
			map[string]interface{}{
				"a": 1,
				"b": "hello",
			},
			&testStruct{},
			&testStruct{A: 1, B: "hello"},
			nil},
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			v := p.v
			err := cfgconv.Struct(p.in, v)
			assert.IsType(t, p.err, err)
			assert.Equal(t, p.x, p.v)
		}
		t.Run(p.name, f)
	}
}

func TestUnmarshalStructFromMap(t *testing.T) {
	type innerConfig struct {
		A       int
		Btagged string `config:"b"`
		C       []int
		E       string
	}
	type testStruct struct {
		A       int
		Btagged string `config:"b_good"`
		C       []int
		p       int // non-exported fields can't be set
		Nested  innerConfig
	}

	patterns := []struct {
		name string
		in   map[string]interface{}
		v    interface{}
		x    interface{}
		err  error
	}{
		{"bad type", map[string]interface{}{"a": 1}, 3, 3, cfgconv.ErrInvalidStruct},
		{"scalars",
			map[string]interface{}{
				"a":      1,
				"b_good": "hello",
				"d":      "ignored",
				"p":      "non-exported fields can't be set",
			},
			&testStruct{},
			&testStruct{A: 1, Btagged: "hello"},
			nil},
		{"arrays",
			map[string]interface{}{
				"c": []int{1, 2, 3, 4},
				"d": "ignored",
			},
			&testStruct{},
			&testStruct{C: []int{1, 2, 3, 4}},
			nil},
		{"maltyped array",
			map[string]interface{}{
				"a": 42,
				"c": 43,
				"d": "ignored",
			},
			&testStruct{},
			&testStruct{A: 42},
			cfgconv.TypeError{}},
		{"maltyped int",
			map[string]interface{}{
				"a":      "bogus",
				"b_good": "banana",
				"d":      "ignored",
			},
			&testStruct{},
			&testStruct{Btagged: "banana"},
			&strconv.NumError{}},
		{"nested",
			map[string]interface{}{
				"b_good": "foo.b",
				"nested": map[string]interface{}{
					"a": 43,
					"b": "foo.nested.b",
					"c": []int{5, 6, 7, 8}},
			},
			&testStruct{},
			&testStruct{
				Btagged: "foo.b",
				Nested: innerConfig{
					A:       43,
					Btagged: "foo.nested.b",
					C:       []int{5, 6, 7, 8}}},
			nil},
		{"nested wrong type",
			map[string]interface{}{
				"nested": map[string]interface{}{
					"a": []int{6, 7}}},
			&testStruct{},
			&testStruct{},
			cfgconv.TypeError{}},
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			err := cfgconv.UnmarshalStructFromMap(p.in, p.v)
			assert.IsType(t, p.err, err)
			assert.Equal(t, p.x, p.v)
		}
		t.Run(p.name, f)
	}
}

func TestTime(t *testing.T) {
	patterns := []struct {
		name string
		in   interface{}
		v    time.Time
		err  error
	}{
		{"bad format bytes", []byte("glob"), time.Time{}, &time.ParseError{}},
		{"bad format string", "foo", time.Time{}, &time.ParseError{}},
		{"bad type int", 34, time.Time{}, cfgconv.TypeError{}},
		{"date string", "2017-03-01", time.Time{}, &time.ParseError{}},
		{"empty bytes", []byte{}, time.Time{}, &time.ParseError{}},
		{"empty string", "", time.Time{}, &time.ParseError{}},
		{"full datetime", "2017-03-01T01:02:03Z", time.Date(2017, 3, 1, 1, 2, 3, 0, time.UTC), nil},
		{"full datetime", []byte("2017-03-01T01:02:03Z"), time.Date(2017, 3, 1, 1, 2, 3, 0, time.UTC), nil},
		{"nil", nil, time.Time{}, cfgconv.TypeError{}},
		{"year string", "2017", time.Time{}, &time.ParseError{}},
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			v, err := cfgconv.Time(p.in)
			assert.IsType(t, p.err, err)
			assert.Equal(t, p.v, v)
		}
		t.Run(p.name, f)
	}
}

func TestUint(t *testing.T) {
	patterns := []struct {
		name string
		in   interface{}
		v    uint64
		err  error
	}{
		{"bool false", false, 0, nil},
		{"bool true", true, 1, nil},
		{"empty string", "", 0, &strconv.NumError{}},
		{"float32 negative", float32(-42), 0, cfgconv.TypeError{}},
		{"float32 truncate negative", float32(-42.6), 0, cfgconv.TypeError{}},
		{"float32 truncate", float32(42.6), 42, nil},
		{"float32", float32(42), 42, nil},
		{"float64 negative", float64(-42), 0, cfgconv.TypeError{}},
		{"float64 truncate negative", float64(-42.6), 0, cfgconv.TypeError{}},
		{"float64 truncate", float64(42.6), 42, nil},
		{"float64 zero", float64(0), 0, nil},
		{"float64", float64(42), 42, nil},
		{"int negative", int(-42), 0, cfgconv.TypeError{}},
		{"int", int(42), 42, nil},
		{"int16 negative", int16(-42), 0, cfgconv.TypeError{}},
		{"int16", int16(42), 42, nil},
		{"int32 negative", int32(-42), 0, cfgconv.TypeError{}},
		{"int32", int32(42), 42, nil},
		{"int64 negative", int64(-42), 0, cfgconv.TypeError{}},
		{"int64", int64(42), 42, nil},
		{"int8 negative", int8(-42), 0, cfgconv.TypeError{}},
		{"int8", int8(42), 42, nil},
		{"nil", nil, 0, nil},
		{"slice", []int{42}, 0, cfgconv.TypeError{}},
		{"string float", "42.5", 0, &strconv.NumError{}},
		{"string int negative", "-42", 0, &strconv.NumError{}},
		{"string int", "42", 42, nil},
		{"string junk", "junk", 0, &strconv.NumError{}},
		{"uint", uint(42), 42, nil},
		{"uint16", uint16(42), 42, nil},
		{"uint32", uint32(42), 42, nil},
		{"uint64", uint64(42), 42, nil},
		{"uint8", uint8(42), 42, nil},
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			v, err := cfgconv.Uint(p.in)
			assert.IsType(t, p.err, err)
			assert.Equal(t, p.v, v)
		}
		t.Run(p.name, f)
	}
}

func TestUintSlice(t *testing.T) {
	slice := []interface{}{[]int{1, 2, 3}}
	intSlice := []int{1, 2, -3}
	stringSlice := []string{"one", "two"}
	uintSlice := []int{1, 2, 3}
	patterns := []struct {
		name string
		in   interface{}
		v    []uint64
		err  error
	}{
		{"slice", slice, []uint64{0}, cfgconv.TypeError{}},
		{"intSlice", intSlice, []uint64{1, 2, 0}, cfgconv.TypeError{}},
		{"stringSlice", stringSlice, []uint64{0, 0}, &strconv.NumError{}},
		{"uintSlice", uintSlice, []uint64{1, 2, 3}, nil},
		{"string int", "42", []uint64{42}, nil},
		{"bool true", true, nil, cfgconv.TypeError{}},
		{"bool false", false, nil, cfgconv.TypeError{}},
		{"int", int(42), nil, cfgconv.TypeError{}},
		{"int negative", int(-42), nil, cfgconv.TypeError{}},
		{"uint", uint(42), nil, cfgconv.TypeError{}},
		{"int8", int8(42), nil, cfgconv.TypeError{}},
		{"int8 negative", int8(-42), nil, cfgconv.TypeError{}},
		{"uint8", uint8(42), nil, cfgconv.TypeError{}},
		{"int16", int16(42), nil, cfgconv.TypeError{}},
		{"int16 negative", int16(-42), nil, cfgconv.TypeError{}},
		{"uint16", uint16(42), nil, cfgconv.TypeError{}},
		{"int32", int32(42), nil, cfgconv.TypeError{}},
		{"int32 negative", int32(-42), nil, cfgconv.TypeError{}},
		{"uint32", uint32(42), nil, cfgconv.TypeError{}},
		{"int64", int64(42), nil, cfgconv.TypeError{}},
		{"int64 negative", int64(-42), nil, cfgconv.TypeError{}},
		{"uint64", uint64(42), nil, cfgconv.TypeError{}},
		{"float64", float64(42), nil, cfgconv.TypeError{}},
		{"float64 zero", float64(0), nil, cfgconv.TypeError{}},
		{"float64 negative", float64(-42), nil, cfgconv.TypeError{}},
		{"float64 truncate", float64(42.6), nil, cfgconv.TypeError{}},
		{"float64 truncate negative", float64(-42.6), nil, cfgconv.TypeError{}},
		{"float32", float32(42), nil, cfgconv.TypeError{}},
		{"float32 negative", float32(-42), nil, cfgconv.TypeError{}},
		{"float32 truncate", float32(42.6), nil, cfgconv.TypeError{}},
		{"float32 truncate negative", float32(-42.6), nil, cfgconv.TypeError{}},
		{"empty string", "", nil, cfgconv.TypeError{}},
		{"nil", nil, nil, cfgconv.TypeError{}},
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			v, err := cfgconv.UintSlice(p.in)
			assert.IsType(t, p.err, err)
			assert.Equal(t, p.v, v)
		}
		t.Run(p.name, f)
	}
}

func TestTypeError(t *testing.T) {
	patterns := []byte{0x00, 0xa0, 0x0a, 0x9a, 0xa9, 0xff}
	for _, p := range patterns {
		f := func(t *testing.T) {
			e := cfgconv.TypeError{Value: p}
			expected := fmt.Sprintf("cfgconv: cannot convert '%#v'(%T) to %s", e.Value, e.Value, e.Kind)
			assert.Equal(t, expected, e.Error())
		}
		t.Run(fmt.Sprintf("%x", p), f)
	}
}

func TestOverflowError(t *testing.T) {
	patterns := []byte{0x00, 0xa0, 0x0a, 0x9a, 0xa9, 0xff}
	for _, p := range patterns {
		f := func(t *testing.T) {
			e := cfgconv.OverflowError{Value: p}
			expected := fmt.Sprintf("cfgconv: overflow converting '%v' to %s", e.Value, e.Kind)
			assert.Equal(t, expected, e.Error())
		}
		t.Run(fmt.Sprintf("%x", p), f)
	}
}
