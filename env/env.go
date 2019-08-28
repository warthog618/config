// Copyright © 2017 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package env provides an environment variable Getter for config.
package env

import (
	"os"
	"strings"

	"github.com/warthog618/config"
	"github.com/warthog618/config/keys"
	"github.com/warthog618/config/list"
	"github.com/warthog618/config/tree"
)

// New creates an environment variable Getter.
func New(options ...Option) *Getter {
	g := Getter{}
	for _, option := range options {
		option(&g)
	}
	if g.keyReplacer == nil {
		g.keyReplacer = keys.ChainReplacer(
			keys.StringReplacer("_", "."),
			keys.LowerCaseReplacer())
	}
	if g.listSplitter == nil {
		g.listSplitter = list.NewSplitter(":")
	}
	g.load()
	return &g
}

// Getter provides the mapping from environment variables to a config.Getter.
// The Getter scans the environment only at construction time, so its config state
// is effectively immutable.
type Getter struct {
	config.GetterAsOption
	// config key=value
	config map[string]interface{}
	// prefix in env space used to identify variables of interest.
	// This must include any separator.
	envPrefix string
	// A replacer that translates from env space to config space.
	// The replacement is applied AFTER the envPrefix has been removed.
	// e.g. environment var APP_MY_CONFIG with envPrefix "APP_"
	// would map from "MY_CONFIG".
	keyReplacer keys.Replacer
	// The splitter for slices stored in string values.
	listSplitter list.Splitter
}

// Get returns the value for a given key and true if found, or
// nil and false if not.
func (g *Getter) Get(key string) (interface{}, bool) {
	return tree.Get(g.config, key, "")
}

// Option is a function which modifies a Getter at construction time.
type Option func(*Getter)

// WithEnvPrefix sets the prefix for environment variables included in this Getter's config.
// The prefix is stripped from the environment variable name during mapping to
// the config space and so should include any separator between it and the
// first tier name.
func WithEnvPrefix(prefix string) Option {
	return func(g *Getter) {
		g.envPrefix = prefix
	}
}

// WithKeyReplacer sets the replacer used to map from env space to config space.
// The default is to replace "_" with "." and convert to lowercase.
func WithKeyReplacer(m keys.Replacer) Option {
	return func(g *Getter) {
		g.keyReplacer = m
	}
}

// WithListSplitter splits slice fields stored as strings in the env space.
// The default splitter separates on ",".
func WithListSplitter(splitter list.Splitter) Option {
	return func(g *Getter) {
		g.listSplitter = splitter
	}
}

func (g *Getter) load() {
	config := map[string]interface{}{}
	for _, env := range os.Environ() {
		if strings.HasPrefix(env, g.envPrefix) {
			keyValue := strings.SplitN(env, "=", 2)
			if len(keyValue) == 2 {
				envKey := keyValue[0][len(g.envPrefix):]
				cfgKey := g.keyReplacer.Replace(envKey)
				config[cfgKey] = g.listSplitter.Split(keyValue[1])
			}
		}
	}
	g.config = config
}
