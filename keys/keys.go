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

// ReplacerFunc is a func that implements a string replacement.
type ReplacerFunc func(key string) string

// Replace calls the ReplacerFunc to perform the map.
func (r ReplacerFunc) Replace(key string) string {
	return r(key)
}

// ChainReplacer returns a replacer that applies a list of replacers, in order.
func ChainReplacer(rr ...ReplacerFunc) ReplacerFunc {
	return func(key string) string {
		for _, r := range rr {
			if r != nil {
				key = r.Replace(key)
			}
		}
		return key
	}
}

// CamelCaseReplacer is a replacer that that forces keys to camel case,
// so each word begins with a capital letter.
// Words are separated by the default separator - ".".
func CamelCaseReplacer() ReplacerFunc {
	return CamelCaseSepReplacer(".")
}

// CamelCaseSepReplacer is a replacer that that forces keys to camel case,
// so each word begins with a capital letter.
// Words are separated by the provided separator.
func CamelCaseSepReplacer(sep string) ReplacerFunc {
	return func(key string) string {
		if key == "" {
			return ""
		}
		path := strings.Split(key, sep)
		for i, p := range path {
			path[i] = camelCase(p)
		}
		return strings.Join(path, sep)
	}
}

// LowerCamelCaseReplacer is a replacer that that forces keys to camel case,
// so each word begins with a capital letter, except the first word which
// is all lower case.
// Words are separated by the default separator - ".".
func LowerCamelCaseReplacer() ReplacerFunc {
	return LowerCamelCaseSepReplacer(".")
}

// LowerCamelCaseSepReplacer is a replacer that that forces keys to camel case,
// so each word begins with a capital letter, except the first word which
// is all lower case.
// Words are separated by the provided separator.
func LowerCamelCaseSepReplacer(sep string) ReplacerFunc {
	return func(key string) string {
		if key == "" {
			return ""
		}
		path := strings.Split(key, sep)
		path[0] = strings.ToLower(path[0])
		for i, p := range path[1:] {
			path[i+1] = camelCase(p)
		}
		return strings.Join(path, sep)
	}
}

// LowerCaseReplacer is a replacer that that forces keys to lower case.
func LowerCaseReplacer() ReplacerFunc {
	return func(key string) string {
		return strings.ToLower(key)
	}
}

// NullReplacer is a replacer that leaves a key unchanged.
// This can be used to override the mapping in Getters that assume
// a default mapping if none (Replacer(nil)) is provided.
func NullReplacer() ReplacerFunc {
	return func(key string) string {
		return key
	}
}

// PrefixReplacer adds a prefix to keys.
// This can be used to logically move the root of a Getter to a
// node of the config space.
func PrefixReplacer(prefix string) ReplacerFunc {
	return func(key string) string {
		return prefix + key
	}
}

// StringReplacer replaces one string in the key with another.
// The ols is replaced with the new using strings.Replace.
// This is typically used to replace tier separators,
// e.g. "." in config space with "_" in env space,
// but can also be used for arbitrary substitutions.
func StringReplacer(old, new string) ReplacerFunc {
	return func(key string) string {
		return strings.Replace(key, old, new, -1)
	}
}

// UpperCaseReplacer forces keys to upper case.
func UpperCaseReplacer() ReplacerFunc {
	return func(key string) string {
		return strings.ToUpper(key)
	}
}

// camelCase returns the CamelCase version of a string, i.e. the first
// letter capitalised and other characters lowercase.
func camelCase(key string) string {
	r, n := utf8.DecodeRuneInString(key)
	return string(unicode.ToUpper(r)) + strings.ToLower(key[n:])
}
