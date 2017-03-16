// Copyright © 2017 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package dict

import (
	"testing"

	"github.com/warthog618/config"
)

func TestNew(t *testing.T) {
	r := New()
	if r == nil {
		t.Fatalf("new returned nil")
	}
	// test provides config.Reader interface.
	cfg := config.New()
	cfg.AppendReader(r)
}

func TestReader(t *testing.T) {
	r := New()
	if v, ok := r.Read("a"); ok {
		t.Errorf("read non-existent a, got %v", v)
	}
	r.Set("a", 1)
	if v, ok := r.Read("a"); ok {
		if v != 1 {
			t.Errorf("read incorrect value for a, got %v, expected %v", v, 1)
		}
	}
}
