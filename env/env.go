// Copyright © 2017 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package env provides an environment variable reader for config.
package env

import (
	"os"
	"strings"

	"github.com/warthog618/config/keys"
)

// New creates an environment variable reader.
//
// The prefix determines the set of environment variables of interest to this reader.
// Environment variables beginning with the prefix are loaded into the config.
// The mapping from environment variable naming to config space naming is
// determined by the prefix and separator fields of the Reader.
func New(prefix string, options ...Option) (*Reader, error) {
	config := map[string]string(nil)
	r := Reader{config, prefix, nil, ":"}
	for _, option := range options {
		option(&r)
	}
	if r.cfgKeyReplacer == nil {
		r.cfgKeyReplacer = keys.NewLowerCaseReplacer("_", ".")
	}
	r.load()
	return &r, nil
}

// Reader provides the mapping from environment variables to a config.Reader.
// The Reader scans the envrionment only at construction time, so its config state
// is effectively immutable.
type Reader struct {
	// config key=value
	config map[string]string
	// prefix in ENV space.
	// This must include any separator - the envSeparator does not separate the
	// prefix from the remainder of the key.
	envPrefix string
	// A replacer that maps from env space to config space.
	// The replacer is applied AFTER the prefix has been removed.
	cfgKeyReplacer Replacer
	// The separator for slices stored in string values.
	listSeparator string
}

// Read returns the value for a given key and true if found, or
// nil and false if not.
func (r *Reader) Read(key string) (interface{}, bool) {
	if v, ok := r.config[key]; ok {
		if len(r.listSeparator) > 0 && strings.Contains(v, r.listSeparator) {
			return strings.Split(v, r.listSeparator), ok
		}
		return v, ok
	}
	return nil, false
}

// Replacer is a string replacer similar to strings.Replacer.
// It must be safe for use by multiple goroutines.
type Replacer interface {
	Replace(s string) string
}

// Option is a function which modifies a Reader at construction time.
type Option func(*Reader)

// WithEnvPrefix sets the prefix for environment variables included in this reader's config.
// The prefix is stripped from the environment variable name during mapping to
// the config namespace and so should include any separator between it and the
// first tier name.
func WithEnvPrefix(prefix string) Option {
	return func(r *Reader) {
		r.envPrefix = prefix
	}
}

// WithCfgKeyReplacer sets the replacer used to map from env space to config space.
// The default is to replace "_" with "." and convert to lowercase.
func WithCfgKeyReplacer(keyReplacer keys.Replacer) Option {
	return func(r *Reader) {
		r.cfgKeyReplacer = keyReplacer
	}
}

// WithListSeparator sets the separator between slice fields in the env namespace.
// The default separator is ":"
func WithListSeparator(separator string) Option {
	return func(r *Reader) {
		r.listSeparator = separator
	}
}

func (r *Reader) load() {
	config := map[string]string{}
	for _, env := range os.Environ() {
		if strings.HasPrefix(env, r.envPrefix) {
			keyValue := strings.SplitN(env, "=", 2)
			if len(keyValue) == 2 {
				envKey := keyValue[0][len(r.envPrefix):]
				cfgKey := r.cfgKeyReplacer.Replace(envKey)
				config[cfgKey] = keyValue[1]
			}
		}
	}
	r.config = config
}
