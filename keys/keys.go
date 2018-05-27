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

// Replacer replaces a string with replacements.
type Replacer struct {
	replace func(s string) string
}

// Replace performs the string replacement using the replace field.
func (r Replacer) Replace(s string) string {
	return r.replace(s)
}

type replacer func(s string) string

// NewReplacer returns a Replacer that will replace the from separator (fromSep)
// with the to separator (toSep), and then apply the case treatment.
func NewReplacer(fromSep, toSep string, treatment Treatment) Replacer {
	switch treatment {
	case LowerCase:
		return NewLowerCaseReplacer(fromSep, toSep)
	case UpperCase:
		return NewUpperCaseReplacer(fromSep, toSep)
	case LowerCamelCase:
		return NewLowerCamelCaseReplacer(fromSep, toSep)
	case UpperCamelCase:
		return NewUpperCamelCaseReplacer(fromSep, toSep)
	default:
		return NewUnchangedReplacer(fromSep, toSep)
	}
}

// NullReplacer performs no change to the key.
type NullReplacer struct {
}

// Replace simply returns the key.
func (r NullReplacer) Replace(s string) string {
	return s
}

// NewNullReplacer creates a Replacer that leaves the key unchanged.
func NewNullReplacer() NullReplacer {
	return NullReplacer{}
}

// NewUnchangedReplacer creates a Replacer that leaves case unchanged.
func NewUnchangedReplacer(fromSep, toSep string) Replacer {
	r := func(s string) string {
		return strings.Replace(s, fromSep, toSep, -1)
	}
	return Replacer{r}
}

// NewLowerCaseReplacer creates a Replacer that forces keys to lower case.
func NewLowerCaseReplacer(fromSep, toSep string) Replacer {
	r := func(s string) string {
		return strings.ToLower(strings.Replace(s, fromSep, toSep, -1))
	}
	return Replacer{r}
}

// NewUpperCaseReplacer creates a Replacer that forces keys to upper case.
func NewUpperCaseReplacer(fromSep, toSep string) Replacer {
	f := func(s string) string {
		return strings.ToUpper(strings.Replace(s, fromSep, toSep, -1))
	}
	return Replacer{f}
}

// NewLowerCamelCaseReplacer creates a Replacer that forces keys to camel case,
// i.e. each word begins with a capital letter, except the first word which
// is all lower case.
func NewLowerCamelCaseReplacer(fromSep, toSep string) Replacer {
	f := func(from string) string {
		path := strings.Split(from, fromSep)
		for idx, p := range path {
			if idx == 0 {
				path[idx] = strings.ToLower(p)
			} else {
				path[idx] = CamelWord(p)
			}
		}
		return strings.Join(path, toSep)
	}
	return Replacer{f}
}

// NewUpperCamelCaseReplacer creates a Replacer that forces keys to camel case,
// i.e. each word begins with a capital letter, including the first work.
func NewUpperCamelCaseReplacer(fromSep, toSep string) Replacer {
	f := func(from string) string {
		path := strings.Split(from, fromSep)
		for idx, p := range path {
			path[idx] = CamelWord(p)
		}
		return strings.Join(path, toSep)
	}
	return Replacer{f}
}

// CamelWord returns the CamelCase version of a word, i.e. the first
// letter capitalised and other characters lowercase.
func CamelWord(s string) string {
	r, n := utf8.DecodeRuneInString(s)
	return string(unicode.ToUpper(r)) + strings.ToLower(s[n:])
}
