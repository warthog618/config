// Copyright © 2017 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package env provides an environment variable Getter for config.
package env

import (
	"os"
	"strings"

	"github.com/warthog618/config/keys"
)

// New creates an environment variable Getter.
func New(options ...Option) (*Getter, error) {
	r := Getter{listSeparator: ":"}
	for _, option := range options {
		option(&r)
	}
	if r.keyReplacer == nil {
		r.keyReplacer = keys.ChainReplacer(
			keys.StringReplacer("_", "."),
			keys.LowerCaseReplacer())
	}
	r.load()
	return &r, nil
}

// Getter provides the mapping from environment variables to a config.Getter.
// The Getter scans the environment only at construction time, so its config state
// is effectively immutable.
type Getter struct {
	// config key=value
	config map[string]string
	// prefix in env space used to identify variables of interest.
	// This must include any separator.
	envPrefix string
	// A replacer that translates from env space to config space.
	// The replacement is applied AFTER the envPrefix has been removed.
	// e.g. environment var APP_MY_CONFIG with envPrefix "APP_"
	// would map from "MY_CONFIG".
	keyReplacer Replacer
	// The separator for slices stored in string values.
	listSeparator string
}

// Replacer maps a key from one space to another.
type Replacer interface {
	Replace(string) string
}

// Get returns the value for a given key and true if found, or
// nil and false if not.
func (r *Getter) Get(key string) (interface{}, bool) {
	if v, ok := r.config[key]; ok {
		if len(r.listSeparator) > 0 && strings.Contains(v, r.listSeparator) {
			return strings.Split(v, r.listSeparator), ok
		}
		return v, ok
	}
	if p, ok := keys.IsArrayLen(key); ok {
		if v, ok := r.config[p]; ok {
			return strings.Count(v, r.listSeparator) + 1, ok
		}
	}
	if p, i := keys.ParseArrayElement(key); len(i) == 1 {
		if v, ok := r.config[p]; ok {
			if len(r.listSeparator) > 0 && strings.Contains(v, r.listSeparator) {
				l := strings.Split(v, r.listSeparator)
				if i[0] < len(l) {
					return l[i[0]], true
				}
				return nil, false
			}
		}
	}
	return nil, false
}

// Option is a function which modifies a Getter at construction time.
type Option func(*Getter)

// WithEnvPrefix sets the prefix for environment variables included in this Getter's config.
// The prefix is stripped from the environment variable name during mapping to
// the config space and so should include any separator between it and the
// first tier name.
func WithEnvPrefix(prefix string) Option {
	return func(r *Getter) {
		r.envPrefix = prefix
	}
}

// WithKeyReplacer sets the replacer used to map from env space to config space.
// The default is to replace "_" with "." and convert to lowercase.
func WithKeyReplacer(m Replacer) Option {
	return func(r *Getter) {
		r.keyReplacer = m
	}
}

// WithListSeparator sets the separator between slice fields in the env space.
// The default separator is ":"
func WithListSeparator(separator string) Option {
	return func(r *Getter) {
		r.listSeparator = separator
	}
}

func (r *Getter) load() {
	config := map[string]string{}
	for _, env := range os.Environ() {
		if strings.HasPrefix(env, r.envPrefix) {
			keyValue := strings.SplitN(env, "=", 2)
			if len(keyValue) == 2 {
				envKey := keyValue[0][len(r.envPrefix):]
				cfgKey := r.keyReplacer.Replace(envKey)
				config[cfgKey] = keyValue[1]
			}
		}
	}
	r.config = config
}
