// Copyright © 2018 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package list_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/warthog618/config/list"
)

func TestNewSplitter(t *testing.T) {
	patterns := []struct {
		name string
		sep  string
		in   string
		out  interface{}
	}{
		{"comma ints", ",", "1,2,3,4", []string{"1", "2", "3", "4"}},
		{"not comma ints", ":", "1,2,3,4", "1,2,3,4"},
		{"none", "", "1,2,3,4", []string{"1", ",", "2", ",", "3", ",", "4"}},
	}
	for _, p := range patterns {
		s := list.NewSplitter(p.sep)
		out := s.Split(p.in)
		assert.Equal(t, p.out, out, p.name)
	}

}
