// Copyright © 2017 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package cfgconv provides type conversions from incoming configuration types
// to requested internal types.  Essentially a more general version of strconv.
//
// The type conversions are flexible, and include automatic conversion from:
//
//	string to numeric
// 	string to bool
// 	numeric to string
// 	numeric to bool
// 	bool to numeric
// 	bool to string
// 	float to int
//
// Performs range checks when converting between types to prevent loss of precision.
//
package cfgconv

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

const maxUint = ^uint64(0)
const minUint = 0
const maxInt = int64(maxUint >> 1)

func int2bool(v int) bool {
	if v == 0 {
		return false
	}
	return true
}

// Bool converts a generic object into a bool, if possible.
// Returns false and an error if conversion is not possible.
func Bool(v interface{}) (bool, error) {
	switch vt := v.(type) {
	case bool:
		return vt, nil
	case int:
		return int2bool(int(vt)), nil
	case uint:
		return int2bool(int(vt)), nil
	case string:
		return strconv.ParseBool(vt)
	case int8:
		return int2bool(int(vt)), nil
	case uint8:
		return int2bool(int(vt)), nil
	case int16:
		return int2bool(int(vt)), nil
	case uint16:
		return int2bool(int(vt)), nil
	case int32:
		return int2bool(int(vt)), nil
	case uint32:
		return int2bool(int(vt)), nil
	case int64:
		return int2bool(int(vt)), nil
	case uint64:
		return int2bool(int(vt)), nil
	case nil:
		return false, nil
	}
	return false, TypeError{Value: v, Kind: reflect.Bool}
}

// Convert converts the value v to the requested type rt, if possible.
// If not possible then returns a zeroed instance and an error.
// Returned errors are typically TypeErrors or OverflowErrors,
// but can also be errors from underlying type converters.
func Convert(v interface{}, rt reflect.Type) (interface{}, error) {
	if rt == nil {
		return v, nil
	}
	rv := reflect.Indirect(reflect.New(rt))
	ri := rv.Interface()
	// First handle specific types.
	switch rt {
	case reflect.TypeOf(time.Duration(0)):
		cv, err := Duration(v)
		if err != nil {
			return ri, err
		}
		rv.SetInt(int64(cv))
		return rv.Interface(), nil
	case reflect.TypeOf(time.Time{}):
		cv, err := Time(v)
		if err != nil {
			return ri, err
		}
		rv.Set(reflect.ValueOf(cv))
		return rv.Interface(), nil
	}
	// Then generic types.
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		cv, err := Int(v)
		if err != nil {
			return ri, err
		}
		if rv.OverflowInt(cv) {
			return ri, OverflowError{Value: v, Kind: rv.Kind()}
		}
		rv.SetInt(cv)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		cv, err := Uint(v)
		if err != nil {
			return ri, err
		}
		if rv.OverflowUint(cv) {
			return ri, OverflowError{Value: v, Kind: rv.Kind()}
		}
		rv.SetUint(cv)
	case reflect.Float32, reflect.Float64:
		cv, err := Float(v)
		if err != nil {
			return ri, err
		}
		if rv.OverflowFloat(cv) {
			return ri, OverflowError{Value: v, Kind: rv.Kind()}
		}
		rv.SetFloat(cv)
	case reflect.String:
		cv, err := String(v)
		if err != nil {
			return ri, err
		}
		rv.SetString(cv)
	case reflect.Bool:
		cv, err := Bool(v)
		if err != nil {
			return ri, err
		}
		rv.SetBool(cv)
	case reflect.Struct:
		err := Struct(v, rv.Addr().Interface())
		return rv.Interface(), err
	case reflect.Slice:
		et := rt.Elem()
		vv := reflect.ValueOf(v)
		switch vv.Kind() {
		case reflect.Slice:
			rv = reflect.MakeSlice(rv.Type(), vv.Len(), vv.Len())
			for idx := 0; idx < vv.Len(); idx++ {
				sv, err := Convert(vv.Index(idx).Interface(), et)
				if err != nil {
					rv = reflect.Indirect(reflect.New(rt))
					return rv.Interface(), err
				}
				rv.Index(idx).Set(reflect.ValueOf(sv))
			}
		case reflect.String:
			sv, err := Convert(vv.Interface(), et)
			if err != nil {
				return rv.Interface(), err
			}
			rv = reflect.MakeSlice(rv.Type(), 1, 1)
			rv.Index(0).Set(reflect.ValueOf(sv))
		default:
			return ri, TypeError{Value: v, Kind: reflect.Slice}
		}
	case reflect.Interface:
		return v, nil
	}
	return rv.Interface(), nil
}

// Duration converts a string to a duration, if possible.
// Returns 0 and an error if conversion is not possible.
func Duration(v interface{}) (time.Duration, error) {
	switch vt := v.(type) {
	case string:
		cv, err := time.ParseDuration(vt)
		if err == nil {
			return cv, nil
		}
		return time.Duration(0), err
	case []byte:
		cv, err := time.ParseDuration(string(vt))
		if err == nil {
			return cv, nil
		}
		return time.Duration(0), err
	}
	return time.Duration(0), TypeError{Value: v, Kind: reflect.Int64}
}

// Float converts a generic object into a float64, if possible.
// Returns 0 and an error if conversion is not possible.
func Float(v interface{}) (float64, error) {
	switch vt := v.(type) {
	case float64:
		return vt, nil
	case float32:
		return float64(vt), nil
	case string:
		return strconv.ParseFloat(vt, 64)
	case bool:
		if vt {
			return 1, nil
		}
		return 0, nil
	case int:
		return float64(vt), nil
	case uint:
		return float64(vt), nil
	case int64:
		return float64(vt), nil
	case uint64:
		return float64(vt), nil
	case int32:
		return float64(vt), nil
	case uint32:
		return float64(vt), nil
	case int8:
		return float64(vt), nil
	case uint8:
		return float64(vt), nil
	case int16:
		return float64(vt), nil
	case uint16:
		return float64(vt), nil
	case nil:
		return 0, nil
	}
	return 0, TypeError{Value: v, Kind: reflect.Float64}
}

// Int converts a generic object into an int64, if possible.
// Returns 0 and an error if conversion is not possible.
func Int(v interface{}) (int64, error) {
	switch vt := v.(type) {
	case bool:
		if vt {
			return 1, nil
		}
		return 0, nil
	case int:
		return int64(vt), nil
	case uint:
		return int64(vt), nil
	case string:
		return strconv.ParseInt(vt, 10, 64)
	case float64:
		return int64(vt), nil
	case float32:
		return int64(vt), nil
	case int64:
		return vt, nil
	case uint64:
		if vt <= uint64(maxInt) {
			return int64(vt), nil
		}
	case int32:
		return int64(vt), nil
	case uint32:
		return int64(vt), nil
	case int8:
		return int64(vt), nil
	case uint8:
		return int64(vt), nil
	case int16:
		return int64(vt), nil
	case uint16:
		return int64(vt), nil
	case nil:
		return 0, nil
	}
	return 0, TypeError{Value: v, Kind: reflect.Int}
}

// Slice converts a slice of something into a []interface{}
//
// Also interprets strings as a single element slice,
// to allow for the case where a Getter cannot distinguish
// between a single entry slice and a literal,
// e.g. the env Getter.
func Slice(v interface{}) ([]interface{}, error) {
	if slice, ok := v.([]interface{}); ok {
		return slice, nil
	}
	vv := reflect.ValueOf(v)
	switch vv.Kind() {
	case reflect.Slice:
		slice := make([]interface{}, vv.Len(), vv.Len())
		for idx := 0; idx < vv.Len(); idx++ {
			slice[idx] = vv.Index(idx).Interface()
		}
		return slice, nil
	case reflect.String:
		if len(v.(string)) != 0 {
			slice := make([]interface{}, 1, 1)
			slice[0] = v
			return slice, nil
		}
	}
	return nil, TypeError{Value: v, Kind: reflect.Slice}
}

// String converts a generic object into a string, if possible.
// Returns 0 and an error if conversion is not possible.
func String(v interface{}) (string, error) {
	switch vt := v.(type) {
	case string:
		return vt, nil
	case []byte:
		return string(vt), nil
	case int, uint, int8, uint8, int16, uint16, int32, uint32, int64, uint64, float32, float64, bool:
		return fmt.Sprintf("%v", v), nil
	case []string:
		// this case undoes accidental conversion of a string into a slice
		// by a Getter as the string contains a list separator character.
		// Of course for the env Getter the separator is ":", so in its case
		// the resulting string will be wrong - with ":" replaced with ",".
		// But it does fix the Getters that use ",", such as flag and properties.
		return strings.Join(vt, ","), nil
	case nil:
		return "", nil
	}
	return "", TypeError{Value: v, Kind: reflect.String}
}

// Struct converts a generic object to a struct.
//
// The obj is a struct with fields corresponding to config values.
// The config values will be converted to the type defined in the corresponding
// struct fields.  Overflow checks are performed during conversion to ensure the
// value returned by the getter can fit within the designated field.
//
// By default the map keys are drawn from the struct field names,
// converted to LowerCamelCase (as per typical JSON naming conventions).
// This can be overridden using `config:"<name>"` tags.
//
// Struct fields which do not have corresponding map keys are ignored,
// as are map keys which have no corresponding struct field,
// and non-exported struct fields.
//
// The error identifies the first type conversion error, if any.
//
// Currently only support conversion from map[string]interface{},
// but may support struct to struct conversions at a later date,
// hence the wrapper around UnmarshalStructFromMap.
func Struct(v interface{}, obj interface{}) error {
	if vm, ok := v.(map[string]interface{}); ok {
		err := UnmarshalStructFromMap(vm, obj)
		return err
	}
	return TypeError{Value: v, Kind: reflect.Struct}
}

// Time converts a string to a duration, if possible.
// Returns 0 and an error if conversion is not possible.
func Time(v interface{}) (time.Time, error) {
	switch vt := v.(type) {
	case string:
		cv, err := time.Parse(time.RFC3339, vt)
		if err == nil {
			return cv, nil
		}
		return time.Time{}, err
	case []byte:
		cv, err := time.Parse(time.RFC3339, string(vt))
		if err == nil {
			return cv, nil
		}
		return time.Time{}, err
	}
	return time.Time{}, TypeError{Value: v, Kind: reflect.Int64}
}

// Uint converts a generic object into a uint64, if possible.
// Returns 0 and an error if conversion is not possible.
func Uint(v interface{}) (uint64, error) {
	switch vt := v.(type) {
	case uint:
		return uint64(vt), nil
	case int:
		if vt >= minUint {
			return uint64(vt), nil
		}
	case uint64:
		return vt, nil
	case int64:
		if vt >= minUint {
			return uint64(vt), nil
		}
	case string:
		return strconv.ParseUint(vt, 10, 64)
	case float64:
		if vt >= 0 {
			return uint64(vt), nil
		}
	case float32:
		if vt >= 0 {
			return uint64(vt), nil
		}
	case bool:
		if vt {
			return 1, nil
		}
		return 0, nil
	case int32:
		if vt >= minUint {
			return uint64(vt), nil
		}
	case uint32:
		return uint64(vt), nil
	case int8:
		if vt >= minUint {
			return uint64(vt), nil
		}
	case uint8:
		return uint64(vt), nil
	case int16:
		if vt >= minUint {
			return uint64(vt), nil
		}
	case uint16:
		return uint64(vt), nil
	case nil:
		return 0, nil
	}
	return 0, TypeError{Value: v, Kind: reflect.Uint}
}

// UnmarshalStructFromMap populates a struct with the values from a map.
//
// The obj is pointer to a struct with fields corresponding to config values.
// The config values will be converted to the type defined in the corresponding
// struct fields.  Overflow checks are performed during conversion to ensure the
// value returned by the getter can fit within the designated field.
//
// By default the map keys are drawn from the struct field names,
// converted to LowerCamelCase (as per typical JSON naming conventions).
// This can be overridden using `config:"<name>"` tags.
//
// Struct fields which do not have corresponding map keys are ignored,
// as are map keys which have no corresponding struct field,
// and non-exported struct fields.
//
// The error identifies the first type conversion error, if any.
func UnmarshalStructFromMap(m map[string]interface{}, obj interface{}) (rerr error) {
	ov := reflect.Indirect(reflect.ValueOf(obj))
	if ov.Kind() != reflect.Struct {
		return ErrInvalidStruct
	}
	for idx := 0; idx < ov.NumField(); idx++ {
		fv := ov.Field(idx)
		if !fv.CanSet() {
			// ignore unexported fields.
			continue
		}
		ft := ov.Type().Field(idx)
		key := ft.Tag.Get("config")
		if len(key) == 0 {
			key = lowerCamelCase(ft.Name)
		}
		v, ok := m[key]
		if !ok {
			continue
		}
		if fv.Kind() == reflect.Struct {
			// nested struct
			if vm, ok := v.(map[string]interface{}); ok {
				err := UnmarshalStructFromMap(vm, fv.Addr().Interface())
				if err != nil && rerr == nil {
					rerr = err
				}
			}
		} else {
			// else assume a leaf
			if cv, err := Convert(v, fv.Type()); err == nil {
				fv.Set(reflect.ValueOf(cv))
			} else if rerr == nil {
				rerr = err
			}
		}
	}
	return rerr
}

// ErrInvalidStruct indicates UnMarshal was provides an object to populate
// which is not a pointer to struct.
var ErrInvalidStruct = errors.New("unmarshal: provided obj is not pointer to struct")

// TypeError indicates a type conversion was not possible.
// Identifies the value being converted and the kind it
// couldn't be converted into.
type TypeError struct {
	Value interface{}
	Kind  reflect.Kind
}

func (e TypeError) Error() string {
	return fmt.Sprintf("cfgconv: cannot convert '%#v'(%T) to %s", e.Value, e.Value, e.Kind)
}

// OverflowError indicates type conversion would lose precision.
// Identifies the value being converted and the kind it
// couldn't be converted into without loss of precision.
type OverflowError struct {
	Value interface{}
	Kind  reflect.Kind
}

func (e OverflowError) Error() string {
	return fmt.Sprintf("cfgconv: overflow converting '%v' to %s", e.Value, e.Kind)
}

// lowerCamelCase converts the first rune of a string to lower case.
// The function assumes key is already camel cased, so only
// lower cases the leading character.
// This is used to convert Go exported field names to config space keys.
// e.g. ConfigFile becomes configFile.
func lowerCamelCase(key string) string {
	r, n := utf8.DecodeRuneInString(key)
	return string(unicode.ToLower(r)) + key[n:]
}
