// Copyright © 2017 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package keys_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/warthog618/config/keys"
)

func TestNewReplacer(t *testing.T) {
	patterns := []struct {
		fromSep   string
		toSep     string
		treatment keys.Treatment
		in        string
		expected  string
	}{
		{"-", ".", keys.Unchanged, "Nested-key", "Nested.key"},
		{"_", ":", keys.LowerCase, "Nested_Key", "nested:key"},
		{".", "_", keys.UpperCase, "Nested.Key", "NESTED_KEY"},
		{"_", "", keys.LowerCamelCase, "NesTed_Key", "nestedKey"},
		{".", "", keys.UpperCamelCase, "nested.key", "NestedKey"},
		{".", "_", 7, "NesteD.Key", "NesteD_Key"},
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			r := keys.NewReplacer(p.fromSep, p.toSep, p.treatment)
			require.NotNil(t, r)
			v := r.Replace(p.in)
			assert.Equal(t, p.expected, v)
		}
		t.Run(p.in, f)
	}
}
