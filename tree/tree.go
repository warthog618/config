// Copyright © 2018 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package tree provides functions to get from common tree structures.
package tree

import (
	"strings"

	"github.com/warthog618/config/keys"
)

// Get returns the element identified by key from configuration stored
// in a map[string]interface{} or map[interface{}]interface{} tree.
func Get(node interface{}, key string, pathSep string) (interface{}, bool) {
	switch nt := node.(type) {
	case map[string]interface{}:
		return getFromMSI(nt, key, pathSep)
	case map[interface{}]interface{}:
		return getFromMII(nt, key, pathSep)
	default:
		return nil, false
	}
}

// getFromMII performs a get from a config tree structures as a map[interface{}]interface{},
// as returned by yaml.Unmarshal.
func getFromMII(node map[interface{}]interface{}, key string, pathSep string) (interface{}, bool) {
	// full key match - also handles leaves
	if v, ok := node[key]; ok {
		return getLeafElement(v)
	}
	lenreq := false
	path := strings.SplitN(key, pathSep, 2)
	if len(path) > 1 {
		// nested path match
		if v, ok := node[path[0]]; ok {
			return getNestedElement(v, path[1], pathSep)
		}
	} else {
		if a, ok := keys.IsArrayLen(path[0]); ok {
			lenreq = true
			path[0] = a
		}
	}
	a, idx := keys.ParseArrayElement(path[0])
	if lenreq || idx != nil {
		if v, ok := node[a]; ok {
			return getArrayElement(v, path, pathSep, idx, lenreq)
		}
	}
	// no match
	return nil, false
}

// getFromMSI performs a get from a config tree structures as a map[string]interface{},
// as returned by json.Unmarshal.
func getFromMSI(node map[string]interface{}, key string, pathSep string) (interface{}, bool) {
	// full key match - also handles leaves
	if v, ok := node[key]; ok {
		return getLeafElement(v)
	}
	lenreq := false
	path := strings.SplitN(key, pathSep, 2)
	if len(path) > 1 {
		// nested path match
		if v, ok := node[path[0]]; ok {
			return getNestedElement(v, path[1], pathSep)
		}
	} else {
		if a, ok := keys.IsArrayLen(path[0]); ok {
			lenreq = true
			path[0] = a
		}
	}
	a, idx := keys.ParseArrayElement(path[0])
	if lenreq || idx != nil {
		if v, ok := node[a]; ok {
			return getArrayElement(v, path, pathSep, idx, lenreq)
		}
	}
	// no match
	return nil, false
}

func getArrayElement(v interface{}, path []string, pathSep string, idx []int, lenreq bool) (interface{}, bool) {
	for _, i := range idx {
		if av, ok := v.([]interface{}); ok {
			if i < len(av) {
				v = av[i]
				continue
			}
		}
		return nil, false
	}
	switch vt := v.(type) {
	case map[interface{}]interface{}:
		if len(path) > 1 {
			return getFromMII(vt, path[1], pathSep)
		}
		return nil, false
	case map[string]interface{}:
		if len(path) > 1 {
			return getFromMSI(vt, path[1], pathSep)
		}
		return nil, false
	case []interface{}:
		if lenreq {
			return len(vt), true
		}
		if len(vt) > 0 {
			switch vt[0].(type) {
			case map[interface{}]interface{}, map[string]interface{}:
				return make([]interface{}, len(vt)), true
			}
		}
	}
	return v, true
}

func getLeafElement(v interface{}) (interface{}, bool) {
	switch vt := v.(type) {
	case map[string]interface{}, map[interface{}]interface{}:
		return nil, false
	case []interface{}:
		if len(vt) > 0 {
			switch vt[0].(type) {
			case map[interface{}]interface{}, map[string]interface{}:
				return make([]interface{}, len(vt)), true
			}
		}
	}
	return v, true
}

func getNestedElement(v interface{}, key string, pathSep string) (interface{}, bool) {
	switch vt := v.(type) {
	case map[string]interface{}:
		return getFromMSI(vt, key, pathSep)
	case map[interface{}]interface{}:
		return getFromMII(vt, key, pathSep)
	default:
		return v, true
	}
}
