package masker

import (
	"github.com/warthog618/config"
)

// Unconditional masker
// Provide set of nodes and leaves which are masked.

// Containment masker
// Provide set of nodes and leaves which are masked.
// But Mask only active if reader.Contains(key)

type reader config.Reader

type masker struct {
	// conditional on the reader containing
	conditional bool
	// set of keys (node or leaf) masked.
	mask map[string]bool
	// the reader this mask applies to.
	reader
}

func New(reader config.Reader, conditional bool) *masker {
	return &masker{
		reader:      reader,
		conditional: conditional,
		mask:        map[string]bool{},
	}
}

func (m *masker) AddMask(key string) {
	m.mask[key] = true
}

func (m *masker) Mask(key string) bool {
	if _, ok := m.mask[key]; !ok {
		return false
	}
	if !m.conditional {
		return true
	}
	return m.Contains(key)
}
