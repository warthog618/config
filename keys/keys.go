// Copyright © 2017 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package keys provides utilities to manipulate key strings.
package keys

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// Treatment defines how the case of key words is treated by the replacer.
type Treatment int

const (
	// Unchanged means the case is left unchange.
	Unchanged Treatment = iota
	// LowerCase means the whole key is forced to lowercase.
	LowerCase
	// UpperCase means the whole key is forced to uppercase.
	UpperCase
	// LowerCamelCase means the first word is lower case, and subsequent words
	// have the leading character capitalized.
	LowerCamelCase
	// UpperCamelCase means all words have the first character capitalized.
	UpperCamelCase
)

// Replacer defines an interface that replaces a key from one namespace
// with a key from another.
// This is a sub-interface of strings.Replacer that lacks the io.Writer.
type Replacer interface {
	Replace(old string) string
}

type replacer struct {
	fromSep string
	toSep   string
}

// A fast replacer for when no treatment is required.
func (r *replacer) Replace(from string) string {
	return strings.Replace(from, r.fromSep, r.toSep, -1)
}

type treatedReplacer struct {
	replacer
	treatment Treatment
}

// CamelWord returns the CamelCase version of a word, i.e. the first
// letter capitalised and other characters lowercase.
func CamelWord(s string) string {
	r, n := utf8.DecodeRuneInString(s)
	return string(unicode.ToUpper(r)) + strings.ToLower(s[n:])
}

func (r *treatedReplacer) Replace(from string) string {
	switch r.treatment {
	case LowerCase:
		return strings.ToLower(strings.Replace(from, r.fromSep, r.toSep, -1))
	case UpperCase:
		return strings.ToUpper(strings.Replace(from, r.fromSep, r.toSep, -1))
	case LowerCamelCase:
		path := strings.Split(from, r.fromSep)
		for idx, p := range path {
			if idx == 0 {
				path[idx] = strings.ToLower(p)
			} else {
				path[idx] = CamelWord(p)
			}
		}
		return strings.Join(path, r.toSep)
	case UpperCamelCase:
		path := strings.Split(from, r.fromSep)
		for idx, p := range path {
			path[idx] = CamelWord(p)
		}
		return strings.Join(path, r.toSep)
	default:
		return strings.Replace(from, r.fromSep, r.toSep, -1)
	}
}

// NewReplacer returns a Replacer that will replace the from separator (fromSep)
// with the to separator (toSep), and then apply the case treatment.
func NewReplacer(fromSep, toSep string, treatment Treatment) Replacer {
	if treatment == Unchanged {
		return &replacer{fromSep, toSep}
	}
	return &treatedReplacer{replacer{fromSep, toSep}, treatment}
}
