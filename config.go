// Copyright © 2017 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package config

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/warthog618/config/cfgconv"
)

// Config defines the functions available to access the config.
type Config interface {
	AddAlias(newKey string, oldKey string)
	AppendReader(reader Reader)
	InsertReader(reader Reader)
	SetSeparator(separator string)
	// Base get
	Get(key string) (interface{}, error)
	// Type gets
	GetBool(key string) (bool, error)
	GetFloat(key string) (float64, error)
	GetInt(key string) (int64, error)
	GetString(key string) (string, error)
	GetUint(key string) (uint64, error)
	// Slice gets
	GetSlice(key string) ([]interface{}, error)
	GetIntSlice(key string) ([]int64, error)
	GetStringSlice(key string) ([]string, error)
	GetUintSlice(key string) ([]uint64, error)
	// Tree gets
	GetConfig(key string) (Config, error)
	Unmarshal(key string, obj interface{}) error
	UnmarshalToMap(key string, objmap map[string]interface{}) error
}

// New creates a new config with no initial state.
func New() Config {
	return &config{
		separator: ".",
		readers:   make([]Reader, 0),
		aliases:   make(map[string][]string),
	}
}

// Reader provides the minimal interface for a configuration reader.
type Reader interface {
	// Read and return the value of the named config leaf key.
	// Also returns an ok, similar to a map read, to indicate if the value
	// was found.
	// The type underlying the returned interface{} must be convertable to
	// the expected type by cfgconv.
	// Read is not expected to be performed on node keys, and its behaviour
	// in that case is not defined.
	Read(key string) (interface{}, bool)
}

type config struct {
	// RWLock covering other fields.
	// It does not prevent concurrent access to the readers themselves,
	// only to the config fields.
	mu sync.RWMutex
	// The prefix common to all keys within this config node.
	// For the root node this is empty.
	// For other nodes this indicates the location of the node in the config tree.
	// The keys passed into the non-root nodes should be local to that node,
	// i.e. they should treat the node as the root of their own config tree.
	prefix string
	// path separator for nested objects
	separator string
	// A list of Readers providing config key/value pairs.
	readers []Reader
	// A map to a list of old names for current config keys.
	aliases map[string][]string
}

// Append a reader to the set of readers for the config node.
// This means this reader is only used as a last resort, relative to
// the existing readers.
//
// This is generally applied to the root node.
// When applied to a non-root node, the reader only applies to that node,
// and any subsequently created children.
func (c *config) AppendReader(reader Reader) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.readers = append(c.readers, reader)
}

// Insert a reader to the set of readers for the config node.
// This means this reader is used before the existing readers.
//
// This is generally applied to the root node.
// When applied to a non-root node, the reader only applies to that node,
// and any subsequently created children.
func (c *config) InsertReader(reader Reader) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.readers = append([]Reader{reader}, c.readers...)
}

// Return the absolute config key for a key relative to the config node.
func (c *config) prefixedKey(key ...string) string {
	rhs := strings.Join(key, c.separator)
	if len(c.prefix) == 0 {
		// we are root.
		return rhs
	}
	return c.prefix + c.separator + rhs
}

// Add an alias from a newKey, which should be used by the code,
// to an old key which may still be present in legacy config.
// As with readers, the aliases are local to the config node, and any
// subsequently created children.
func (c *config) AddAlias(newKey string, oldKey string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	lNewKey := c.prefixedKey(strings.ToLower(newKey))
	lOldKey := c.prefixedKey(strings.ToLower(oldKey))
	if lNewKey == lOldKey {
		return
	}
	aliases, ok := c.aliases[lNewKey]
	if !ok {
		aliases = make([]string, 1)
	}
	// prepended so alias are searched in LIFO order.
	c.aliases[lNewKey] = append([]string{lOldKey}, aliases...)
}

func (c *config) SetSeparator(separator string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.separator = separator
}

// Get the raw string value corresponding to the key.
// Iterates through the list of readers, searching for a matching key,
// or matching alias.  Returns the first match found, or an error if none is found.
func (c *config) Get(key string) (interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	fullKey := c.prefixedKey(strings.ToLower(key))
	for _, reader := range c.readers {
		if v, ok := reader.Read(fullKey); ok {
			return v, nil
		}
		if len(c.aliases) > 0 {
			// leaf alias
			if aliases, ok := c.aliases[fullKey]; ok {
				for _, alias := range aliases {
					if v, ok := reader.Read(alias); ok {
						return v, nil
					}
				}
			}
			// node alias
			path := strings.Split(fullKey, c.separator)
			for plen := len(path) - 1; plen >= 0; plen-- {
				nodeKey := strings.Join(path[:plen], c.separator)
				if aliases, ok := c.aliases[nodeKey]; ok {
					for _, alias := range aliases {
						if len(alias) > 0 {
							alias = alias + c.separator
						}
						idx := len(nodeKey)
						if idx > 0 {
							idx += len(c.separator)
						}
						aliasKey := alias + fullKey[idx:]
						if v, ok := reader.Read(aliasKey); ok {
							return v, nil
						}
					}
				}
			}
		}
	}
	return "", NotFoundError{Key: fullKey}
}

func (c *config) GetBool(key string) (bool, error) {
	v, err := c.Get(key)
	if err != nil {
		return false, err
	}
	return cfgconv.Bool(v)
}

func (c *config) GetConfig(key string) (Config, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	aliases := make(map[string][]string, len(c.aliases))
	for k, v := range c.aliases {
		aliases[k] = v[:]
	}
	readers := c.readers[:]
	return &config{
		prefix:    c.prefixedKey(key),
		separator: c.separator,
		readers:   readers,
		aliases:   aliases,
	}, nil
}

func (c *config) GetFloat(key string) (float64, error) {
	v, err := c.Get(key)
	if err != nil {
		return 0, err
	}
	return cfgconv.Float(v)
}

func (c *config) GetInt(key string) (int64, error) {
	v, err := c.Get(key)
	if err != nil {
		return 0, err
	}
	return cfgconv.Int(v)
}

func (c *config) GetIntSlice(key string) ([]int64, error) {
	slice, err := c.GetSlice(key)
	if err != nil {
		return []int64(nil), err
	}
	retval := make([]int64, 0, len(slice))
	for _, v := range slice {
		cv, err := cfgconv.Int(v)
		if err != nil {
			return []int64(nil), err
		}
		retval = append(retval, cv)
	}
	return retval, nil
}

func (c *config) GetSlice(key string) ([]interface{}, error) {
	v, err := c.Get(key)
	if err != nil {
		return []interface{}(nil), err
	}
	return cfgconv.Slice(v)
}

func (c *config) GetString(key string) (string, error) {
	v, err := c.Get(key)
	if err != nil {
		return "", err
	}
	return cfgconv.String(v)
}

func (c *config) GetStringSlice(key string) ([]string, error) {
	slice, err := c.GetSlice(key)
	if err != nil {
		return []string(nil), err
	}
	retval := make([]string, 0, len(slice))
	for _, v := range slice {
		cv, err := cfgconv.String(v)
		if err != nil {
			return []string(nil), err
		}
		retval = append(retval, cv)
	}
	return retval, nil
}

func (c *config) GetUint(key string) (uint64, error) {
	v, err := c.Get(key)
	if err != nil {
		return 0, err
	}
	return cfgconv.Uint(v)
}

func (c *config) GetUintSlice(key string) ([]uint64, error) {
	slice, err := c.GetSlice(key)
	if err != nil {
		return []uint64(nil), err
	}
	retval := make([]uint64, 0, len(slice))
	for _, v := range slice {
		cv, err := cfgconv.Uint(v)
		if err != nil {
			return []uint64(nil), err
		}
		retval = append(retval, cv)
	}
	return retval, nil
}

// NotFoundError indicates that the Key could not be found in the config tree.
type NotFoundError struct {
	Key string
}

func (e NotFoundError) Error() string {
	return "config: key '" + e.Key + "' not found"
}

// UnmarshalError indicates an error occured while unmarhalling config into
// a struct or map.  The error indicates the problematic Key and the specific
// error.
type UnmarshalError struct {
	Key string
	Err error
}

func (e UnmarshalError) Error() string {
	return "config: cannot unmarshal " + e.Key + " - " + e.Err.Error()
}

// Unmarshal a section of the config tree into a struct.
//
// The node identifies the section of the tree to unmarshal.
// The obj is struct with fields corresponding to config values.
// The config values will be converted to the type defined in the corresponding
// struct fields.
// If non-nil, the config values will be converted to the type already contained in the map.
// If nil then the value is set to the raw value returned by the Reader.
//
// By default the config field names are drawn from the struct field, lower cased.
// This can be overridden using `config:"<name>"` tags.
//
// Struct fields which do not have corresponding config fields are ignored,
// as are config fields which have no corresponding struct field.
//
// The error identifies first type conversion error, if any.
func (c *config) Unmarshal(node string, obj interface{}) (rerr error) {
	nodeCfg, _ := c.GetConfig(node)
	ov := reflect.Indirect(reflect.ValueOf(obj))
	if ov.Kind() != reflect.Struct {
		return fmt.Errorf("Unmarshal: obj is not a struct - %v", obj)
	}
	for idx := 0; idx < ov.NumField(); idx++ {
		fv := ov.Field(idx)
		ft := ov.Type().Field(idx)
		key := ft.Tag.Get("config")
		if len(key) == 0 {
			key = ft.Name
		}
		if fv.Kind() == reflect.Struct {
			// nested struct
			err := nodeCfg.Unmarshal(key, fv.Addr().Interface())
			if err != nil && rerr == nil {
				rerr = err
			}
		} else {
			// else assume a leaf
			if v, err := nodeCfg.Get(key); err == nil {
				if cv, err := cfgconv.Convert(v, fv.Type()); err == nil {
					fv.Set(reflect.ValueOf(cv))
				} else if rerr == nil {
					rerr = UnmarshalError{c.prefixedKey(node, strings.ToLower(key)), err}
				}
			}
		}
	}
	return rerr
}

// Unmarshal a section of the config tree into a map[string]interface{}.
//
// The node identifies the section of the tree to unmarshal.
// The objmap keys define the fields to be populated from config.
// If non-nil, the config values will be converted to the type already contained in the map.
// If nil then the value is set to the raw value returned by the Reader.
//
// Nested objects can be populated by adding them as map[string]interface{},
// with keys set corresponding to the nested field names.
//
// Map keys which do not have corresponding config fields are ignored,
// as are config fields which have no corresponding map key.
//
// The error identifies first type conversion error, if any.
func (c *config) UnmarshalToMap(node string, objmap map[string]interface{}) (rerr error) {
	nodeCfg, _ := c.GetConfig(node)
	for key := range objmap {
		vv := reflect.ValueOf(objmap[key])
		if !vv.IsValid() {
			// raw value
			if v, err := nodeCfg.Get(key); err == nil {
				objmap[key] = v
			}
		} else if v, ok := objmap[key].(map[string]interface{}); ok {
			// nested map
			err := nodeCfg.UnmarshalToMap(key, v)
			if err != nil && rerr == nil {
				rerr = err
			}
		} else if v, err := nodeCfg.Get(key); err == nil {
			// else assume a leaf
			if cv, err := cfgconv.Convert(v, vv.Type()); err == nil {
				objmap[key] = cv
			} else if rerr == nil {
				rerr = UnmarshalError{c.prefixedKey(node, key), err}
			}
		}
	}
	return rerr
}
