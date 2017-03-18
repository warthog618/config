// Copyright © 2017 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package env

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/warthog618/config"
	"github.com/warthog618/config/keys"
)

func setup(prefix string) {
	os.Clearenv()
	os.Setenv(prefix+"LEAF", "42")
	os.Setenv(prefix+"SLICE", "a:b")
	os.Setenv(prefix+"NESTED_LEAF", "44")
	os.Setenv(prefix+"NESTED_SLICE", "c:d")
}

func TestNew(t *testing.T) {
	setup("CFGENV_")
	e, err := New("CFGENV_")
	if err != nil {
		t.Fatalf("new returned error %v", err)
	}
	if len(e.config) != 4 {
		t.Errorf("incorrect number of leaves, expected 4, got %v", len(e.config))
	}
	// test provides config.Reader interface.
	cfg := config.New()
	cfg.AppendReader(e)
}

func TestReaderRead(t *testing.T) {
	prefix := "CFGENV_"
	setup(prefix)
	e, err := New(prefix)
	if err != nil {
		t.Fatalf("new returned error %v", err)
	}
	// leaf
	expected := "42"
	if v, ok := e.Read("leaf"); ok {
		if v != expected {
			t.Errorf("read leaf %v, expected %v", v, expected)
		}
	} else {
		t.Errorf("failed to read leaf")
	}
	expectedSlice := []string{"a", "b"}
	if v, ok := e.Read("slice"); ok {
		if !reflect.DeepEqual(v, expectedSlice) {
			t.Errorf("read slice %v, expected %v", v, expectedSlice)
		}
	} else {
		t.Errorf("failed to read slice")
	}
	if v, ok := e.Read("nested"); ok {
		t.Errorf("read nested, got %v", v)
	} else if v != nil {
		t.Errorf("returned non-nil on failed read for nested, got %v", v)
	}
	expected = "44"
	if v, ok := e.Read("nested.leaf"); ok {
		if v != expected {
			t.Errorf("read nested.leaf %v, expected %v", v, expected)
		}
	} else {
		t.Errorf("failed to read nested.leaf")
	}
	expectedSlice = []string{"c", "d"}
	if v, ok := e.Read("nested.slice"); ok {
		if !reflect.DeepEqual(v, expectedSlice) {
			t.Errorf("read slice %v, expected %v", v, expectedSlice)
		}
	} else {
		t.Errorf("failed to read slice")
	}
	// node
	if v, ok := e.Read("nested"); ok {
		t.Errorf("contains nested - got %v", v)
	}
	// neither
	if v, ok := e.Read("nonsense"); ok {
		t.Errorf("contains nonsense - got %v", v)
	}
}

func TestReaderSetCfgKeyReplacer(t *testing.T) {
	prefix := "CFGENV_"
	setup(prefix)
	e, err := New(prefix)
	if err != nil {
		t.Fatalf("new returned error %v", err)
	}
	// null - leaves case untouched
	e.SetCfgKeyReplacer(keys.NewReplacer("_", "_", keys.Unchanged))
	expected := "44"
	if v, ok := e.Read("NESTED_LEAF"); ok {
		if v != expected {
			t.Errorf("read NESTED_LEAF %v, expected %v", v, expected)
		}
	} else {
		t.Errorf("failed to read NESTED_LEAF")
	}
	// multi
	e.SetCfgKeyReplacer(keys.NewReplacer("_", "_X_", keys.LowerCase))
	if v, ok := e.Read("nested_x_leaf"); ok {
		if v != expected {
			t.Errorf("read nested_x_leaf %v, expected %v", v, expected)
		}
	} else {
		t.Errorf("failed to read nested_x_leaf")
	}
	// none
	e.SetCfgKeyReplacer(keys.NewReplacer("_", "", keys.LowerCase))
	if v, ok := e.Read("nestedleaf"); ok {
		if v != expected {
			t.Errorf("read nestedleaf %v, expected %v", v, expected)
		}
	} else {
		t.Errorf("failed to read nestedleaf")
	}
	// standard
	e.SetCfgKeyReplacer(keys.NewReplacer("_", ".", keys.LowerCase))
	expected = "44"
	if v, ok := e.Read("nested.leaf"); ok {
		if v != expected {
			t.Errorf("read nested.leaf %v, expected %v", v, expected)
		}
	} else {
		t.Errorf("failed to read nested.leaf")
	}
	// multi
	e.SetCfgKeyReplacer(strings.NewReplacer("TED_", "."))
	if v, ok := e.Read("NES.LEAF"); ok {
		if v != expected {
			t.Errorf("read NES.LEAF %v, expected %v", v, expected)
		}
	} else {
		t.Errorf("failed to read NES.LEAF")
	}
}

func TestReaderListSeparator(t *testing.T) {
	prefix := "CFGENV_"
	setup(prefix)
	os.Setenv(prefix+"SLICE", "a:#b")
	e, err := New(prefix)
	if err != nil {
		t.Fatalf("new returned error %v", err)
	}
	// single
	e.SetListSeparator(":")
	if v, ok := e.Read("slice"); ok {
		if !reflect.DeepEqual(v, []string{"a", "#b"}) {
			t.Errorf("read slice %v, expected %v", v, []string{"a", "#b"})
		}
	} else {
		t.Errorf("failed to read slice")
	}
	// multi
	e.SetListSeparator(":#")
	if v, ok := e.Read("slice"); ok {
		if !reflect.DeepEqual(v, []string{"a", "b"}) {
			t.Errorf("read slice %v, expected %v", v, []string{"a", "b"})
		}
	} else {
		t.Errorf("failed to read slice")
	}
	// none
	e.SetListSeparator("")
	if v, ok := e.Read("slice"); ok {
		if v != "a:#b" {
			t.Errorf("read slice %v, expected %v", v, "a:#b")
		}
	} else {
		t.Errorf("failed to read slice")
	}
}

func TestReaderSetPrefix(t *testing.T) {
	prefix := "CFGENV_"
	setup(prefix)
	e, err := New(prefix)
	if err != nil {
		t.Fatalf("new returned error %v", err)
	}
	expected := "44"
	if v, ok := e.Read("nested.leaf"); ok {
		if v != expected {
			t.Errorf("read nested.leaf %v, expected %v", v, expected)
		}
	} else {
		t.Errorf("failed to read nested.leaf")
	}
	e.SetEnvPrefix("CFG")
	if v, ok := e.Read("env.nested.leaf"); ok {
		if v != expected {
			t.Errorf("read env.nested.leaf %v, expected %v", v, expected)
		}
	} else {
		t.Errorf("failed to read env.nested.leaf")
	}
	// none
	e.SetEnvPrefix("")
	if v, ok := e.Read("cfgenv.nested.leaf"); ok {
		if v != expected {
			t.Errorf("read cfgenv.nested.leaf %v, expected %v", v, expected)
		}
	} else {
		t.Errorf("failed to read cfgenv.nested.leaf")
	}
}
