// Copyright © 2017 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package prefixer provides a Reader decorator that relocates the
// root of the reader's keys within the config namespace.
package prefixer

import (
	"strings"

	"github.com/warthog618/config"
)

type reader config.Reader

// A wrapper around Reader that relocates the Reader config
// to a subtree of the full tree.
type prefixer struct {
	// The prefix of the reader config within the config tree.
	// This is typically a config node, plus separator.
	prefix string
	// The reader.
	reader
}

// New returns a new prefixer decorating the provided Reader.
// The prefix defines the root node for the config returned by the reader.
// e.g. with a prefix "module", reading the key "module.field" from the
// prefixer will return the "field" reader from the Reader.
func New(prefix string, reader config.Reader) config.Reader {
	return &prefixer{prefix, reader}
}

func (p *prefixer) Read(key string) (interface{}, bool) {
	if !strings.HasPrefix(key, p.prefix) {
		return nil, false
	}
	key = key[len(p.prefix):]
	return p.reader.Read(key)
}
