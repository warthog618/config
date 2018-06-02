// Copyright © 2017 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package config

import (
	"strings"
)

// Getter provides the minimal interface for a configuration Getter.
type Getter interface {
	// Get the value of the named config leaf key.
	// Also returns an ok, similar to a map read, to indicate if the value
	// was found.
	// The type underlying the returned interface{} must be convertable to
	// the expected type by cfgconv.
	// Get is not expected to be performed on node keys, but in case it is
	// the Get should return a nil interface{} and false, even if the node
	// exists in the config tree.
	// Must be safe to call from multiple goroutines.
	Get(key string) (interface{}, bool)
}

// A wrapper around Getter that relocates the Getter config
// to a subtree of the full config tree.
type mapped struct {
	// The wrapped Getter.
	Getter
	// The mapper applied to the key before being passed to the Getter.
	Mapper
}

// Mapper maps a key from one space to another.
type Mapper interface {
	Map(key string) string
}

// MappedGetter returns a getter decorating the wrapped Getter.
// The mapper performs key mapping from config space to getter space.
func MappedGetter(mapper Mapper, g Getter) Getter {
	return &mapped{g, mapper}
}

func (m *mapped) Get(key string) (interface{}, bool) {
	return m.Getter.Get(m.Map(key))
}

// A wrapper around Getter that relocates the Getter config
// to a subtree of the full config tree.
type prefixed struct {
	// The Getter.
	Getter
	// The prefix of the Getter config within the config tree.
	// This is typically a config node, plus trailing separator.
	prefix string
}

// PrefixedGetter returns a new getter decorating the wrapped Getter.
// The prefix defines the root node for the config returned by the wrapped Getter.
// e.g. with a prefix "module", reading the key "module.field" from the
// prefixed will return the "field" Getter from the Getter.
func PrefixedGetter(prefix string, g Getter) Getter {
	return &prefixed{g, prefix}
}

func (p *prefixed) Get(key string) (interface{}, bool) {
	if !strings.HasPrefix(key, p.prefix) {
		return nil, false
	}
	key = key[len(p.prefix):]
	return p.Getter.Get(key)
}
