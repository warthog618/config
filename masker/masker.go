// Copyright © 2017 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package masker provides a Reader decorator that can mask values stored
// in readers that are lower in the config stack.
// It does not mask the values of the Reader it decorates.
package masker

import (
	"github.com/warthog618/config"
)

type reader config.Reader

// Masker is a reader decorator that implements the config.Masker interface.
type Masker struct {
	// conditional on the reader containing
	conditional bool
	// set of keys (node or leaf) masked.
	mask map[string]bool
	// the reader this mask applies to.
	reader
}

// New returns a new Masker decorating the provided Reader.
// If conditional is set then the masker only masks nodes if the reader
// contains config values in the node's subtree.
// If unconditional then the masker masks independent of the values in the reader.
func New(reader config.Reader, conditional bool) *Masker {
	return &Masker{
		reader:      reader,
		conditional: conditional,
		mask:        map[string]bool{},
	}
}

// AddMask adds a mask to the Masker.
// For a leaf this masks only that key.
// For a node this masks the node's subtree.
func (m *Masker) AddMask(key string) {
	m.mask[key] = true
}

// Mask returns true if the key is masked by the Masker.
func (m *Masker) Mask(key string) bool {
	if _, ok := m.mask[key]; !ok {
		return false
	}
	if !m.conditional {
		return true
	}
	return m.Contains(key)
}
