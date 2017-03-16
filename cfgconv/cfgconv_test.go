// Copyright © 2017 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package cfgconv

import (
	"reflect"
	"strings"
	"testing"
)

func assertBool(t *testing.T, v interface{}, expected bool, comment string) {
	if result, err := Bool(v); err != nil {
		t.Errorf("conversion failed for %s with error %v", comment, err)
	} else {
		if result != expected {
			t.Errorf("conversion failed for %s, expected %v got %v", comment, expected, result)
		}
	}
}

func refuteBool(t *testing.T, v interface{}, comment string) {
	if result, err := Bool(v); err == nil {
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
	refuteBool(t, "", "empty string")
}

func TestConvert(t *testing.T) {

	// int
	// good
	ct := reflect.TypeOf(0)
	var cin interface{}
	cin = "42"
	if cv, err := Convert(cin, ct); err == nil {
		if cv != 42 {
			t.Errorf("failed to convert '%v' to int, got %v", cin, cv)
		}
	} else {
		t.Errorf("failed to convert '%v' to int, got %v", cin, err)
	}
	// bad type
	cin = []int{}
	if cv, err := Convert(cin, ct); err == nil {
		t.Errorf("converted '%v' to int, got %v", cin, cv)
	} else {
		if !strings.Contains(err.Error(), "to int") {
			t.Errorf("overflow error doesn't indicate target type")
		}
		if cv != 0 {
			t.Errorf("didn't return zero on conversion to int, got %v", cv)
		}
	}
	// bad parse
	cin = "glob"
	if cv, err := Convert(cin, ct); err == nil {
		t.Errorf("converted '%v' to int, got %v", cin, cv)
	} else if cv != 0 {
		t.Errorf("didn't return zero on conversion to int, got %v", cv)
	}
	// overflow
	ct = reflect.TypeOf(int8(0))
	cin = 257
	if cv, err := Convert(cin, ct); err == nil {
		t.Errorf("converted '%v' to int, got %v", cin, cv)
	} else {
		if _, ok := err.(OverflowError); !ok {
			t.Errorf("didn't return overflow error, got %v", err)
		} else if !strings.Contains(err.Error(), "to int") {
			t.Errorf("overflow error doesn't indicate target type")
		}
		if cv != int8(0) {
			t.Errorf("didn't return zero on overflow to int, got %v", cv)
		}
	}

	// uint
	// good
	ct = reflect.TypeOf(uint(0))
	cin = "42"
	if cv, err := Convert(cin, ct); err == nil {
		if cv != uint(42) {
			t.Errorf("failed to convert '%v' to uint, got %v", cin, cv)
		}
	} else {
		t.Errorf("failed to convert '%v' to uint, got %v", cin, err)
	}
	// bad type
	cin = []int{}
	if cv, err := Convert(cin, ct); err == nil {
		t.Errorf("converted '%v' to uint, got %v", []int{}, cv)
	} else {
		if !strings.Contains(err.Error(), "to uint") {
			t.Errorf("overflow error doesn't indicate target type")
		}
		if cv != uint(0) {
			t.Errorf("didn't return zero on conversion to uint, got %v", cv)
		}
	}
	// bad parse
	cin = "glob"
	if cv, err := Convert(cin, ct); err == nil {
		t.Errorf("converted '%v' to uint, got %v", cin, cv)
	} else if cv != uint(0) {
		t.Errorf("didn't return zero on conversion to uint, got %v", cv)
	}
	// overflow
	ct = reflect.TypeOf(uint8(0))
	cin = 257
	if cv, err := Convert(cin, ct); err == nil {
		t.Errorf("converted '%v' to uint, got %v", cin, cv)
	} else {
		if _, ok := err.(OverflowError); !ok {
			t.Errorf("didn't return overflow error, got %v", err)
		} else if !strings.Contains(err.Error(), "to uint") {
			t.Errorf("overflow error doesn't indicate target type")
		}
		if cv != uint8(0) {
			t.Errorf("didn't return zero on overflow to uint, got %v", cv)
		}
	}

	// float
	// good
	ct = reflect.TypeOf(float32(0))
	cin = "42"
	if cv, err := Convert(cin, ct); err == nil {
		if cv != float32(42) {
			t.Errorf("failed to convert '%v' to float, got %v", cin, cv)
		}
	} else {
		t.Errorf("failed to convert '%v' to float, got %v", cin, err)
	}
	// bad type
	cin = []int{}
	if cv, err := Convert(cin, ct); err == nil {
		t.Errorf("converted '%v' to float, got %v", []int{}, cv)
	} else {
		if !strings.Contains(err.Error(), "to float") {
			t.Errorf("overflow error doesn't indicate target type")
		}
		if cv != float32(0) {
			t.Errorf("didn't return zero on conversion to float, got %v", cv)
		}
	}
	// bad parse
	cin = "glob"
	if cv, err := Convert(cin, ct); err == nil {
		t.Errorf("converted '%v' to float, got %v", cin, cv)
	} else if cv != float32(0) {
		t.Errorf("didn't return zero on conversion to float, got %v", cv)
	}
	// overflow
	cin = float64(340282356779733642748073463979561713664)
	if cv, err := Convert(cin, ct); err == nil {
		t.Errorf("converted '%v' to float, got %v", cin, cv)
	} else {
		if _, ok := err.(OverflowError); !ok {
			t.Errorf("didn't return overflow error, got %v", err)
		} else if !strings.Contains(err.Error(), "to float32") {
			t.Errorf("overflow error doesn't indicate target type")
		}
		if cv != float32(0) {
			t.Errorf("didn't return zero on overflow to float, got %v", cv)
		}
	}

	// string
	// good
	ct = reflect.TypeOf("")
	cin = 42
	if cv, err := Convert(cin, ct); err == nil {
		if cv != "42" {
			t.Errorf("failed to convert '%v' to string, got %v", cin, cv)
		}
	} else {
		t.Errorf("failed to convert '%v' to string, got %v", cin, err)
	}
	// bad type
	cin = []int{}
	if cv, err := Convert(cin, ct); err == nil {
		t.Errorf("converted '%v' to string, got %v", []int{}, cv)
	} else {
		if !strings.Contains(err.Error(), "to string") {
			t.Errorf("overflow error doesn't indicate target type")
		}
		if cv != "" {
			t.Errorf("didn't return empty string on conversion to string, got %v", cv)
		}
	}

	// bool
	// good
	ct = reflect.TypeOf(true)
	cin = 42
	if cv, err := Convert(cin, ct); err == nil {
		if cv != true {
			t.Errorf("failed to convert '%v' to bool, got %v", cin, cv)
		}
	} else {
		t.Errorf("failed to convert '%v' to bool, got %v", cin, err)
	}
	// bad type
	cin = []int{}
	if cv, err := Convert(cin, ct); err == nil {
		t.Errorf("converted '%v' to bool, got %v", []int{}, cv)
	} else {
		if !strings.Contains(err.Error(), "to bool") {
			t.Errorf("overflow error doesn't indicate target type")
		}
		if cv != false {
			t.Errorf("didn't return false on conversion to bool, got %v", cv)
		}
	}
	// bad parse
	cin = "glob"
	if cv, err := Convert(cin, ct); err == nil {
		t.Errorf("converted '%v' to bool, got %v", cin, cv)
	} else if cv != false {
		t.Errorf("didn't return false on conversion to bool, got %v", cv)
	}

	// slice
	// good
	ct = reflect.TypeOf([]int{})
	cin = []string{"1", "2", "3"}
	if cv, err := Convert(cin, ct); err == nil {
		if !reflect.DeepEqual(cv, []int{1, 2, 3}) {
			t.Errorf("failed to convert '%v' to slice, got %v", cin, cv)
		}
	} else {
		t.Errorf("failed to convert '%v' to slice, got %v", cin, err)
	}
	cin = "42"
	if cv, err := Convert(cin, ct); err == nil {
		if !reflect.DeepEqual(cv, []int{42}) {
			t.Errorf("failed to convert '%v' to slice, got %v", cin, cv)
		}
	} else {
		t.Errorf("failed to convert '%v' to slice, got %v", cin, err)
	}
	// bad type
	cin = 3
	if cv, err := Convert(cin, ct); err == nil {
		t.Errorf("converted '%v' to slice, got %v", []int{}, cv)
	} else if !reflect.DeepEqual(cv, []int(nil)) {
		t.Errorf("didn't return nil slice on conversion to slice, got %v %T", cv, cv)
	}
	// bad parse
	cin = []string{"1", "2", "3", "glob"}
	if cv, err := Convert(cin, ct); err == nil {
		t.Errorf("converted '%v' to slice, got %v", cin, cv)
	} else if !reflect.DeepEqual(cv, []int(nil)) {
		t.Errorf("didn't return nil slice on conversion to slice, got %v %T", cv, cv)
	}
	cin = "glob"
	if cv, err := Convert(cin, ct); err == nil {
		t.Errorf("converted '%v' to slice, got %v", cin, cv)
	} else if !reflect.DeepEqual(cv, []int(nil)) {
		t.Errorf("didn't return nil slice on conversion to slice, got %v %T", cv, cv)
	}
}

func assertFloat(t *testing.T, v interface{}, expected float64, comment string) {
	if result, err := Float(v); err != nil {
		t.Errorf("conversion failed for %s with error %v", comment, err)
	} else {
		if result != expected {
			t.Errorf("conversion failed for %s, expected %v got %v", comment, expected, result)
		}
	}
}

func refuteFloat(t *testing.T, v interface{}, comment string) {
	if result, err := Float(v); err == nil {
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
	refuteFloat(t, "", "empty string")
	refuteFloat(t, []int{42}, "slice")
}

func assertInt(t *testing.T, v interface{}, expected int64, comment string) {
	if result, err := Int(v); err != nil {
		t.Errorf("conversion failed for %s with error %v", comment, err)
	} else {
		if result != expected {
			t.Errorf("conversion failed for %s, expected %v got %v", comment, expected, result)
		}
	}
}

func refuteInt(t *testing.T, v interface{}, comment string) {
	if result, err := Int(v); err == nil {
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
	refuteInt(t, "42.5", "string float")
	refuteInt(t, "", "empty string")
	refuteInt(t, "junk", "string junk")
	refuteInt(t, []int{42}, "slice")
}

func assertSlice(t *testing.T, v interface{}, expected []interface{}, comment string) {
	if result, err := Slice(v); err != nil {
		t.Errorf("conversion failed for %s with error %v", comment, err)
	} else {
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("conversion failed for %s, expected %v got %v", comment, expected, result)
		}
	}
}

func refuteSlice(t *testing.T, v interface{}, comment string) {
	if result, err := Slice(v); err == nil {
		t.Errorf("conversion succeeded for %s , got %v", comment, result)
	}
}

func TestSlice(t *testing.T) {
	slice := []interface{}{[]int{1, 2, 3}}
	intSlice := []int{1, 2, -3}
	stringSlice := []string{"one", "two"}
	uintSlice := []int{1, 2, 3}
	// success cases
	assertSlice(t, slice, slice, "slice")
	assertSlice(t, intSlice, []interface{}{1, 2, -3}, "intSlice")
	assertSlice(t, stringSlice, []interface{}{"one", "two"}, "stringSlice")
	assertSlice(t, uintSlice, []interface{}{1, 2, 3}, "uintSlice")
	assertSlice(t, "42", []interface{}{"42"}, "string int")
	// failure cases
	refuteSlice(t, true, "bool true")
	refuteSlice(t, false, "bool false")
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
	refuteSlice(t, "", "empty string")
	refuteSlice(t, nil, "nil")
}
func assertString(t *testing.T, v interface{}, expected string, comment string) {
	if result, err := String(v); err != nil {
		t.Errorf("conversion failed for %s with error %v", comment, err)
	} else {
		if result != expected {
			t.Errorf("conversion failed for %s, expected %v got %v", comment, expected, result)
		}
	}
}

func refuteString(t *testing.T, v interface{}, comment string) {
	if result, err := String(v); err == nil {
		t.Errorf("conversion succeeded for %s , got %v", comment, result)
	}
}

func TestString(t *testing.T) {
	// success cases
	assertString(t, false, "false", "bool false")
	assertString(t, true, "true", "bool true")
	assertString(t, "junk", "junk", "string junk")
	assertString(t, "", "", "empty string")
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
	refuteUint(t, "", "empty string")
	refuteUint(t, []int{42}, "slice")
}
