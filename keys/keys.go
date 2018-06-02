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

// CamelCaseMapper is a mapper that that forces keys to camel case,
// so each word begins with a capital letter.
type CamelCaseMapper struct {
	// Sep is the string separating words - usually "." in config space.
	Sep string
}

// Map converts the string to camel case.
// e.g. CONFIG.FILE becomes Config.File
func (m CamelCaseMapper) Map(key string) string {
	path := strings.Split(key, m.Sep)
	for idx, p := range path {
		path[idx] = CamelCase(p)
	}
	return strings.Join(path, m.Sep)
}

// LowerCamelCaseMapper is a mapper that that forces keys to camel case,
// so each word begins with a capital letter, except the first word which
// is all lower case.
type LowerCamelCaseMapper struct {
	// Sep is the string separating words - usually "." in config space.
	Sep string
}

// Map converts the string to lower camel case.
// e.g. CONFIG.FILE becomes config.File
func (m LowerCamelCaseMapper) Map(key string) string {
	path := strings.Split(key, m.Sep)
	for idx, p := range path {
		if idx == 0 {
			path[idx] = strings.ToLower(p)
		} else {
			path[idx] = CamelCase(p)
		}
	}
	return strings.Join(path, m.Sep)
}

// LowerCaseMapper is a mapper that that forces keys to lower case.
type LowerCaseMapper struct{}

// Map converts the string to lower case.
// e.g. CONFIG.FILE becomes config.file
func (m LowerCaseMapper) Map(key string) string {
	return strings.ToLower(key)
}

// MultiMapper applied mupltiple maps to a key.
// The mappings are applied in the order they are listed in MM.
type MultiMapper struct {
	// MM is the set of maps to be applied.
	MM []Mapper
}

// Mapper maps a key from one space to another.
type Mapper interface {
	Map(key string) string
}

// Map applies the maps in the order they are listed in MM.
func (m MultiMapper) Map(key string) string {
	for _, mapper := range m.MM {
		key = mapper.Map(key)
	}
	return key
}

// NullMapper leaves a key unchanged.
// This can be used to override the mapping in Getters that assume
// a default mapping if none (Mapper(nil)) is provided.
type NullMapper struct{}

// Map simply returns the key unchanged.
func (m NullMapper) Map(key string) string {
	return key
}

// PrefixMapper adds a prefix to keys.
// This can be used to logically move the root of a Getter to a
// node of the config space.
type PrefixMapper struct {
	Prefix string
}

// Map returns the key with the prefix prepended.
func (m PrefixMapper) Map(key string) string {
	return m.Prefix + key
}

// ReplaceMapper replaces one string in the key with another.
// The From is replaced with the To.
// This is typically used to replace tier separators,
// e.g. "." in config space with "_" in env space,
// but can also be used for arbitrary substitutions.
type ReplaceMapper struct {
	From string
	To   string
}

// Map performs a replacement of instances of From in the key with To.
func (m ReplaceMapper) Map(key string) string {
	return strings.Replace(key, m.From, m.To, -1)
}

// UpperCaseMapper forces keys to upper case.
type UpperCaseMapper struct{}

// Map returns the upper case of the key.
func (m UpperCaseMapper) Map(key string) string {
	return strings.ToUpper(key)
}

// CamelCase returns the CamelCase version of a string, i.e. the first
// letter capitalised and other characters lowercase.
func CamelCase(key string) string {
	r, n := utf8.DecodeRuneInString(key)
	return string(unicode.ToUpper(r)) + strings.ToLower(key[n:])
}
