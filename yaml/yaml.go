// Copyright © 2017 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package yaml provides a YAML format reader for config.
package yaml

import (
	"fmt"
	"io/ioutil"
	"strings"

	"gopkg.in/yaml.v2"
)

// Reader provides the mapping from JSON to a config.Reader.
type Reader struct {
	config map[interface{}]interface{}
}

// Read returns the value for a given key and true if found, or
// nil and false if not.
func (r *Reader) Read(key string) (interface{}, bool) {
	if v, err := getFromMapTree(r.config, key, "."); err == nil {
		return v, true
	}
	return nil, false
}

// NewBytes returns a YAML reader that reads config from a []byte.
func NewBytes(cfg []byte) (*Reader, error) {
	var config map[interface{}]interface{}
	err := yaml.Unmarshal(cfg, &config)
	if err != nil {
		return &Reader{}, err
	}
	return &Reader{config}, nil
}

// NewFile returns a YAML reader that reads config from a named file.
func NewFile(filename string) (*Reader, error) {
	cfg, err := ioutil.ReadFile(filename)
	if err != nil {
		return &Reader{}, err
	}
	return NewBytes(cfg)
}

func getFromMapTree(node map[interface{}]interface{}, key string, pathSep string) (interface{}, error) {
	// full key match - also handles leaves
	if v, ok := node[key]; ok {
		if _, ok := v.(map[interface{}]interface{}); !ok {
			return v, nil
		}
	}
	// nested path match
	path := strings.Split(key, pathSep)
	if v, ok := node[path[0]]; ok {
		switch vt := v.(type) {
		case map[interface{}]interface{}:
			return getFromMapTree(vt, strings.Join(path[1:], pathSep), pathSep)
		}
	}
	// no match
	return nil, fmt.Errorf("key '%v' not found", key)
}
