// Copyright © 2018 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package tree provides functions to get from common tree structures.
package tree

import (
	"strconv"
	"strings"
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
		if a, ok := isArrayLen(path[0]); ok {
			lenreq = true
			path[0] = a
		}
	}
	a, idx := parseArrayElement(path[0])
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
		if a, ok := isArrayLen(path[0]); ok {
			lenreq = true
			path[0] = a
		}
	}
	a, idx := parseArrayElement(path[0])
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

// isArrayLen determines if the key corresponds to an array length.
// i.e. is of the form a[].
// If so isArrayLen returns true and the name of the array.
func isArrayLen(key string) (string, bool) {
	if strings.HasSuffix(key, "[]") {
		return key[:len(key)-2], true
	}
	return key, false
}

// parseArrayElement determines if the key corresponds to an array element.
// i.e. is of the form a[i].
// The name of the array and the a list of indicies into the array.
func parseArrayElement(key string) (string, []int) {
	if !strings.HasSuffix(key, "]") {
		return key, nil
	}
	start := strings.Index(key, "[")
	if start == -1 {
		return key, nil
	}
	i := strings.Split(key[start+1:len(key)-1], "][")
	ii := make([]int, len(i))
	for i, is := range i {
		idx, err := strconv.Atoi(is)
		if err != nil {
			return key, nil
		}
		ii[i] = idx
	}
	return key[0:start], ii
}
