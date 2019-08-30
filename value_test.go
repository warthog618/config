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

	"github.com/stretchr/testify/assert"
	"github.com/warthog618/config"
	"github.com/warthog618/config/cfgconv"
)

func TestNewValue(t *testing.T) {
	v := config.NewValue(1)
	assert.Equal(t, int64(1), v.Int())
	assert.NotPanics(t, func() {
		v.Time()
	})
	v = config.NewValue(2, config.WithMust())
	assert.Equal(t, int64(2), v.Int())
	assert.Panics(t, func() {
		v.Time()
	})
}

func TestBool(t *testing.T) {
	mr := mockGetter{
		"bool":       true,
		"boolString": "true",
		"boolInt":    1,
		"notabool":   "bogus",
	}
	patterns := []struct {
		k   string
		v   bool
		err error
	}{
		{"bool", true, nil},
		{"boolString", true, nil},
		{"boolInt", true, nil},
		{"notabool", false, &strconv.NumError{}},
	}
	c := config.New(&mr)
	for _, p := range patterns {
		f := func(t *testing.T) {
			var eherr error
			val, err := c.Get(p.k, config.WithErrorHandler(
				config.ErrorHandler(func(e error) error {
					eherr = e
					return nil
				})))
			assert.Nil(t, err)
			v := val.Bool()
			assert.IsType(t, p.err, eherr)
			assert.Equal(t, p.v, v)
		}
		t.Run(p.k, f)
	}
}

func TestDuration(t *testing.T) {
	mr := mockGetter{
		"duration":     "123ms",
		"notaduration": "bogus",
	}
	patterns := []struct {
		k   string
		v   time.Duration
		err error
	}{
		{"duration", time.Duration(123000000), nil},
		{"notaduration", time.Duration(0), errors.New("")},
	}
	c := config.New(&mr)
	for _, p := range patterns {
		f := func(t *testing.T) {
			var eherr error
			val, err := c.Get(p.k, config.WithErrorHandler(
				config.ErrorHandler(func(e error) error {
					eherr = e
					return nil
				})))
			assert.Nil(t, err)
			v := val.Duration()
			assert.IsType(t, p.err, eherr)
			assert.Equal(t, p.v, v)
		}
		t.Run(p.k, f)
	}
}

func TestFloat(t *testing.T) {
	mr := mockGetter{
		"float":       3.1415,
		"floatString": "3.1415",
		"floatInt":    1,
		"notafloat":   "bogus",
	}
	patterns := []struct {
		k   string
		v   float64
		err error
	}{
		{"float", 3.1415, nil},
		{"floatString", 3.1415, nil},
		{"floatInt", 1, nil},
		{"notafloat", 0, &strconv.NumError{}},
	}
	c := config.New(&mr)
	for _, p := range patterns {
		f := func(t *testing.T) {
			var eherr error
			val, err := c.Get(p.k, config.WithErrorHandler(
				config.ErrorHandler(func(e error) error {
					eherr = e
					return nil
				})))
			assert.Nil(t, err)
			v := val.Float()
			assert.IsType(t, p.err, eherr)
			assert.Equal(t, p.v, v)
		}
		t.Run(p.k, f)
	}
}

func TestInt(t *testing.T) {
	mr := mockGetter{
		"int":       42,
		"intString": "43",
		"notaint":   "bogus",
	}
	patterns := []struct {
		k   string
		v   int64
		err error
	}{
		{"int", 42, nil},
		{"intString", 43, nil},
		{"notaint", 0, &strconv.NumError{}},
	}
	c := config.New(&mr)
	for _, p := range patterns {
		f := func(t *testing.T) {
			var eherr error
			val, err := c.Get(p.k, config.WithErrorHandler(
				config.ErrorHandler(func(e error) error {
					eherr = e
					return nil
				})))
			assert.Nil(t, err)
			v := val.Int()
			assert.IsType(t, p.err, eherr)
			assert.Equal(t, p.v, v)
		}
		t.Run(p.k, f)
	}
}

func TestString(t *testing.T) {
	mr := mockGetter{
		"string":     "a string",
		"stringInt":  42,
		"notastring": struct{}{},
	}
	patterns := []struct {
		k   string
		v   string
		err error
	}{
		{"string", "a string", nil},
		{"stringInt", "42", nil},
		{"notastring", "", cfgconv.TypeError{}},
	}
	c := config.New(&mr)
	for _, p := range patterns {
		f := func(t *testing.T) {
			var eherr error
			val, err := c.Get(p.k, config.WithErrorHandler(
				config.ErrorHandler(func(e error) error {
					eherr = e
					return nil
				})))
			assert.Nil(t, err)
			v := val.String()
			assert.IsType(t, p.err, eherr)
			assert.Equal(t, p.v, v)
		}
		t.Run(p.k, f)
	}
}

func TestTime(t *testing.T) {
	mr := mockGetter{
		"time":     "2017-03-01T01:02:03Z",
		"notatime": "bogus",
	}
	patterns := []struct {
		k   string
		v   time.Time
		err error
	}{
		{"time", time.Date(2017, 3, 1, 1, 2, 3, 0, time.UTC), nil},
		{"notatime", time.Time{}, &time.ParseError{}},
	}
	c := config.New(&mr)
	for _, p := range patterns {
		f := func(t *testing.T) {
			var eherr error
			val, err := c.Get(p.k, config.WithErrorHandler(
				config.ErrorHandler(func(e error) error {
					eherr = e
					return nil
				})))
			v := val.Time()
			assert.IsType(t, p.err, eherr)
			assert.Nil(t, err)
			assert.Equal(t, p.v, v)
		}
		t.Run(p.k, f)
	}
}

func TestUint(t *testing.T) {
	mr := mockGetter{
		"uint":       42,
		"uintString": "43",
		"notaUint":   "bogus",
	}
	patterns := []struct {
		k   string
		v   uint64
		err error
	}{
		{"uint", 42, nil},
		{"uintString", 43, nil},
		{"notaUint", 0, &strconv.NumError{}},
	}
	c := config.New(&mr)
	for _, p := range patterns {
		f := func(t *testing.T) {
			var eherr error
			val, err := c.Get(p.k, config.WithErrorHandler(
				config.ErrorHandler(func(e error) error {
					eherr = e
					return nil
				})))
			assert.Nil(t, err)
			v := val.Uint()
			assert.IsType(t, p.err, eherr)
			assert.Equal(t, p.v, v)
		}
		t.Run(p.k, f)
	}
}

func TestSlice(t *testing.T) {
	mr := mockGetter{
		"slice": []interface{}{1, 2, 3, 4},
		"animals": []interface{}{
			map[string]interface{}{"Name": "Platypus", "Order": "Monotremata"},
			map[string]interface{}{"Order": "Dasyuromorphia", "Name": "Quoll"},
		},
		"casttoslice": "bogus",
		"notaslice":   struct{}{},
	}
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
	}
	c := config.New(&mr)
	for _, p := range patterns {
		f := func(t *testing.T) {
			var eherr error
			val, err := c.Get(p.k, config.WithErrorHandler(
				config.ErrorHandler(func(e error) error {
					eherr = e
					return nil
				})))
			assert.Nil(t, err)
			v := val.Slice()
			assert.IsType(t, p.err, eherr)
			assert.Equal(t, p.v, v)
		}
		t.Run(p.k, f)
	}
}

func TestIntSlice(t *testing.T) {
	mr := mockGetter{
		"slice":       []int64{1, 2, -3, 4},
		"casttoslice": "42",
		"stringslice": []string{"one", "two", "three"},
		"notaslice":   "bogus",
	}
	patterns := []struct {
		k   string
		v   []int64
		err error
	}{
		{"slice", []int64{1, 2, -3, 4}, nil},
		{"casttoslice", []int64{42}, nil},
		{"stringslice", []int64{0, 0, 0}, &strconv.NumError{}},
		{"notaslice", []int64{0}, &strconv.NumError{}},
	}
	c := config.New(&mr)
	for _, p := range patterns {
		f := func(t *testing.T) {
			var eherr error
			val, err := c.Get(p.k, config.WithErrorHandler(
				config.ErrorHandler(func(e error) error {
					eherr = e
					return nil
				})))
			assert.Nil(t, err)
			v := val.IntSlice()
			assert.IsType(t, p.err, eherr)
			assert.Equal(t, p.v, v)
		}
		t.Run(p.k, f)
	}
}

func TestStringSlice(t *testing.T) {
	mr := mockGetter{
		"intslice":        []int64{1, 2, -3, 4},
		"stringslice":     []string{"one", "two", "three"},
		"uintslice":       []uint64{1, 2, 3, 4},
		"notastringslice": []interface{}{1, 2, struct{}{}},
		"casttoslice":     "bogus",
		"notaslice":       struct{}{},
	}
	patterns := []struct {
		k   string
		v   []string
		err error
	}{
		{"intslice", []string{"1", "2", "-3", "4"}, nil},
		{"uintslice", []string{"1", "2", "3", "4"}, nil},
		{"stringslice", []string{"one", "two", "three"}, nil},
		{"casttoslice", []string{"bogus"}, nil},
		{"notastringslice", []string{"1", "2", ""}, cfgconv.TypeError{}},
		{"notaslice", nil, cfgconv.TypeError{}},
	}
	c := config.New(&mr)
	for _, p := range patterns {
		f := func(t *testing.T) {
			var eherr error
			val, err := c.Get(p.k, config.WithErrorHandler(
				config.ErrorHandler(func(e error) error {
					eherr = e
					return nil
				})))
			assert.Nil(t, err)
			v := val.StringSlice()
			assert.IsType(t, p.err, eherr)
			assert.Equal(t, p.v, v)
		}
		t.Run(p.k, f)
	}
}

func TestUintSlice(t *testing.T) {
	mr := mockGetter{
		"slice":       []uint64{1, 2, 3, 4},
		"casttoslice": "42",
		"intslice":    []int64{1, 2, -3, 4},
		"stringslice": []string{"one", "two", "three"},
		"notaslice":   "bogus",
	}
	patterns := []struct {
		k   string
		v   []uint64
		err error
	}{
		{"slice", []uint64{1, 2, 3, 4}, nil},
		{"casttoslice", []uint64{42}, nil},
		{"intslice", []uint64{1, 2, 0, 4}, cfgconv.TypeError{}},
		{"stringslice", []uint64{0, 0, 0}, &strconv.NumError{}},
		{"notaslice", []uint64{0}, &strconv.NumError{}},
	}
	c := config.New(&mr)
	for _, p := range patterns {
		f := func(t *testing.T) {
			var eherr error
			val, err := c.Get(p.k, config.WithErrorHandler(
				config.ErrorHandler(func(e error) error {
					eherr = e
					return nil
				})))
			assert.Nil(t, err)
			v := val.UintSlice()
			assert.IsType(t, p.err, eherr)
			assert.Equal(t, p.v, v)
		}
		t.Run(p.k, f)
	}
}
