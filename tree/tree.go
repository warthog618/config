// Copyright © 2018 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package tree provides functions to get from common tree structures.
package tree

import (
	"reflect"
	"strings"

	"github.com/warthog618/config/keys"
)

// Get returns the element identified by key from configuration stored
// in a map[string]interface{} or map[interface{}]interface{} tree.
func Get(node interface{}, key string, pathSep string) (interface{}, bool) {
	switch nt := node.(type) {
	case map[interface{}]interface{}:
		return getFromFunc(func(k string) (interface{}, bool) {
			v, ok := nt[k]
			return v, ok
		}, key, pathSep)
	case map[string]interface{}:
		return getFromFunc(func(k string) (interface{}, bool) {
			v, ok := nt[k]
			return v, ok
		}, key, pathSep)
	default:
		return nil, false
	}
}

type getterFunc func(string) (interface{}, bool)

// getFromFunc gets from a tree structure with the provided getterFunc.
func getFromFunc(g getterFunc, key string, pathSep string) (interface{}, bool) {
	// full key match - also handles leaves
	if v, ok := g(key); ok {
		return getLeafElement(v)
	}
	lenreq := false
	var path []string
	if len(pathSep) == 0 {
		path = []string{key}
	} else {
		path = strings.SplitN(key, pathSep, 2)
	}
	if len(path) > 1 {
		// nested path match
		if v, ok := g(path[0]); ok {
			return Get(v, path[1], pathSep)
		}
	} else {
		if a, ok := keys.IsArrayLen(path[0]); ok {
			lenreq = true
			path[0] = a
		}
	}
	a, idx := keys.ParseArrayElement(path[0])
	if lenreq || idx != nil {
		if v, ok := g(a); ok {
			return getArrayElement(v, path, pathSep, idx, lenreq)
		}
	}
	// no match
	return nil, false
}

func getArrayElement(v interface{}, path []string, pathSep string, idx []int, lenreq bool) (interface{}, bool) {
	for _, i := range idx {
		vv := reflect.ValueOf(v)
		vk := vv.Kind()
		switch vk {
		case reflect.Array, reflect.Slice:
			if i >= vv.Len() {
				return nil, false
			}
			v = vv.Index(i).Interface()
		default:
			return nil, false
		}
	}
	switch vt := v.(type) {
	case map[interface{}]interface{}:
		if len(path) > 1 {
			return getFromFunc(func(k string) (interface{}, bool) {
				v, ok := vt[k]
				return v, ok
			}, path[1], pathSep)
		}
		return nil, false
	case map[string]interface{}:
		if len(path) > 1 {
			return getFromFunc(func(k string) (interface{}, bool) {
				v, ok := vt[k]
				return v, ok
			}, path[1], pathSep)
		}
		return nil, false
	default:
		// handle arrays of all types
		vv := reflect.ValueOf(v)
		vk := vv.Kind()
		switch vk {
		case reflect.Array, reflect.Slice:
			if lenreq {
				return vv.Len(), true
			}
			if vv.Len() > 0 {
				if vv.Type().Elem().Kind() == reflect.Map {
					return make([]interface{}, vv.Len()), true
				}
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
