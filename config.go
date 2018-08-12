// Copyright © 2017 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package config

import (
	"context"
	"fmt"
	"reflect"
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

// ErrorHandler handles an error.
type ErrorHandler func(error) error

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
	// notifier indicates the config has been updated.
	notifier *Notifier
	// error handler for gets
	// Propagated to Values unless overridden by ValueOption.
	eh ErrorHandler
}

// Get gets the raw value corresponding to the key.
// Returns a zero Value and an error if the value cannot be retrieved.
func (c *Config) Get(key string, opts ...ValueOption) (Value, error) {
	v, ok := c.getter.Get(key)
	if !ok {
		var err error
		err = NotFoundError{Key: key}
		if c.eh != nil {
			err = c.eh(err)
		}
		if err != nil {
			return Value{}, err
		}
	}
	if c.eh != nil {
		opts = append([]ValueOption{WithErrorHandler(c.eh)}, opts...)
	}
	val := Value{value: v}
	for _, option := range opts {
		option.applyValueOption(&val)
	}
	return val, nil
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

// MustGet gets the value corresponding to the key, or panics if the key is not
// found. This is a convenience wrapper that allows chaining of calls to value
// conversions when the application is certain the config field will be present.
func (c *Config) MustGet(key string, opts ...ValueOption) Value {
	v, err := c.Get(key, opts...)
	if err != nil {
		panic(err)
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
//
// The error identifies the first type conversion error, if any.
func (c *Config) Unmarshal(node string, obj interface{}) (rerr error) {
	ov := getStructFromPtr(obj)
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
		key := ft.Tag.Get(nodeCfg.tag)
		if len(key) == 0 {
			key = lowerCamelCase(ft.Name)
		}
		switch fv.Kind() {
		case reflect.Struct:
			// nested struct
			err := nodeCfg.Unmarshal(key, fv.Addr().Interface())
			if rerr == nil {
				rerr = err
			}
		case reflect.Array, reflect.Slice:
			if fv.Type().Elem().Kind() == reflect.Struct {
				a, err := unmarshalObjectArray(nodeCfg, key, fv.Type())
				if rerr == nil {
					rerr = err
				}
				if !a.IsNil() {
					fv.Set(a)
				}
				continue
			}
			fallthrough
		default:
			// else assume a leaf
			if v, err := nodeCfg.Get(key); err == nil {
				if cv, err := cfgconv.Convert(v.Value(), fv.Type()); err == nil {
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
				objmap[key] = v.Value()
			}
			continue
		}
		switch v := objmap[key].(type) {
		case map[string]interface{}:
			// nested map
			err := nodeCfg.UnmarshalToMap(key, v)
			if rerr == nil {
				rerr = err
			}
		case []map[string]interface{}:
			// array of objects
			a, err := unmarshalObjectArrayToMap(nodeCfg, key, v)
			if rerr == nil {
				rerr = err
			}
			if a != nil {
				objmap[key] = a
			}
		default:
			if v, err := nodeCfg.Get(key); err == nil {
				// else assume a leaf
				if cv, err := cfgconv.Convert(v.Value(), vv.Type()); err == nil {
					objmap[key] = cv
				} else if rerr == nil {
					rerr = UnmarshalError{node + c.pathSep + key, err}
				}
			}
		}
	}
	return rerr
}

// Watcher provides a synchronous watch.
type Watcher struct {
	updated <-chan struct{}
	c       *Config
	ctx     context.Context
}

// Watch creates a watch on the whole configuration.
func (c *Config) Watch(ctx context.Context) *Watcher {
	var updated <-chan struct{}
	if c.notifier == nil {
		// static config so make next block forever on updated
		updated = make(chan struct{})
	} else {
		updated = c.notifier.Notified()
	}
	return &Watcher{updated: updated, c: c, ctx: ctx}
}

// Next returns when the configuration has been changed.
func (w *Watcher) Next() error {
	select {
	case <-w.ctx.Done():
		return w.ctx.Err()
	case <-w.updated:
		w.updated = w.c.notifier.Notified()
		return nil
	}
}

// ValueWatcher watches a particular key/value, with the next returning
// changed values.
type ValueWatcher struct {
	w      *Watcher
	getter func() (Value, error)
	last   *Value
}

// WatchKey creates a watch on the given key.
// The key should correspond to a field, not a node.
func (c *Config) WatchKey(ctx context.Context, key string, opts ...ValueOption) *ValueWatcher {
	getter := func() (Value, error) {
		return c.Get(key, opts...)
	}
	return &ValueWatcher{w: c.Watch(ctx), getter: getter}
}

// Next returns the next value of the watched field.
// On the first call it immediately returns the current value.
// On subsequent calls it blocks until the value changes or the ctx is done.
func (w *ValueWatcher) Next() (Value, error) {
	for {
		if w.last != nil {
			err := w.w.Next()
			if err != nil {
				return Value{}, err
			}
		}
		v, err := w.getter()
		if err != nil {
			return Value{}, err
		}
		if w.last == nil || v.Value() != w.last.Value() {
			w.last = &v
			return v, nil
		}
	}
}

func getStructFromPtr(obj interface{}) reflect.Value {
	ov := reflect.ValueOf(obj)
	if ov.Kind() != reflect.Ptr {
		return reflect.Value{}
	}
	return reflect.Indirect(reflect.ValueOf(obj))
}

func unmarshalObjectArray(node *Config, key string, t reflect.Type) (a reflect.Value, rerr error) {
	if v, err := node.Get(key + "[]"); err == nil {
		al64, err := cfgconv.Int(v.Value())
		if err != nil {
			return reflect.Zero(t), err
		}
		al := int(al64)
		a = reflect.MakeSlice(t, al, al)
		for i := 0; i < al; i++ {
			k := fmt.Sprintf("%s[%d]", key, i)
			err := node.Unmarshal(k, a.Index(i).Addr().Interface())
			if rerr == nil {
				rerr = err
			}
		}
		return a, rerr
	}
	return reflect.Zero(t), nil
}

func unmarshalObjectArrayToMap(node *Config, key string, tmpl []map[string]interface{}) (a []map[string]interface{}, rerr error) {
	if len(tmpl) == 0 {
		return
	}
	if alv, err := node.Get(key + "[]"); err == nil {
		al64, err := cfgconv.Int(alv.Value())
		if err != nil {
			rerr = err
		}
		if al64 == 0 {
			return a, rerr
		}
		al := int(al64)
		a = make([]map[string]interface{}, al, al)
		for i := 0; i < al; i++ {
			a[i] = make(map[string]interface{}, len(tmpl[0]))
			for k, v := range tmpl[0] {
				a[i][k] = v
			}
			k := fmt.Sprintf("%s[%d]", key, i)
			err := node.UnmarshalToMap(k, a[i])
			if rerr == nil {
				rerr = err
			}
		}
	}
	return a, rerr
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
