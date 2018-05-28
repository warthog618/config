// Copyright © 2017 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package prefixer provides a Getter decorator that relocates the
// root of the Getter's keys within the config namespace.
package prefixer

import (
	"strings"
)

// Getter is the same as config.Getter.
// Redfined to avoid dependency.
type Getter interface {
	Get(key string) (interface{}, bool)
}

// A wrapper around Getter that relocates the Getter config
// to a subtree of the full tree.
type prefixer struct {
	// The Getter.
	Getter
	// The prefix of the Getter config within the config tree.
	// This is typically a config node, plus trailing separator.
	prefix string
}

// New returns a new prefixer decorating the provided Getter.
// The prefix defines the root node for the config returned by the Getter.
// e.g. with a prefix "module", reading the key "module.field" from the
// prefixer will return the "field" Getter from the Getter.
func New(prefix string, g Getter) Getter {
	return &prefixer{g, prefix}
}

func (p *prefixer) Get(key string) (interface{}, bool) {
	if !strings.HasPrefix(key, p.prefix) {
		return nil, false
	}
	key = key[len(p.prefix):]
	return p.Getter.Get(key)
}
