// Copyright © 2017 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package config provides tools to retrieve configuration from various sources
// and present it to the application through a unified API.
package config

import (
	"fmt"
	"reflect"
	"sync"
	"unicode"
	"unicode/utf8"

	"github.com/warthog618/config/cfgconv"
)

// NewConfig creates a new Config with minimal initial state.
func NewConfig(g Getter, options ...Option) *Config {
	c := Config{
		getter:   g,
		pathSep:  ".",
		tag:      "config",
		notifier: NewNotifier(),
		bgmu:     &sync.RWMutex{},
		donech:   make(chan struct{}),
	}
	for _, option := range options {
		if gas, ok := option.(Getter); ok {
			c.getter = Overlay(c.getter, gas)
			continue
		}
		option.applyConfigOption(&c)
	}
	if wg, ok := g.(WatchableGetter); ok {
		c.gw = wg.NewWatcher(c.donech)
		if c.gw != nil {
			go c.watcher()
		}
	}
	return &c
}

// ErrorHandler handles an error.
type ErrorHandler func(error) error

// Config is a wrapper around a Getter that provides a set of conversion
// functions that return the requested type, if possible, or an error if not.
type Config struct {
	getter Getter
	// default Getter - used if a field is not found in the main getter.
	defg Getter
	// path separator for nested objects. This is used to split keys into a list
	// of tier names and leaf key. By default this is ".". e.g. a key
	// "db.postgres.client" splits into "db","postgres","client"
	pathSep string
	// tag identifies the field tags used to specify field names for Unmarshal.
	tag string
	// notifier indicates the config has been updated.
	notifier *Notifier
	// error handler for gets. Is propagated to Values unless overridden by
	// ValueOption.
	eh ErrorHandler
	// Mutex covering block gets - inhibits changes to underlying Loaders while
	// unmarshalling blocks, as that could result in inconsistent and
	// unpredicatable results.
	bgmu *sync.RWMutex
	// donech is closed to terminate all active watches and make config static.
	donech chan struct{}
	// the watcher for the getter.
	// Is nil if the getter is not watchable.
	gw GetterWatcher
}

func (c *Config) watcher() {
	for {
		select {
		case <-c.donech:
			return
		case update, ok := <-c.gw.Update():
			if !ok {
				return
			}
			c.bgmu.Lock()
			update.Commit()
			c.bgmu.Unlock()
			c.notifier.Notify()
		}
	}
}

// Append adds a getter to the end of the list of getters searched by the
// config, but still before a default getter specified by WithDefault.
// This function is not safe to call from multiple goroutines, and should only
// be called to set up configuration in a single goroutine before passing the
// final config to other goroutines.
func (c *Config) Append(g Getter) {
	if g == nil {
		return
	}
	c.getter = Overlay(c.getter, g)
}

// Close releases any resources allocated to the Config including
// cancelling any actve watches.
func (c *Config) Close() error {
	select {
	case <-c.donech:
		// already closed
	default:
		close(c.donech)
	}
	return nil
}

// Get gets the raw value corresponding to the key.
// Returns a zero Value and an error if the value cannot be retrieved.
func (c *Config) Get(key string, opts ...ValueOption) (Value, error) {
	var v interface{}
	var ok bool
	if c.getter != nil {
		v, ok = c.getter.Get(key)
	}
	if !ok && c.defg != nil {
		v, ok = c.defg.Get(key)
	}
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
func (c *Config) GetConfig(node string, options ...Option) *Config {
	g := c.getter
	d := c.defg
	if node != "" {
		g = Decorate(c.getter, WithPrefix(node+c.pathSep))
		d = Decorate(c.defg, WithPrefix(node+c.pathSep))
	}
	v := &Config{
		getter:   g,
		defg:     d,
		pathSep:  c.pathSep,
		tag:      c.tag,
		notifier: c.notifier,
		bgmu:     c.bgmu,
	}
	for _, option := range options {
		option.applyConfigOption(v)
	}
	return v
}

// Insert adds a getter to the beginning of the list of getters searched by the
// config.
// This function is not safe to call from multiple goroutines, and should only
// be called to set up configuration in a single goroutine before passing the
// final config to other goroutines.
func (c *Config) Insert(g Getter) {
	if g == nil {
		return
	}
	c.getter = Overlay(g, c.getter)
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
	c.bgmu.RLock()
	defer c.bgmu.RUnlock()
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
	c.bgmu.RLock()
	defer c.bgmu.RUnlock()
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

// Watcher provides a synchronous watch of the overall configuration state.
// The Watcher should not be called from multiple goroutines at a time.
// If you need to watch the config in multiple goroutines then create a Watcher
// per goroutine.
type Watcher struct {
	mu      sync.Mutex
	updated <-chan struct{}
	c       *Config
}

// NewWatcher creates a watch on the whole configuration.
func (c *Config) NewWatcher() *Watcher {
	return &Watcher{updated: c.notifier.Notified(), c: c}
}

// Watch returns when the configuration has been changed or the done is closed.
// Returns an error if the watch was ended unexpectedly.
// Watch should only be called once at a time.
// To prevent races, multiple simultaneous calls are serialised.
// If you need to watch the config in multiple goroutines then create a Watcher
// per goroutine.
func (w *Watcher) Watch(done <-chan struct{}) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	select {
	case <-w.c.donech:
		return ErrClosed
	case <-done:
		return ErrCanceled
	case <-w.updated:
		w.updated = w.c.notifier.Notified()
		return nil
	}
}

// KeyWatcher watches a particular key/value, with the next returning
// changed values.
// The KeyWatcher should not be called from multiple goroutines simulataneously.
// If you need to watch the key in multiple goroutines then create a KeyWatcher
// per goroutine.
type KeyWatcher struct {
	w      *Watcher
	getter func() (Value, error)
	last   *Value
}

// NewKeyWatcher creates a watch on the given key.
// The key should correspond to a field, not a node.
func (c *Config) NewKeyWatcher(key string, opts ...ValueOption) *KeyWatcher {
	getter := func() (Value, error) {
		return c.Get(key, opts...)
	}
	return &KeyWatcher{w: c.NewWatcher(), getter: getter}
}

// Watch returns the next value of the watched field.
// On the first call it immediately returns the current value.
// On subsequent calls it blocks until the value changes or the done is closed.
// Returns an error if the watch was ended unexpectedly.
// Watch should only be called once at a time - it does not support being called
// by multiple goroutines simultaneously.
func (w *KeyWatcher) Watch(done <-chan struct{}) (Value, error) {
	for {
		if w.last != nil {
			err := w.w.Watch(done)
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
