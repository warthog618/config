// Copyright © 2017 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package config

import (
	"reflect"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/warthog618/config/cfgconv"
)

// NewConfig creates a new Config with minimal initial state.
func NewConfig(g Getter, options ...ConfigOption) *Config {
	c := Config{
		getter:  g,
		pathSep: ".",
		tag:     "config",
	}
	for _, option := range options {
		option.applyConfigOption(&c)
	}
	return &c
}

// NewMust creates a new Must with minimal initial state.
func NewMust(g Getter, options ...MustOption) *Must {
	m := Must{
		c: &Config{getter: g,
			pathSep: ".",
			tag:     "config",
		}}
	for _, option := range options {
		option.applyMustOption(&m)
	}
	return &m
}

// ErrorHandler handles an error.
type ErrorHandler func(error)

// Config is a wrapper around a Getter that provides a set
// of conversion functions that return the requested type, if possible,
// or an error if not.
type Config struct {
	getter Getter
	// path separator for nested objects.
	// This is used to split keys into a list of tier names and leaf key.
	// By default this is ".".
	// e.g. a key "db.postgres.client" splits into  "db","postgres","client"
	pathSep string
	// tag identifies the field tags used to specify field names for Unmarshal.
	tag string
}

// Must is a wrapper around a Getter that provides a set
// of conversion functions that return the requested type, if possible,
// or a zero value if not.
type Must struct {
	c *Config
	e ErrorHandler
}

// Get gets the raw value corresponding to the key.
func (c *Config) Get(key string) (interface{}, error) {
	if v, ok := c.getter.Get(key); ok {
		return v, nil
	}
	return nil, NotFoundError{Key: key}
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

// GetConfig gets the Config corresponding to a subtree of the config,
// where the node identifies the root node of the config returned.
func (c *Config) GetConfig(node string, options ...ConfigOption) *Config {
	g := c.getter
	if node != "" {
		g = Decorate(c.getter, WithPrefix(node+c.pathSep))
	}
	v := &Config{
		getter:  g,
		pathSep: c.pathSep,
		tag:     c.tag,
	}
	for _, option := range options {
		option.applyConfigOption(v)
	}
	return v
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
func (c *Config) GetIntSlice(key string) (retval []int64, rerr error) {
	slice, err := c.GetSlice(key)
	if err != nil {
		return nil, err
	}
	retval = make([]int64, len(slice))
	for i, v := range slice {
		cv, err := cfgconv.Int(v)
		if err == nil {
			retval[i] = cv
		} else if rerr == nil {
			rerr = err
		}
	}
	return
}

// GetMust gets the Must corresponding to a subtree of the config,
// where the node identifies the root node of the config returned.
func (c *Config) GetMust(node string, options ...MustOption) *Must {
	m := &Must{
		c: c.GetConfig(node),
		e: nil,
	}
	for _, option := range options {
		option.applyMustOption(m)
	}
	return m
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
func (c *Config) GetStringSlice(key string) (retval []string, rerr error) {
	slice, err := c.GetSlice(key)
	if err != nil {
		return nil, err
	}
	retval = make([]string, len(slice))
	for i, v := range slice {
		cv, err := cfgconv.String(v)
		if err == nil {
			retval[i] = cv
		} else if rerr == nil {
			rerr = err
		}
	}
	return
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
func (c *Config) GetUintSlice(key string) (retval []uint64, rerr error) {
	slice, err := c.GetSlice(key)
	if err != nil {
		return nil, err
	}
	retval = make([]uint64, len(slice))
	for i, v := range slice {
		cv, err := cfgconv.Uint(v)
		if err == nil {
			retval[i] = cv
		} else if rerr == nil {
			rerr = err
		}
	}
	return
}

// Unmarshal a section of the config tree into a struct.
//
// The node identifies the section of the tree to unmarshal.
// The obj is a pointer to a struct with fields corresponding to config values.
// The config values will be converted to the type defined in the corresponding
// struct fields.  Overflow checks are performed during conversion to ensure the
// value returned by the getter can fit within the designated field.
//
// By default the config field names are drawn from the struct field,
// converted to LowerCamelCase (as per typical JSON naming conventions).
// This can be overridden using `config:"<name>"` tags.
//
// Struct fields which do not have corresponding config fields are ignored,
// as are config fields which have no corresponding struct field.
//
// The error identifies the first type conversion error, if any.
func (c *Config) Unmarshal(node string, obj interface{}) (rerr error) {
	ov := reflect.ValueOf(obj)
	if ov.Kind() != reflect.Ptr {
		return ErrInvalidStruct
	}
	ov = reflect.Indirect(reflect.ValueOf(obj))
	if ov.Kind() != reflect.Struct {
		return ErrInvalidStruct
	}
	nodeCfg := c.GetConfig(node)
	for idx := 0; idx < ov.NumField(); idx++ {
		fv := ov.Field(idx)
		if !fv.CanSet() {
			// ignore unexported fields.
			continue
		}
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
					rerr = UnmarshalError{node + c.pathSep + key, err}
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
// The error identifies the first type conversion error, if any.
func (c *Config) UnmarshalToMap(node string, objmap map[string]interface{}) (rerr error) {
	nodeCfg := c.GetConfig(node)
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
				rerr = UnmarshalError{node + c.pathSep + key, err}
			}
		}
	}
	return rerr
}

// Get gets the raw value corresponding to the key.
// It iterates through the list of getters, searching for a matching key,
// or failing that for a matching alias.
// Returns the first match found, or nil if none is found.
func (m *Must) Get(key string) interface{} {
	v, err := m.c.Get(key)
	if err != nil && m.e != nil {
		m.e(err)
	}
	return v
}

// GetBool gets the value corresponding to the key and converts it to a bool,
// if possible.
// Returns false if not possible.
func (m *Must) GetBool(key string) bool {
	v, err := m.c.GetBool(key)
	if err != nil && m.e != nil {
		m.e(err)
	}
	return v
}

// GetConfig gets the Config corresponding to a subtree of the config,
// where the node identifies the root node of the config returned.
func (m *Must) GetConfig(node string, options ...ConfigOption) *Config {
	return m.c.GetConfig(node, options...)
}

// GetDuration gets the value corresponding to the key and converts it to
// a time.Duration, if possible.
// Returns 0 if not possible.
func (m *Must) GetDuration(key string) time.Duration {
	v, err := m.c.GetDuration(key)
	if err != nil && m.e != nil {
		m.e(err)
	}
	return v
}

// GetFloat gets the value corresponding to the key and converts it to
// a float64, if possible.
// Returns 0 and an error if not possible.
func (m *Must) GetFloat(key string) float64 {
	v, err := m.c.GetFloat(key)
	if err != nil && m.e != nil {
		m.e(err)
	}
	return v
}

// GetInt gets the value corresponding to the key and converts it to
// an int64, if possible.
// Returns 0 and an error if not possible.
func (m *Must) GetInt(key string) int64 {
	v, err := m.c.GetInt(key)
	if err != nil && m.e != nil {
		m.e(err)
	}
	return v
}

// GetIntSlice gets the value corresponding to the key and converts it to
// a slice of int64s, if possible.
// Returns nil if not possible.
func (m *Must) GetIntSlice(key string) []int64 {
	v, err := m.c.GetIntSlice(key)
	if err != nil && m.e != nil {
		m.e(err)
	}
	return v
}

// GetMust gets the Must corresponding to a subtree of the config,
// where the node identifies the root node of the config returned.
func (m *Must) GetMust(node string, options ...MustOption) *Must {
	m = &Must{
		c: m.c.GetConfig(node),
		e: m.e,
	}
	for _, option := range options {
		option.applyMustOption(m)
	}
	return m
}

// GetSlice gets the value corresponding to the key and converts it to
// a slice of []interface{}, if possible.
// Returns nil if not possible.
func (m *Must) GetSlice(key string) []interface{} {
	v, err := m.c.GetSlice(key)
	if err != nil && m.e != nil {
		m.e(err)
	}
	return v
}

// GetString gets the value corresponding to the key and converts it to
// an string, if possible.
// Returns an empty string if not possible.
func (m *Must) GetString(key string) string {
	v, err := m.c.GetString(key)
	if err != nil && m.e != nil {
		m.e(err)
	}
	return v
}

// GetStringSlice gets the value corresponding to the key and converts it to
// a slice of string, if possible.
// Returns nil if not possible to convert to array, and empty strings if
// elements cannot be converted to string.
func (m *Must) GetStringSlice(key string) []string {
	v, err := m.c.GetStringSlice(key)
	if err != nil && m.e != nil {
		m.e(err)
	}
	return v
}

// GetTime gets the value corresponding to the key and converts it to
// a time.Time, if possible.
// Returns time.Time{} if not possible.
func (m *Must) GetTime(key string) time.Time {
	v, err := m.c.GetTime(key)
	if err != nil && m.e != nil {
		m.e(err)
	}
	return v
}

// GetUint gets the value corresponding to the key and converts it to
// a iint64, if possible.
// Returns 0 and an error if not possible.
func (m *Must) GetUint(key string) uint64 {
	v, err := m.c.GetUint(key)
	if err != nil && m.e != nil {
		m.e(err)
	}
	return v
}

// GetUintSlice gets the value corresponding to the key and converts it to
// a slice of uint64, if possible.
// Returns nil and an error if not possible.
func (m *Must) GetUintSlice(key string) []uint64 {
	v, err := m.c.GetUintSlice(key)
	if err != nil && m.e != nil {
		m.e(err)
	}
	return v
}

// Unmarshal a section of the config tree into a struct.
//
// The node identifies the section of the tree to unmarshal.
// The obj is a pointer to a struct with fields corresponding to config values.
// The config values will be converted to the type defined in the corresponding
// struct fields.  Overflow checks are performed during conversion to ensure the
// value returned by the getter can fit within the designated field.
//
// By default the config field names are drawn from the struct field,
// converted to LowerCamelCase (as per typical JSON naming conventions).
// This can be overridden using `config:"<name>"` tags.
//
// Struct fields which do not have corresponding config fields are ignored,
// as are config fields which have no corresponding struct field.
func (m *Must) Unmarshal(node string, obj interface{}) {
	err := m.c.Unmarshal(node, obj)
	if err != nil && m.e != nil {
		m.e(err)
	}
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
func (m *Must) UnmarshalToMap(node string, objmap map[string]interface{}) {
	err := m.c.UnmarshalToMap(node, objmap)
	if err != nil && m.e != nil {
		m.e(err)
	}
}

// lowerCamelCase converts the first rune of a string to lower case.
// The function assumes key is already camel cased, so only
// lower cases the leading character.
// This is used to convert Go exported field names to config space keys.
// e.g. ConfigFile becomes configFile.
func lowerCamelCase(key string) string {
	r, n := utf8.DecodeRuneInString(key)
	return string(unicode.ToLower(r)) + key[n:]
}
