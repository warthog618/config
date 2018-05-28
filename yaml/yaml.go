// Copyright © 2017 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package yaml provides a YAML format getter for config.
package yaml

import (
	"fmt"
	"io/ioutil"
	"strings"

	"gopkg.in/yaml.v2"
)

// Getter provides the mapping from YAML to a config.Getter.
// The Getter parses the YAML only at construction time, so its config state
// is effectively immutable.
type Getter struct {
	config map[interface{}]interface{}
}

// Get returns the value for a given key and true if found, or
// nil and false if not.
func (r *Getter) Get(key string) (interface{}, bool) {
	if v, err := getFromMapTree(r.config, key, "."); err == nil {
		return v, true
	}
	return nil, false
}

// NewBytes returns a YAML Getter that reads config from a []byte.
func NewBytes(cfg []byte) (*Getter, error) {
	var config map[interface{}]interface{}
	err := yaml.Unmarshal(cfg, &config)
	if err != nil {
		return nil, err
	}
	return &Getter{config}, nil
}

// NewFile returns a YAML Getter that reads config from a named file.
func NewFile(filename string) (*Getter, error) {
	cfg, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
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
