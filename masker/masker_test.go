package masker

import (
	"github.com/warthog618/config"
	"testing"
)

// A simple mock reader wrapping an accessible map.
type mockReader struct {
	Config map[string]string
}

func (m *mockReader) Contains(key string) bool {
	_, ok := m.Config[key]
	return ok
}

func (m *mockReader) Read(key string) (interface{}, bool) {
	v, ok := m.Config[key]
	return v, ok
}

func TestNew(t *testing.T) {
	r := mockReader{map[string]string{}}
	m := New(&r, false)
	if m == nil {
		t.Fatalf("new returned nil")
	}
	// test provides config.Reader interface.
	cfg := config.New()
	cfg.AppendReader(m)
}

func TestMaskUnconditional(t *testing.T) {
	r := mockReader{map[string]string{"a": "a", "foo.b": "foo.b"}}
	m := New(&r, false)
	if m.Mask("a") {
		t.Errorf("reader masks 'a'")
	}
	m.AddMask("a")
	if !m.Mask("a") {
		t.Errorf("reader doesn't mask 'a'")
	}
	if m.Mask("foo.b") {
		t.Errorf("reader masks 'foo.b'")
	}
	m.AddMask("foo.b")
	if !m.Mask("foo.b") {
		t.Errorf("reader doesn't mask 'foo.b'")
	}
	if m.Mask("bar.c") {
		t.Errorf("reader masks 'bar.c'")
	}
	m.AddMask("bar.c")
	if !m.Mask("bar.c") {
		t.Errorf("reader doesn't mask 'bar.c'")
	}
}

func TestMaskConditional(t *testing.T) {
	r := mockReader{map[string]string{"a": "a", "foo.b": "foo.b"}}
	m := New(&r, true)
	if m.Mask("a") {
		t.Errorf("reader masks 'a'")
	}
	m.AddMask("a")
	if !m.Mask("a") {
		t.Errorf("reader doesn't mask 'a'")
	}
	if m.Mask("foo.b") {
		t.Errorf("reader masks 'foo.b'")
	}
	m.AddMask("foo.b")
	if !m.Mask("foo.b") {
		t.Errorf("reader doesn't mask 'foo.b'")
	}
	if m.Mask("bar.c") {
		t.Errorf("reader masks 'bar.c'")
	}
	m.AddMask("bar.c")
	if m.Mask("bar.c") {
		t.Errorf("reader unconditionally masks 'bar.c'")
	}
}
