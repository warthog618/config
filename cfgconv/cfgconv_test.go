package cfgconv

import (
	"reflect"
	"testing"
)

func assertBool(t *testing.T, val interface{}, expected bool, comment string) {
	if result, err := Bool(val); err != nil {
		t.Errorf("conversion failed for %s with error %v", comment, err)
	} else {
		if result != expected {
			t.Errorf("conversion failed for %s, expected %v got %v", comment, expected, result)
		}
	}
}

func refuteBool(t *testing.T, val interface{}, comment string) {
	if result, err := Bool(val); err == nil {
		t.Errorf("conversion succeeded for %s , got %v", comment, result)
	}
}

func TestBool(t *testing.T) {
	// success cases
	assertBool(t, false, false, "bool false")
	assertBool(t, true, true, "bool true")
	assertBool(t, "false", false, "string false")
	assertBool(t, "true", true, "string true")
	assertBool(t, nil, false, "nil")
	assertBool(t, "0", false, "string 0")
	assertBool(t, "1", true, "string 1")
	assertBool(t, int(0), false, "int 0")
	assertBool(t, int(1), true, "int 1")
	assertBool(t, int(42), true, "int positive")
	assertBool(t, int(-42), true, "int negative")
	assertBool(t, uint(0), false, "uint 0")
	assertBool(t, uint(1), true, "uint 1")
	assertBool(t, uint(42), true, "uint positive")
	assertBool(t, int8(0), false, "int8 0")
	assertBool(t, int8(1), true, "int8 1")
	assertBool(t, uint8(0), false, "uint8 0")
	assertBool(t, uint8(1), true, "uint8 1")
	assertBool(t, int16(0), false, "int16 0")
	assertBool(t, int16(1), true, "int16 1")
	assertBool(t, uint16(0), false, "uint16 0")
	assertBool(t, uint16(1), true, "uint16 1")
	assertBool(t, int32(0), false, "int32 0")
	assertBool(t, int32(1), true, "int32 1")
	assertBool(t, uint32(0), false, "uint32 0")
	assertBool(t, uint32(1), true, "uint32 1")
	assertBool(t, int64(0), false, "int64 0")
	assertBool(t, int64(1), true, "int64 1")
	assertBool(t, uint64(0), false, "uint64 0")
	assertBool(t, uint64(1), true, "uint64 1")
	// failure cases
	refuteBool(t, float64(0), "float64 0")
	refuteBool(t, float64(1), "float64 1")
	refuteBool(t, float32(0), "float32 0")
	refuteBool(t, float32(1), "float32 1")
	refuteBool(t, "junk", "string junk")
}

func assertFloat(t *testing.T, val interface{}, expected float64, comment string) {
	if result, err := Float(val); err != nil {
		t.Errorf("conversion failed for %s with error %v", comment, err)
	} else {
		if result != expected {
			t.Errorf("conversion failed for %s, expected %v got %v", comment, expected, result)
		}
	}
}

func refuteFloat(t *testing.T, val interface{}, comment string) {
	if result, err := Float(val); err == nil {
		t.Errorf("conversion succeeded for %s , got %v", comment, result)
	}
}

func TestFloat(t *testing.T) {
	pi := float64(3.1415)
	pi32 := float32(3.1415)
	// success cases
	assertFloat(t, pi, pi, "float64 pi")
	assertFloat(t, pi32, float64(pi32), "float32 pi")
	assertFloat(t, false, 0, "bool false")
	assertFloat(t, true, 1, "bool true")
	assertFloat(t, "3.1415", pi, "string pi")
	assertFloat(t, "42", 42, "string int")
	assertFloat(t, "-42", -42, "string int negative")
	assertFloat(t, int(42), 42, "int")
	assertFloat(t, int(-42), -42, "int negative")
	assertFloat(t, uint(42), 42, "uint")
	assertFloat(t, int8(42), 42, "int8")
	assertFloat(t, int8(-42), -42, "int8 negative")
	assertFloat(t, uint8(42), 42, "uint8")
	assertFloat(t, int16(42), 42, "int16")
	assertFloat(t, int16(-42), -42, "int16 negative")
	assertFloat(t, uint16(42), 42, "uint16")
	assertFloat(t, int32(42), 42, "int32")
	assertFloat(t, int32(-42), -42, "int32 negative")
	assertFloat(t, uint32(42), 42, "uint32")
	assertFloat(t, int64(42), 42, "int64")
	assertFloat(t, int64(-42), -42, "int64 negative")
	assertFloat(t, uint64(42), 42, "uint64")
	assertFloat(t, nil, 0, "nil")
	// failure cases
	refuteFloat(t, "junk", "string junk")
	refuteFloat(t, []int{42}, "slice")
}

func assertInt(t *testing.T, val interface{}, expected int64, comment string) {
	if result, err := Int(val); err != nil {
		t.Errorf("conversion failed for %s with error %v", comment, err)
	} else {
		if result != expected {
			t.Errorf("conversion failed for %s, expected %v got %v", comment, expected, result)
		}
	}
}

func refuteInt(t *testing.T, val interface{}, comment string) {
	if result, err := Int(val); err == nil {
		t.Errorf("conversion succeeded for %s , got %v", comment, result)
	}
}

func TestInt(t *testing.T) {
	// success cases
	assertInt(t, false, 0, "bool false")
	assertInt(t, true, 1, "bool true")
	assertInt(t, "42", 42, "string int")
	assertInt(t, "-42", -42, "string int negative")
	assertInt(t, int(42), 42, "int")
	assertInt(t, int(-42), -42, "int negative")
	assertInt(t, uint(42), 42, "uint")
	assertInt(t, int8(42), 42, "int8")
	assertInt(t, int8(-42), -42, "int8 negative")
	assertInt(t, uint8(42), 42, "uint8")
	assertInt(t, int16(42), 42, "int16")
	assertInt(t, int16(-42), -42, "int16 negative")
	assertInt(t, uint16(42), 42, "uint16")
	assertInt(t, int32(42), 42, "int32")
	assertInt(t, int32(-42), -42, "int32 negative")
	assertInt(t, uint32(42), 42, "uint32")
	assertInt(t, int64(42), 42, "int64")
	assertInt(t, int64(-42), -42, "int64 negative")
	assertInt(t, uint64(42), 42, "uint64")
	assertInt(t, float64(42), 42, "float64")
	assertInt(t, float64(0), 0, "float64 zero")
	assertInt(t, float64(-42), -42, "float64 negative")
	assertInt(t, float64(42.6), 42, "float64 truncate")
	assertInt(t, float64(-42.6), -42, "float64 truncate negative")
	assertInt(t, float32(42), 42, "float32")
	assertInt(t, float32(-42), -42, "float32 negative")
	assertInt(t, float32(42.6), 42, "float32 truncate")
	assertInt(t, float32(-42.6), -42, "float32 truncate negative")
	assertInt(t, nil, 0, "nil")
	// failure cases
	refuteInt(t, "junk", "string junk")
	refuteInt(t, "42.5", "string float")
	refuteInt(t, []int{42}, "slice")
}

func assertObject(t *testing.T, val interface{}, expected map[string]interface{}, comment string) {
	if result, err := Object(val); err != nil {
		t.Errorf("conversion failed for %s with error %v", comment, err)
	} else {
		if reflect.DeepEqual(result, expected) == false {
			t.Errorf("conversion failed for %s, expected %v got %v", comment, expected, result)
		}
	}
}

func refuteObject(t *testing.T, val interface{}, comment string) {
	if result, err := Object(val); err == nil {
		t.Errorf("conversion succeeded for %s , got %v", comment, result)
	}
}

func TestObject(t *testing.T) {
	obj := map[string]interface{}{"key": 42, "slice": []string{"one", "two"}}
	// success cases
	assertObject(t, obj, obj, "object")
	// failure cases
	refuteObject(t, true, "bool true")
	refuteObject(t, false, "bool false")
	refuteObject(t, "42", "string int")
	refuteObject(t, int(42), "int")
	refuteObject(t, int(-42), "int negative")
	refuteObject(t, uint(42), "uint")
	refuteObject(t, int8(42), "int8")
	refuteObject(t, int8(-42), "int8 negative")
	refuteObject(t, uint8(42), "uint8")
	refuteObject(t, int16(42), "int16")
	refuteObject(t, int16(-42), "int16 negative")
	refuteObject(t, uint16(42), "uint16")
	refuteObject(t, int32(42), "int32")
	refuteObject(t, int32(-42), "int32 negative")
	refuteObject(t, uint32(42), "uint32")
	refuteObject(t, int64(42), "int64")
	refuteObject(t, int64(-42), "int64 negative")
	refuteObject(t, uint64(42), "uint64")
	refuteObject(t, float64(42), "float64")
	refuteObject(t, float64(0), "float64 zero")
	refuteObject(t, float64(-42), "float64 negative")
	refuteObject(t, float64(42.6), "float64 truncate")
	refuteObject(t, float64(-42.6), "float64 truncate negative")
	refuteObject(t, float32(42), "float32")
	refuteObject(t, float32(-42), "float32 negative")
	refuteObject(t, float32(42.6), "float32 truncate")
	refuteObject(t, float32(-42.6), "float32 truncate negative")
	refuteObject(t, nil, "nil")
}

func assertSlice(t *testing.T, val interface{}, expected []interface{}, comment string) {
	if result, err := Slice(val); err != nil {
		t.Errorf("conversion failed for %s with error %v", comment, err)
	} else {
		if reflect.DeepEqual(result, expected) == false {
			t.Errorf("conversion failed for %s, expected %v got %v", comment, expected, result)
		}
	}
}

func refuteSlice(t *testing.T, val interface{}, comment string) {
	if result, err := Slice(val); err == nil {
		t.Errorf("conversion succeeded for %s , got %v", comment, result)
	}
}

func TestSlice(t *testing.T) {
	intSlice := []interface{}{[]int{1, 2, 3}}
	stringSlice := []interface{}{[]string{"one", "two"}}
	uintSlice := []interface{}{[]int{1, 2, 3}}
	// success cases
	assertSlice(t, intSlice, intSlice, "object")
	assertSlice(t, stringSlice, stringSlice, "object")
	assertSlice(t, uintSlice, uintSlice, "object")
	// failure cases
	refuteSlice(t, true, "bool true")
	refuteSlice(t, false, "bool false")
	refuteSlice(t, "42", "string int")
	refuteSlice(t, int(42), "int")
	refuteSlice(t, int(-42), "int negative")
	refuteSlice(t, uint(42), "uint")
	refuteSlice(t, int8(42), "int8")
	refuteSlice(t, int8(-42), "int8 negative")
	refuteSlice(t, uint8(42), "uint8")
	refuteSlice(t, int16(42), "int16")
	refuteSlice(t, int16(-42), "int16 negative")
	refuteSlice(t, uint16(42), "uint16")
	refuteSlice(t, int32(42), "int32")
	refuteSlice(t, int32(-42), "int32 negative")
	refuteSlice(t, uint32(42), "uint32")
	refuteSlice(t, int64(42), "int64")
	refuteSlice(t, int64(-42), "int64 negative")
	refuteSlice(t, uint64(42), "uint64")
	refuteSlice(t, float64(42), "float64")
	refuteSlice(t, float64(0), "float64 zero")
	refuteSlice(t, float64(-42), "float64 negative")
	refuteSlice(t, float64(42.6), "float64 truncate")
	refuteSlice(t, float64(-42.6), "float64 truncate negative")
	refuteSlice(t, float32(42), "float32")
	refuteSlice(t, float32(-42), "float32 negative")
	refuteSlice(t, float32(42.6), "float32 truncate")
	refuteSlice(t, float32(-42.6), "float32 truncate negative")
	refuteSlice(t, nil, "nil")
}
func assertString(t *testing.T, val interface{}, expected string, comment string) {
	if result, err := String(val); err != nil {
		t.Errorf("conversion failed for %s with error %v", comment, err)
	} else {
		if result != expected {
			t.Errorf("conversion failed for %s, expected %v got %v", comment, expected, result)
		}
	}
}

func refuteString(t *testing.T, val interface{}, comment string) {
	if result, err := String(val); err == nil {
		t.Errorf("conversion succeeded for %s , got %v", comment, result)
	}
}

func TestString(t *testing.T) {
	// success cases
	assertString(t, false, "false", "bool false")
	assertString(t, true, "true", "bool true")
	assertString(t, "junk", "junk", "string junk")
	assertString(t, "42", "42", "string int")
	assertString(t, "-42", "-42", "string int negative")
	assertString(t, "42.5", "42.5", "string float")
	assertString(t, []byte{0x31, 0x32, 0x33, 0x34}, "1234", "byte slice")
	assertString(t, int(42), "42", "int")
	assertString(t, int(-42), "-42", "int negative")
	assertString(t, uint(42), "42", "uint")
	assertString(t, int8(42), "42", "int8")
	assertString(t, int8(-42), "-42", "int8 negative")
	assertString(t, uint8(42), "42", "uint8")
	assertString(t, int16(42), "42", "int16")
	assertString(t, int16(-42), "-42", "int16 negative")
	assertString(t, uint16(42), "42", "uint16")
	assertString(t, int32(42), "42", "int32")
	assertString(t, int32(-42), "-42", "int32 negative")
	assertString(t, uint32(42), "42", "uint32")
	assertString(t, int64(42), "42", "int64")
	assertString(t, int64(-42), "-42", "int64 negative")
	assertString(t, uint64(42), "42", "uint64")
	assertString(t, float64(42), "42", "float64")
	assertString(t, float64(-42), "-42", "float64 negative")
	assertString(t, float64(0), "0", "float64 zero")
	assertString(t, float64(42.6), "42.6", "float64")
	assertString(t, float32(42), "42", "float32")
	assertString(t, float32(-42), "-42", "float32 negative")
	assertString(t, float32(42.6), "42.6", "float32")
	assertString(t, nil, "", "nil")
	// failure cases
	refuteString(t, []int{42}, "slice")
}

func assertUint(t *testing.T, val interface{}, expected uint64, comment string) {
	if result, err := Uint(val); err != nil {
		t.Errorf("conversion failed for %s with error %v", comment, err)
	} else {
		if result != expected {
			t.Errorf("conversion failed for %s, expected %v got %v", comment, expected, result)
		}
	}
}

func refuteUint(t *testing.T, val interface{}, comment string) {
	if result, err := Uint(val); err == nil {
		t.Errorf("conversion succeeded for %s , got %v", comment, result)
	}
}

func TestUint(t *testing.T) {
	// success cases
	assertUint(t, false, 0, "bool false")
	assertUint(t, true, 1, "bool true")
	assertUint(t, "42", 42, "string int")
	assertUint(t, int(42), 42, "int")
	assertUint(t, uint(42), 42, "uint")
	assertUint(t, int8(42), 42, "int8")
	assertUint(t, uint8(42), 42, "uint8")
	assertUint(t, int16(42), 42, "int16")
	assertUint(t, uint16(42), 42, "uint16")
	assertUint(t, int32(42), 42, "int32")
	assertUint(t, uint32(42), 42, "uint32")
	assertUint(t, int64(42), 42, "int64")
	assertUint(t, uint64(42), 42, "uint64")
	assertUint(t, float64(42), 42, "float64")
	assertUint(t, float64(0), 0, "float64 zero")
	assertUint(t, float64(42.6), 42, "float64 truncate")
	assertUint(t, float32(42), 42, "float32")
	assertUint(t, float32(42.6), 42, "float32 truncate")
	assertUint(t, nil, 0, "nil")
	// failure cases
	refuteUint(t, "-42", "string int negative")
	refuteUint(t, int(-42), "int negative")
	refuteUint(t, int8(-42), "int8 negative")
	refuteUint(t, int16(-42), "int16 negative")
	refuteUint(t, int32(-42), "int32 negative")
	refuteUint(t, int64(-42), "int64 negative")
	refuteUint(t, float64(-42), "float64 negative")
	refuteUint(t, float32(-42), "float32 negative")
	refuteUint(t, "junk", "string junk")
	refuteUint(t, "42.5", "string float")
	refuteUint(t, []int{42}, "slice")
}
