// Copyright © 2017 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package flag provides a command line Getter using Go's flag.
//
package flag

import (
	"flag"
	"strings"

	"github.com/warthog618/config/keys"
	"github.com/warthog618/config/tree"
)

// New creates a new Getter.
//
// The getter will return the values
// By default the Getter will:
// - parse the flags provides by flag using flag.Visit
// - replace '-' in the flag space with '.' in the config space.
// - split list values with the ',' separator.
func New(options ...Option) (*Getter, error) {
	r := Getter{listSeparator: ","}
	for _, option := range options {
		option(&r)
	}
	if r.keyReplacer == nil {
		r.keyReplacer = keys.StringReplacer("-", ".")
	}
	if r.visit == nil {
		r.visit = flag.Visit
	}
	r.parse()
	return &r, nil
}

// Getter provides the mapping from flags to a config.Getter.
// The Getter scans the command line flags only at construction time, so its config state
// is effectively immutable.
type Getter struct {
	// The parsed config.
	config map[string]interface{}
	// A replacer that maps from flag space to config space.
	keyReplacer Replacer
	// The separator for slices stored in string values.
	listSeparator string
	// Visitor function to use to scan flags.
	visit func(func(*flag.Flag))
}

// Replacer maps a key from one space to another.
type Replacer interface {
	Replace(string) string
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
func WithKeyReplacer(keyReplacer Replacer) Option {
	return func(g *Getter) {
		g.keyReplacer = keyReplacer
	}
}

// WithListSeparator sets the separator between slice fields in the flag space.
// The default separator is ","
func WithListSeparator(separator string) Option {
	return func(g *Getter) {
		g.listSeparator = separator
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
		config[key] = splitList(f.Value.String(), g.listSeparator)
	})
	g.config = config
}

func splitList(v interface{}, l string) interface{} {
	if vstr, ok := v.(string); ok {
		if len(l) > 0 && strings.Contains(vstr, l) {
			return strings.Split(vstr, l)
		}
	}
	return v
}
