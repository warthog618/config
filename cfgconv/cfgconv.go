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
	"strconv"
)

const maxUint = ^uint64(0)
const minUint = 0
const maxInt = int64(maxUint >> 1)
const minInt = -maxInt - 1

func int2bool(val int) bool {
	if val == 0 {
		return false
	}
	return true
}

func Bool(val interface{}) (bool, error) {
	switch v := val.(type) {
	case bool:
		return v, nil
	case int:
		return int2bool(int(v)), nil
	case uint:
		return int2bool(int(v)), nil
	case string:
		return strconv.ParseBool(v)
	case int8:
		return int2bool(int(v)), nil
	case uint8:
		return int2bool(int(v)), nil
	case int16:
		return int2bool(int(v)), nil
	case uint16:
		return int2bool(int(v)), nil
	case int32:
		return int2bool(int(v)), nil
	case uint32:
		return int2bool(int(v)), nil
	case int64:
		return int2bool(int(v)), nil
	case uint64:
		return int2bool(int(v)), nil
	case nil:
		return false, nil
	}
	return false, fmt.Errorf("can't convert %#v to boolean", val)
}

func Float(val interface{}) (float64, error) {
	switch v := val.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case string:
		return strconv.ParseFloat(v, 64)
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	case int:
		return float64(v), nil
	case uint:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case uint64:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case uint32:
		return float64(v), nil
	case int8:
		return float64(v), nil
	case uint8:
		return float64(v), nil
	case int16:
		return float64(v), nil
	case uint16:
		return float64(v), nil
	case nil:
		return 0, nil
	}
	return 0, fmt.Errorf("can't convert %#v to float", val)
}

func Int(val interface{}) (int64, error) {
	switch v := val.(type) {
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	case int:
		return int64(v), nil
	case uint:
		return int64(v), nil
	case string:
		return strconv.ParseInt(v, 10, 64)
	case float64:
		return int64(v), nil
	case float32:
		return int64(v), nil
	case int64:
		return v, nil
	case uint64:
		if v <= uint64(maxInt) {
			return int64(v), nil
		}
	case int32:
		return int64(v), nil
	case uint32:
		return int64(v), nil
	case int8:
		return int64(v), nil
	case uint8:
		return int64(v), nil
	case int16:
		return int64(v), nil
	case uint16:
		return int64(v), nil
	case nil:
		return 0, nil
	}
	return 0, fmt.Errorf("can't convert %#v to int", val)
}

func Object(val interface{}) (map[string]interface{}, error) {
	switch v := val.(type) {
	case map[string]interface{}:
		return v, nil
	}
	return map[string]interface{}{}, fmt.Errorf("can't convert %#v to object", val)
}

func Slice(val interface{}) ([]interface{}, error) {
	switch v := val.(type) {
	case []interface{}:
		return v, nil
	}
	return []interface{}{}, fmt.Errorf("can't convert %#v to slice", val)
}

func String(val interface{}) (string, error) {
	switch v := val.(type) {
	case string:
		return v, nil
	case []byte:
		return string(v), nil
	case int, uint, int8, uint8, int16, uint16, int32, uint32, int64, uint64, float32, float64, bool:
		return fmt.Sprintf("%v", v), nil
	case nil:
		return "", nil
	}
	return "", fmt.Errorf("can't convert %#v to string", val)
}

func Uint(val interface{}) (uint64, error) {
	switch v := val.(type) {
	case uint:
		return uint64(v), nil
	case int:
		if v >= minUint {
			return uint64(v), nil
		}
	case uint64:
		return v, nil
	case int64:
		if v >= minUint {
			return uint64(v), nil
		}
	case string:
		return strconv.ParseUint(v, 10, 64)
	case float64:
		if v >= 0 {
			return uint64(v), nil
		}
	case float32:
		if v >= 0 {
			return uint64(v), nil
		}
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	case int32:
		if v >= minUint {
			return uint64(v), nil
		}
	case uint32:
		return uint64(v), nil
	case int8:
		if v >= minUint {
			return uint64(v), nil
		}
	case uint8:
		return uint64(v), nil
	case int16:
		if v >= minUint {
			return uint64(v), nil
		}
	case uint16:
		return uint64(v), nil
	case nil:
		return 0, nil
	}
	return 0, fmt.Errorf("can't convert %#v to uint", val)
}
