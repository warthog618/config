// Performs type conversions from incoming configuration types to expected
// internal type.
// The type conversions are flexible, and include automatic conversion
// from string to numeric,
// from string to bool,
// from numeric to string,
// from numeric to bool,
// from bool to numeric
// from bool to string
// from float to int
//
// Performs range checks when converting between types to prevent loss of precision.
//
package cfgconv

import (
	"fmt"
	"reflect"
	"strconv"
)

const maxUint = ^uint64(0)
const minUint = 0
const maxInt = int64(maxUint >> 1)

func int2bool(val int) bool {
	if val == 0 {
		return false
	}
	return true
}

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

// Converts the value v to the requested type rt, if possible.
// If not possible then returns a zeroed instance and an error.
// Returned errors are typically TypeErrors or OverflowErrors,
// but can also be errors from underlying type converters.
func Convert(v interface{}, rt reflect.Type) (interface{}, error) {
	rv := reflect.Indirect(reflect.New(rt))
	ri := rv.Interface()
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		cv, err := Int(v)
		if err != nil {
			return ri, err
		}
		if rv.OverflowInt(cv) {
			return ri, OverflowError{Value: v, Kind: rv.Kind()}
		} else {
			rv.SetInt(cv)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		cv, err := Uint(v)
		if err != nil {
			return ri, err
		}
		if rv.OverflowUint(cv) {
			return ri, OverflowError{Value: v, Kind: rv.Kind()}
		} else {
			rv.SetUint(cv)
		}
	case reflect.Float32, reflect.Float64:
		cv, err := Float(v)
		if err != nil {
			return ri, err
		}
		if rv.OverflowFloat(cv) {
			return ri, OverflowError{Value: v, Kind: rv.Kind()}
		} else {
			rv.SetFloat(cv)
		}
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
	case reflect.Slice:
		vv := reflect.ValueOf(v)
		if vv.Kind() != reflect.Slice {
			return ri, TypeError{Value: v, Kind: reflect.Slice}
		}
		et := rt.Elem()
		rv = reflect.MakeSlice(rv.Type(), vv.Len(), vv.Len())
		for idx := 0; idx < vv.Len(); idx++ {
			sv, err := Convert(vv.Index(idx).Interface(), et)
			if err != nil {
				rv = reflect.Indirect(reflect.New(rt))
				return rv.Interface(), err
			}
			rv.Index(idx).Set(reflect.ValueOf(sv))
		}
	}
	return rv.Interface(), nil
}

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

func Slice(v interface{}) ([]interface{}, error) {
	if slice, ok := v.([]interface{}); ok {
		return slice, nil
	}
	vv := reflect.ValueOf(v)
	if vv.Kind() == reflect.Slice {
		slice := make([]interface{}, vv.Len(), vv.Len())
		for idx := 0; idx < vv.Len(); idx++ {
			slice[idx] = vv.Index(idx).Interface()
		}
		return slice, nil
	}
	return []interface{}{}, TypeError{Value: v, Kind: reflect.Slice}
}

func String(v interface{}) (string, error) {
	switch vt := v.(type) {
	case string:
		return vt, nil
	case []byte:
		return string(vt), nil
	case int, uint, int8, uint8, int16, uint16, int32, uint32, int64, uint64, float32, float64, bool:
		return fmt.Sprintf("%v", v), nil
	case nil:
		return "", nil
	}
	return "", TypeError{Value: v, Kind: reflect.String}
}

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

type TypeError struct {
	Value interface{}
	Kind  reflect.Kind
}

func (e TypeError) Error() string {
	return fmt.Sprintf("cfgconv: cannot convert '%#v'(%T) to %s", e.Value, e.Value, e.Kind)
}

type OverflowError struct {
	Value interface{}
	Kind  reflect.Kind
}

func (e OverflowError) Error() string {
	return fmt.Sprintf("cfgconv: overflow converting '%v' to %s", e.Value, e.Kind)
}
