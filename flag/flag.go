// Copyright © 2017 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package flag provides a command line Getter using Go's flag.
//
package flag

import (
	"flag"

	"github.com/warthog618/config"
	"github.com/warthog618/config/keys"
	"github.com/warthog618/config/list"
	"github.com/warthog618/config/tree"
)

// New creates a new Getter.
//
// The getter will return the values derived from command line flags.
// By default the Getter will:
// - parse the flags provides by flag using flag.Visit
// - replace '-' in the flag space with '.' in the config space.
// - split list values with the ',' separator.
func New(options ...Option) *Getter {
	g := Getter{}
	for _, option := range options {
		option(&g)
	}
	if g.keyReplacer == nil {
		g.keyReplacer = keys.StringReplacer("-", ".")
	}
	if g.listSplitter == nil {
		g.listSplitter = list.NewSplitter(",")
	}
	if g.visit == nil {
		g.visit = flag.Visit
	}
	g.parse()
	return &g
}

// Getter provides the mapping from flags to a config.Getter.
// The Getter scans the command line flags only at construction time, so its config state
// is effectively immutable.
type Getter struct {
	config.GetterAsOption
	// The parsed config.
	config map[string]interface{}
	// A replacer that maps from flag space to config space.
	keyReplacer keys.Replacer
	// The splitter for slices stored in string values.
	listSplitter list.Splitter
	// Visitor function to use to scan flags.
	visit func(func(*flag.Flag))
}

// Option is a function which modifies a Getter at construction time.
type Option func(*Getter)

// WithAllFlags visits all the flags using flag.VisitAll, instead of only the
// set flags using flag.Visit.
func WithAllFlags() Option {
	return func(g *Getter) {
		g.visit = flag.VisitAll
	}
}

// WithKeyReplacer sets the replacer used to map from flag space to config space.
// The default replaces '-' in the flag space with '.' in the config space.
func WithKeyReplacer(keyReplacer keys.Replacer) Option {
	return func(g *Getter) {
		g.keyReplacer = keyReplacer
	}
}

// WithListSplitter splits slice fields stored as strings in the flag space.
// The default splitter separates on ",".
func WithListSplitter(splitter list.Splitter) Option {
	return func(g *Getter) {
		g.listSplitter = splitter
	}
}

// Get returns the value for a given key and true if found, or
// nil and false if not.
func (g *Getter) Get(key string) (interface{}, bool) {
	return tree.Get(g.config, key, "")
}

func (g *Getter) parse() {
	config := map[string]interface{}{}
	g.visit(func(f *flag.Flag) {
		key := g.keyReplacer.Replace(f.Name)
		config[key] = g.listSplitter.Split(f.Value.String())
	})
	g.config = config
}
