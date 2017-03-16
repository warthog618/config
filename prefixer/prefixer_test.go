// Copyright © 2017 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package prefixer

import (
	"testing"

	"github.com/warthog618/config"
)

// A simple mock reader wrapping an accessible map.
type mockReader struct {
	Config map[string]string
}

func (m *mockReader) Read(key string) (interface{}, bool) {
	v, ok := m.Config[key]
	return v, ok
}

func TestNew(t *testing.T) {
	m := mockReader{map[string]string{}}
	p := New("blah.", &m)
	if p == nil {
		t.Fatalf("new returned nil")
	}
	// test provides config.Reader interface.
	cfg := config.New()
	cfg.AppendReader(p)
}

func TestRead(t *testing.T) {
	m := mockReader{map[string]string{"a": "a", "foo.b": "foo.b"}}
	p := New("blah.", &m)
	if v, ok := p.Read("blah.a"); ok {
		if v != "a" {
			t.Errorf("reader read 'blah.a' returned '%v', expected, %v", v, "a")
		}
	} else {
		t.Errorf("read 'blah.a' failed")
	}
	if v, ok := p.Read("blah.foo.b"); ok {
		if v != "foo.b" {
			t.Errorf("reader read 'blah.foo.b' returned '%v', expected, %v", v, "foo.b")
		}
	} else {
		t.Errorf("read 'blah.foo.b' failed")
	}

	v, ok := p.Read("notblah.a")
	if ok {
		t.Errorf("reader unexpectedly read 'notblah.a'")
	}
	if v != nil {
		t.Errorf("reader read 'notblah.a' returned '%v', expected, %v", v, nil)
	}

	v, ok = p.Read("notblah.foo.a")
	if ok {
		t.Errorf("reader unexpectedly read 'notblah.foo.a'")
	}
	if v != nil {
		t.Errorf("reader read 'notblah.foo.a' returned '%v', expected, %v", v, nil)
	}

	v, ok = p.Read("")
	if ok {
		t.Errorf("reader unexpectedly read ''")
	}
	if v != nil {
		t.Errorf("reader read '' returned '%v', expected, %v", v, nil)
	}

	v, ok = p.Read("a")
	if ok {
		t.Errorf("reader unexpectedly read 'a'")
	}
	if v != nil {
		t.Errorf("reader read 'a' returned '%v', expected, %v", v, nil)
	}

}
