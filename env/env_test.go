package env

import (
	"github.com/warthog618/config"
	"os"
	"reflect"
	"testing"
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
		t.Fatalf("new returned error", err)
	}
	if len(e.nodes) != 1 {
		t.Errorf("incorrect number of nodes, expected 1, got %v", len(e.nodes))
	}
	if len(e.config) != 4 {
		t.Errorf("incorrect number of leaves, expected 4, got %v", len(e.config))
	}
	// test provides config.Reader interface.
	cfg := config.New()
	cfg.AppendReader(e)
}

func TestReaderContains(t *testing.T) {
	prefix := "CFGENV_"
	setup(prefix)
	c, err := New(prefix)
	if err != nil {
		t.Fatalf("new returned error", err)
	}
	// leaf
	if !c.Contains("leaf") {
		t.Errorf("doesn't contain leaf")
	}
	if !c.Contains("nested.leaf") {
		t.Errorf("doesn't contain nested.leaf")
	}
	// node
	if !c.Contains("nested") {
		t.Errorf("doesn't contain nested.int")
	}
	// neither
	if c.Contains("nonsense") {
		t.Errorf("contains nonsense")
	}
}

func TestReaderRead(t *testing.T) {
	prefix := "CFGENV_"
	setup(prefix)
	c, err := New(prefix)
	if err != nil {
		t.Fatalf("new returned error", err)
	}
	// leaf
	if v, ok := c.Read("leaf"); ok {
		if v != "42" {
			t.Errorf("read leaf %v, expected %v", v, "42")
		}
	} else {
		t.Errorf("failed to read leaf")
	}
	if v, ok := c.Read("slice"); ok {
		if !reflect.DeepEqual(v, []string{"a", "b"}) {
			t.Errorf("read slice %v, expected %v", v, []string{"a", "b"})
		}
	} else {
		t.Errorf("failed to read slice")
	}
	if v, ok := c.Read("nested.leaf"); ok {
		if v != "44" {
			t.Errorf("read nested.leaf %v, expected %v", v, "44")
		}
	} else {
		t.Errorf("failed to read nested.leaf")
	}
	if v, ok := c.Read("nested.slice"); ok {
		if !reflect.DeepEqual(v, []string{"c", "d"}) {
			t.Errorf("read slice %v, expected %v", v, []string{"c", "d"})
		}
	} else {
		t.Errorf("failed to read slice")
	}
	// node
	if v, ok := c.Read("nested"); ok {
		t.Errorf("contains nested - got %v", v)
	}
	// neither
	if v, ok := c.Read("nonsense"); ok {
		t.Errorf("contains nonsense - got %v", v)
	}
}

func TestReaderSetCfgSeparator(t *testing.T) {
	prefix := "CFGENV_"
	setup(prefix)
	c, err := New(prefix)
	if err != nil {
		t.Fatalf("new returned error", err)
	}
	// single
	c.SetCfgSeparator("_")
	if v, ok := c.Read("nested_leaf"); ok {
		if v != "44" {
			t.Errorf("read nested_leaf %v, expected %v", v, "44")
		}
	} else {
		t.Errorf("failed to read nested_leaf")
	}
	// multi
	c.SetCfgSeparator("_X_")
	if v, ok := c.Read("nested_x_leaf"); ok {
		if v != "44" {
			t.Errorf("read nested_x_leaf %v, expected %v", v, "44")
		}
	} else {
		t.Errorf("failed to read nested_x_leaf")
	}
	// none
	c.SetCfgSeparator("")
	if v, ok := c.Read("nestedleaf"); ok {
		if v != "44" {
			t.Errorf("read nestedleaf %v, expected %v", v, "44")
		}
	} else {
		t.Errorf("failed to read nestedleaf")
	}
}
func TestReaderSetEnvSeparator(t *testing.T) {
	prefix := "CFGENV_"
	setup(prefix)
	c, err := New(prefix)
	if err != nil {
		t.Fatalf("new returned error", err)
	}
	// single
	c.SetEnvSeparator("_")
	if v, ok := c.Read("nested.leaf"); ok {
		if v != "44" {
			t.Errorf("read nested.leaf %v, expected %v", v, "44")
		}
	} else {
		t.Errorf("failed to read nested.leaf")
	}
	// multi
	c.SetEnvSeparator("TED_")
	if v, ok := c.Read("nes.leaf"); ok {
		if v != "44" {
			t.Errorf("read nes.leaf %v, expected %v", v, "44")
		}
	} else {
		t.Errorf("failed to read nes.leaf")
	}
	// none
	c.SetEnvSeparator("")
	if v, ok := c.Read("l.e.a.f"); ok {
		if v != "42" {
			t.Errorf("read l.e.a.f %v, expected %v", v, "42")
		}
	} else {
		t.Errorf("failed to read l.e.a.f")
	}
}

func TestReaderListSeparator(t *testing.T) {
	prefix := "CFGENV_"
	setup(prefix)
	os.Setenv(prefix+"SLICE", "a:#b")
	c, err := New(prefix)
	if err != nil {
		t.Fatalf("new returned error", err)
	}
	// single
	c.SetListSeparator(":")
	if v, ok := c.Read("slice"); ok {
		if !reflect.DeepEqual(v, []string{"a", "#b"}) {
			t.Errorf("read slice %v, expected %v", v, []string{"a", "#b"})
		}
	} else {
		t.Errorf("failed to read slice")
	}
	// multi
	c.SetListSeparator(":#")
	if v, ok := c.Read("slice"); ok {
		if !reflect.DeepEqual(v, []string{"a", "b"}) {
			t.Errorf("read slice %v, expected %v", v, []string{"a", "#b"})
		}
	} else {
		t.Errorf("failed to read slice")
	}
	// none
	c.SetListSeparator("")
	if v, ok := c.Read("slice"); ok {
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
	c, err := New(prefix)
	if err != nil {
		t.Fatalf("new returned error", err)
	}
	if v, ok := c.Read("nested.leaf"); ok {
		if v != "44" {
			t.Errorf("read nested.leaf %v, expected %v", v, "44")
		}
	} else {
		t.Errorf("failed to read nested.leaf")
	}
	c.SetEnvPrefix("CFG")
	if v, ok := c.Read("env.nested.leaf"); ok {
		if v != "44" {
			t.Errorf("read env.nested.leaf %v, expected %v", v, "44")
		}
	} else {
		t.Errorf("failed to read env.nested.leaf")
	}
	// none
	c.SetEnvPrefix("")
	if v, ok := c.Read("cfgenv.nested.leaf"); ok {
		if v != "44" {
			t.Errorf("read cfgenv.nested.leaf %v, expected %v", v, "44")
		}
	} else {
		t.Errorf("failed to read cfgenv.nested.leaf")
	}
}
