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
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/warthog618/config/cfgconv"
)

// New creates a new Config with minimal initial state.
func New(options ...Option) *Config {
	c := Config{
		separator: ".",
	}
	for _, option := range options {
		option(&c)
	}
	return &c
}

// Config provides a unified key/value store of configuration.
type Config struct {
	// RWLock covering other fields.
	// It does not prevent concurrent access to the getters themselves,
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
	// A list of Getters providing config key/value pairs.
	gg []Getter
	// The getter of last resort, assuming it is set.
	// If it is set then it is also the last entry in gg.
	def Getter
	// A map to a list of old names for current config keys.
	aliases map[string][]string
}

// Getter provides the minimal interface for a configuration Getter.
type Getter interface {
	// Get the value of the named config leaf key.
	// Also returns an ok, similar to a map read, to indicate if the value
	// was found.
	// The type underlying the returned interface{} must be convertable to
	// the expected type by cfgconv.
	// Get is not expected to be performed on node keys, but in case it is
	// the Get should return a nil interface{} and false, even if the node
	// exists in the config tree.
	// Must be safe to call from multiple goroutines.
	Get(key string) (interface{}, bool)
}

// Option is a function which modifies a Config at construction time.
type Option func(*Config)

// WithDefault is an Option that passes the default config.
// The default getter is searched only when the key is not found in
// any of the getters.
// If called multiple times then all calls other than the final are ignored.
func WithDefault(d Getter) Option {
	return func(c *Config) {
		if c.def != nil && c.gg != nil {
			c.gg = c.gg[:len(c.gg)-1]
		}
		c.def = d
		if c.def != nil {
			c.gg = append(c.gg, c.def)
		}
	}
}

// WithGetters is an Option that passes an initial set of getters.
// The provided set is copied, so subsequent changes to that set will
// not alter the Config.
func WithGetters(gg []Getter) Option {
	return func(c *Config) {
		c.gg = append([]Getter(nil), gg...)
		if c.def != nil {
			c.gg = append(c.gg, c.def)
		}
	}
}

// WithSeparator is an Option that sets the config namespace separator.
// This is an option to ensure it can only set at construction time,
// as changing it at runtime makes no sense.
func WithSeparator(separator string) Option {
	return func(c *Config) {
		c.separator = separator
	}
}

// AppendGetter appends a getter to the set of getters for the config node.
// This means this getter is only used as a last resort, relative to
// the existing getters.
//
// This is generally applied to the root node.
// When applied to a non-root node, the getter only applies to that node,
// and any subsequently created children of that node.
func (c *Config) AppendGetter(g Getter) {
	if g == nil {
		return
	}
	c.mu.Lock()
	if c.def == nil {
		c.gg = append(c.gg, g)
	} else {
		c.gg = append(append(c.gg[:len(c.gg)-1], g), c.def)
	}
	c.mu.Unlock()
}

// InsertGetter inserts a getter to the set of getters for the config node.
// This means this getter is used before the existing getters.
//
// This is generally applied to the root node.
// When applied to a non-root node, the getter only applies to that node,
// and any subsequently created children of that node.
func (c *Config) InsertGetter(g Getter) {
	if g == nil {
		return
	}
	c.mu.Lock()
	c.gg = append([]Getter{g}, c.gg...)
	c.mu.Unlock()
}

// prefixedKey returns the absolute config key for a key relative to the config node.
func (c *Config) prefixedKey(key ...string) string {
	rhs := strings.Join(key, c.separator)
	if len(c.prefix) == 0 {
		// we are root.
		return rhs
	}
	return c.prefix + c.separator + rhs
}

// AddAlias adds an alias from an old key, which may still be present in legacy config,
// to a new key, which should be the one used by the code.
// Aliases are ignored by Get if there is a config field matching the new key.
// Multiple aliases may be added for a new key, and they are searched for
// in the config in the reverse order they are added.
// Multiple aliases may also be added for an old key.
// When applied to a non-root node, the alias only applies to that node,
// and any subsequently created children.
func (c *Config) AddAlias(newKey string, oldKey string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	lNewKey := c.prefixedKey(newKey)
	lOldKey := c.prefixedKey(oldKey)
	if lNewKey == lOldKey {
		return
	}
	if c.aliases == nil {
		c.aliases = make(map[string][]string)
	}
	aliases, ok := c.aliases[lNewKey]
	if !ok {
		aliases = make([]string, 1)
	}
	// prepended so alias are searched in LIFO order.
	c.aliases[lNewKey] = append([]string{lOldKey}, aliases...)
}

// Get gets the raw value corresponding to the key.
// It iterates through the list of getters, searching for a matching key,
// or failing that for a matching alias.
// Returns the first match found, or an error if none is found.
func (c *Config) Get(key string) (interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	fullKey := c.prefixedKey(key)
	for _, g := range c.gg {
		if v, ok := g.Get(fullKey); ok {
			return v, nil
		}
		if len(c.aliases) > 0 {
			if v, ok := c.getLeafAlias(g, fullKey); ok {
				return v, nil
			}
			if v, ok := c.getNodeAlias(g, fullKey); ok {
				return v, nil
			}
		}
	}
	return nil, NotFoundError{Key: fullKey}
}

func (c *Config) getLeafAlias(g Getter, key string) (interface{}, bool) {
	if aliases, ok := c.aliases[key]; ok {
		for _, alias := range aliases {
			if v, ok := g.Get(alias); ok {
				return v, true
			}
		}
	}
	return nil, false
}

func (c *Config) getNodeAlias(g Getter, key string) (interface{}, bool) {
	path := strings.Split(key, c.separator)
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
				aliasKey := alias + key[idx:]
				if v, ok := g.Get(aliasKey); ok {
					return v, true
				}
			}
		}
	}
	return nil, false
}

// GetBool gets the value corresponding to the key and converts it to a bool,
// if possible.
// Returns false and an error if not possible.
func (c *Config) GetBool(key string) (bool, error) {
	v, err := c.Get(key)
	if err != nil {
		return false, err
	}
	return cfgconv.Bool(v)
}

// GetConfig gets the config corresponding to a subtree of the config,
// where the node identifies the root node of the config returned.
func (c *Config) GetConfig(node string) (*Config, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	aliases := make(map[string][]string, len(c.aliases))
	for k, v := range c.aliases {
		aliases[k] = append([]string(nil), v...)
	}
	return &Config{
		prefix:    c.prefixedKey(node),
		separator: c.separator,
		def:       c.def,
		gg:        append([]Getter(nil), c.gg...),
		aliases:   aliases,
	}, nil
}

// GetDuration gets the value corresponding to the key and converts it to
// a time.Duration, if possible.
// Returns 0 and an error if not possible.
func (c *Config) GetDuration(key string) (time.Duration, error) {
	v, err := c.Get(key)
	if err != nil {
		return 0, err
	}
	return cfgconv.Duration(v)
}

// GetFloat gets the value corresponding to the key and converts it to
// a float64, if possible.
// Returns 0 and an error if not possible.
func (c *Config) GetFloat(key string) (float64, error) {
	v, err := c.Get(key)
	if err != nil {
		return 0, err
	}
	return cfgconv.Float(v)
}

// GetInt gets the value corresponding to the key and converts it to
// an int64, if possible.
// Returns 0 and an error if not possible.
func (c *Config) GetInt(key string) (int64, error) {
	v, err := c.Get(key)
	if err != nil {
		return 0, err
	}
	return cfgconv.Int(v)
}

// GetIntSlice gets the value corresponding to the key and converts it to
// a slice of int64s, if possible.
// Returns nil and an error if not possible.
func (c *Config) GetIntSlice(key string) ([]int64, error) {
	slice, err := c.GetSlice(key)
	if err != nil {
		return nil, err
	}
	retval := make([]int64, 0, len(slice))
	for _, v := range slice {
		cv, err := cfgconv.Int(v)
		if err != nil {
			return nil, err
		}
		retval = append(retval, cv)
	}
	return retval, nil
}

// GetSlice gets the value corresponding to the key and converts it to
// a slice of []interface{}, if possible.
// Returns nil and an error if not possible.
func (c *Config) GetSlice(key string) ([]interface{}, error) {
	v, err := c.Get(key)
	if err != nil {
		return nil, err
	}
	return cfgconv.Slice(v)
}

// GetString gets the value corresponding to the key and converts it to
// an string, if possible.
// Returns an empty string and an error if not possible.
func (c *Config) GetString(key string) (string, error) {
	v, err := c.Get(key)
	if err != nil {
		return "", err
	}
	return cfgconv.String(v)
}

// GetStringSlice gets the value corresponding to the key and converts it to
// a slice of string, if possible.
// Returns nil and an error if not possible.
func (c *Config) GetStringSlice(key string) ([]string, error) {
	slice, err := c.GetSlice(key)
	if err != nil {
		return nil, err
	}
	retval := make([]string, 0, len(slice))
	for _, v := range slice {
		cv, err := cfgconv.String(v)
		if err != nil {
			return nil, err
		}
		retval = append(retval, cv)
	}
	return retval, nil
}

// GetTime gets the value corresponding to the key and converts it to
// a time.Time, if possible.
// Returns time.Time{} and an error if not possible.
func (c *Config) GetTime(key string) (time.Time, error) {
	v, err := c.Get(key)
	if err != nil {
		return time.Time{}, err
	}
	return cfgconv.Time(v)
}

// GetUint gets the value corresponding to the key and converts it to
// a iint64, if possible.
// Returns 0 and an error if not possible.
func (c *Config) GetUint(key string) (uint64, error) {
	v, err := c.Get(key)
	if err != nil {
		return 0, err
	}
	return cfgconv.Uint(v)
}

// GetUintSlice gets the value corresponding to the key and converts it to
// a slice of uint64, if possible.
// Returns nil and an error if not possible.
func (c *Config) GetUintSlice(key string) ([]uint64, error) {
	slice, err := c.GetSlice(key)
	if err != nil {
		return nil, err
	}
	retval := make([]uint64, 0, len(slice))
	for _, v := range slice {
		cv, err := cfgconv.Uint(v)
		if err != nil {
			return nil, err
		}
		retval = append(retval, cv)
	}
	return retval, nil
}

// Unmarshal a section of the config tree into a struct.
//
// The node identifies the section of the tree to unmarshal.
// The obj is a struct with fields corresponding to config values.
// The config values will be converted to the type defined in the corresponding
// struct fields.  Overflow checks are performed during conversion to ensure the
// value returned by the getter can fit within the designated field.
//
// By default the config field names are drawn from the struct field, converted to
// lowerCamelCase.
// This can be overridden using `config:"<name>"` tags.
//
// Struct fields which do not have corresponding config fields are ignored,
// as are config fields which have no corresponding struct field.
//
// The error identifies first type conversion error, if any.
func (c *Config) Unmarshal(node string, obj interface{}) (rerr error) {
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
			key = lowerCamelCase(ft.Name)
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
					rerr = UnmarshalError{c.prefixedKey(node, key), err}
				}
			}
		}
	}
	return rerr
}

// UnmarshalToMap unmarshals a section of the config tree into a map[string]interface{}.
//
// The node identifies the section of the tree to unmarshal.
// The objmap keys define the fields to be populated from config.
// If non-nil, the config values will be converted to the type already contained in the map.
// If nil then the value is set to the raw value returned by the Getter.
//
// Nested objects can be populated by adding them as map[string]interface{},
// with keys set corresponding to the nested field names.
//
// Map keys which do not have corresponding config fields are ignored,
// as are config fields which have no corresponding map key.
//
// The error identifies first type conversion error, if any.
func (c *Config) UnmarshalToMap(node string, objmap map[string]interface{}) (rerr error) {
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

// NotFoundError indicates that the Key could not be found in the config tree.
type NotFoundError struct {
	Key string
}

func (e NotFoundError) Error() string {
	return "config: key '" + e.Key + "' not found"
}

// UnmarshalError indicates an error occurred while unmarhalling config into
// a struct or map.  The error indicates the problematic Key and the specific
// error.
type UnmarshalError struct {
	Key string
	Err error
}

func (e UnmarshalError) Error() string {
	return "config: cannot unmarshal " + e.Key + " - " + e.Err.Error()
}

func lowerCamelCase(s string) string {
	r, n := utf8.DecodeRuneInString(s)
	return string(unicode.ToLower(r)) + s[n:]
}
