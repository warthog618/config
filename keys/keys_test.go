// Copyright © 2017 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package keys

import "testing"

type replacerData struct {
	fromSep   string
	toSep     string
	treatment Treatment
	in        string
	expected  string
}

var testPatterns = []replacerData{
	{"-", ".", Unchanged, "Nested-key", "Nested.key"},
	{"_", ":", LowerCase, "Nested_Key", "nested:key"},
	{".", "_", UpperCase, "Nested.Key", "NESTED_KEY"},
	{"_", "", LowerCamelCase, "NesTed_Key", "nestedKey"},
	{".", "", UpperCamelCase, "nested.key", "NestedKey"},
	{".", "_", 7, "NesteD.Key", "NesteD_Key"},
}

func TestNewReplacer(t *testing.T) {
	for _, p := range testPatterns {
		r := NewReplacer(p.fromSep, p.toSep, p.treatment)
		v := r.Replace(p.in)
		if v != p.expected {
			t.Errorf("failed to replace %s, got %v, expected %v", p.in, v, p.expected)
		}
	}
}
