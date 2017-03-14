// Copyright © 2017 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package env provides an environment variable reader for config.
package env

import (
	"os"
	"strings"
)

// New creates an enviroment variable reader.
//
// The prefix determines the set of environment variables of interest to this reader.
// Environment variables beginning with the prefix are loaded into the config.
// The mapping from environment variable naming to config space naming is
// determined by the prefix and separator fields of the Reader.
func New(prefix string) (*Reader, error) {
	config := map[string]string(nil)
	nodes := map[string]bool(nil)
	r := Reader{config, nodes, prefix, "_", ".", ":"}
	r.load()
	return &r, nil
}

// Reader provides the mapping from environment variables to a config.Reader.
type Reader struct {
	// config key=value
	config map[string]string
	// set of nodes contained in config
	nodes map[string]bool
	// prefix in ENV space.
	// This must include any separator - the envSeparator does not separate the
	// prefix from the remainder of the key.
	envPrefix string
	// separator between key tiers in ENV space.
	envSeparator string
	// separator between key tiers in config space.
	cfgSeparator string
	// The separator for slices stored in string values.
	listSeparator string
}

// Contains returns true if the Reader contains a value corresponding to the
// provided key.  For node keys it returns true if there is at least one value
// available within that node's config tree.
func (r *Reader) Contains(key string) bool {
	if _, ok := r.config[key]; ok {
		return true
	}
	if _, ok := r.nodes[key]; ok {
		return true
	}
	return false
}

// Read returns the value for a given key and true if found, or
// nil and false if not.
func (r *Reader) Read(key string) (interface{}, bool) {
	v, ok := r.config[key]
	if ok && len(r.listSeparator) > 0 && strings.Contains(v, r.listSeparator) {
		return strings.Split(v, r.listSeparator), ok
	}
	return v, ok
}

// SetCfgSeparator sets the separator between tiers in the config namespace.
// The default separator is "."
func (r *Reader) SetCfgSeparator(separator string) {
	r.cfgSeparator = strings.ToLower(separator)
	r.load()
}

// SetEnvPrefix sets the prefix for environment variables included in this reader's config.
// The prefix is stripped from the environment variable name during mapping to
// the config namespace and so should include any separator between it and the
// first tier name.
func (r *Reader) SetEnvPrefix(prefix string) {
	r.envPrefix = prefix
	r.load()
}

// SetEnvSeparator sets the separator between tiers in the env namespace.
// The default separator is "_"
func (r *Reader) SetEnvSeparator(separator string) {
	r.envSeparator = separator
	r.load()
}

// SetListSeparator sets the separator between slice fields in the env namespace.
// The default separator is ":"
func (r *Reader) SetListSeparator(separator string) {
	r.listSeparator = separator
}

func (r *Reader) load() {
	config := map[string]string{}
	nodes := map[string]bool{}
	for _, env := range os.Environ() {
		if strings.HasPrefix(env, r.envPrefix) {
			keyValue := strings.SplitN(env, "=", 2)
			if len(keyValue) == 2 {
				envKey := keyValue[0][len(r.envPrefix):]
				path := strings.Split(envKey, r.envSeparator)
				cfgKey := strings.ToLower(strings.Join(path, r.cfgSeparator))
				config[cfgKey] = keyValue[1]
				nodePath := path[0]
				for idx := 1; idx < len(path); idx++ {
					nodes[strings.ToLower(nodePath)] = true
					nodePath = nodePath + r.cfgSeparator + path[idx]
				}
			}
		}
	}
	r.config, r.nodes = config, nodes
}
