// Copyright © 2017 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package json provides a JSON format Getter for config.
package json

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
)

// Getter provides the mapping from JSON to a config.Getter.
// The Getter parses the JSON only at construction time, so its config state
// is effectively immutable.
type Getter struct {
	config map[string]interface{}
}

// Get returns the value for a given key and true if found, or
// nil and false if not.
func (r *Getter) Get(key string) (interface{}, bool) {
	if v, err := getFromMapTree(r.config, key, "."); err == nil {
		return v, true
	}
	return nil, false
}

// NewBytes returns a JSON Getter that reads config from a []byte.
func NewBytes(cfg []byte) (*Getter, error) {
	var config map[string]interface{}
	err := json.Unmarshal(cfg, &config)
	if err != nil {
		return nil, err
	}
	return &Getter{config}, nil
}

// NewFile returns a JSON Getter that reads config from a named file.
func NewFile(filename string) (*Getter, error) {
	cfg, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return NewBytes(cfg)
}

func getFromMapTree(node map[string]interface{}, key string, pathSep string) (interface{}, error) {
	// full key match - also handles leaves
	if v, ok := node[key]; ok {
		if _, ok := v.(map[string]interface{}); !ok {
			return v, nil
		}
	}
	// nested path match
	path := strings.Split(key, pathSep)
	if v, ok := node[path[0]]; ok {
		switch vt := v.(type) {
		case map[string]interface{}:
			return getFromMapTree(vt, strings.Join(path[1:], pathSep), pathSep)
		}
	}
	// no match
	return nil, fmt.Errorf("key '%v' not found", key)
}
