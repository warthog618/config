package prefixer

import (
	"github.com/warthog618/config"
	"strings"
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

func New(prefix string, reader config.Reader) config.Reader {
	return &prefixer{prefix, reader}
}

func (p *prefixer) Contains(key string) bool {
	if !strings.HasPrefix(key, p.prefix) {
		return false
	}
	key = key[len(p.prefix):]
	return p.reader.Contains(key)
}

func (p *prefixer) Read(key string) (interface{}, bool) {
	if !strings.HasPrefix(key, p.prefix) {
		return nil, false
	}
	key = key[len(p.prefix):]
	return p.reader.Read(key)
}
