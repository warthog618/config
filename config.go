package config

import (
	"config/cfgconv"
	"fmt"
	"strings"
)

type Config interface {
	AddAlias(newKey string, oldKey string)
	AddReader(reader Reader)
	Contains(key string) bool
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
	GetUintSlice(key string) ([]uint64, error)
	GetStringSlice(key string) ([]string, error)
	// Tree get
	GetConfig(key string) (Config, error)
	GetObject(key string) (map[string]interface{}, error)
	// Defaulted gets - returns the def if not found in the config.
	GetBoolDef(key string, def bool) bool
	GetFloatDef(key string, def float64) float64
	GetIntDef(key string, def int64) int64
	GetStringDef(key string, def string) string
	GetUintDef(key string, def uint64) uint64
}

func New() Config {
	return &config{"", make([]Reader, 0), make(map[string][]string)}
}

type Reader interface {
	// Indicate whether this Reader provides the named config key.
	Contains(key string) bool
	// Read and return the value of the named config key.
	// Also returns an ok, similar to a map read, to indicate if the value
	// was found.
	// The underlying type must be convertable to the expected type by cfgconv.
	Read(key string) (interface{}, bool)
}

type config struct {
	// The prefix section of all keys within this config node.
	// For the root node this is empty.
	// For other nodes this indicates the location of the node in the config tree.
	// The keys passed into the non-root nodes should be local to that node,
	// i.e. they should treat the node as the root of their own config tree.
	// The node
	prefix string
	// A list of Readers providing config key/value pairs.
	readers []Reader
	// A map to a list of old names for existing config keys.
	aliases map[string][]string
}

// Add a reader to the config node.
// This is generally applied to the root node.
// When applied to a non-root node, the reader only applies to that node,
// and any subsequently created children.
func (c *config) AddReader(reader Reader) {
	// prepended so readers are searched in LIFO order.
	c.readers = append([]Reader{reader}, c.readers...)
}

// Return the absolute config key for a key relative to the config node.
func (c *config) prefixedKey(key string) string {
	if len(c.prefix) == 0 {
		// we are root.
		return key
	}
	return c.prefix + "." + key
}

// Add an alias from a newKey, which should be used by the code,
// to an old key which may still be present in legacy config.
// As with readers, the aliases are local to the config node, and any
// subsequently created children.
func (c *config) AddAlias(newKey string, oldKey string) {
	lNewKey := c.prefixedKey(strings.ToLower(newKey))
	lOldKey := c.prefixedKey(strings.ToLower(oldKey))
	aliases, ok := c.aliases[lNewKey]
	if ok != true {
		aliases = make([]string, 1)
	}
	// prepended so alias are searched in LIFO order.
	c.aliases[lNewKey] = append([]string{lOldKey}, aliases...)
}

// Returns true of the key is contained in the config tree.
// Key may be a leaf or a node.
func (c *config) Contains(key string) bool {
	lowerKey := c.prefixedKey(strings.ToLower(key))
	for _, reader := range c.readers {
		if reader.Contains(lowerKey) {
			return true
		}
		if aliases, ok := c.aliases[lowerKey]; ok {
			for _, alias := range aliases {
				if reader.Contains(alias) {
					return true
				}
			}
		}
	}
	return false
}

// Get the raw string value corresponding to the key.
// Iterates through the list of readers, searching for a matching key,
// or matching alias.  Returns the first match found, or an error if none is found.
func (c *config) Get(key string) (interface{}, error) {
	lowerKey := strings.ToLower(key)
	for _, reader := range c.readers {
		if val, ok := reader.Read(c.prefixedKey(lowerKey)); ok {
			return val, nil
		}
		if aliases, ok := c.aliases[lowerKey]; ok {
			for _, alias := range aliases {
				if val, ok := reader.Read(c.prefixedKey(alias)); ok {
					return val, nil
				}
			}
		}
	}
	return "", fmt.Errorf("config key '%s' not found", key)
}

func (c *config) GetBool(key string) (bool, error) {
	val, err := c.Get(key)
	if err != nil {
		return false, err
	}
	return cfgconv.Bool(val)
}

func (c *config) GetBoolDef(key string, def bool) bool {
	if val, err := c.GetBool(key); err == nil {
		return val
	} else {
		return def
	}
}

func (c *config) GetConfig(key string) (Config, error) {
	return &config{c.prefixedKey(key), c.readers, c.aliases}, nil
}

func (c *config) GetFloat(key string) (float64, error) {
	val, err := c.Get(key)
	if err != nil {
		return 0, err
	}
	return cfgconv.Float(val)
}

func (c *config) GetFloatDef(key string, def float64) float64 {
	if val, err := c.GetFloat(key); err == nil {
		return val
	} else {
		return def
	}
}

func (c *config) GetInt(key string) (int64, error) {
	val, err := c.Get(key)
	if err != nil {
		return 0, err
	}
	return cfgconv.Int(val)
}

func (c *config) GetIntDef(key string, def int64) int64 {
	if val, err := c.GetInt(key); err == nil {
		return val
	} else {
		return def
	}
}

func (c *config) GetIntSlice(key string) ([]int64, error) {
	slice, err := c.GetSlice(key)
	if err != nil {
		return []int64{}, fmt.Errorf("config key '%s' is not a slice", key)
	}
	retval := make([]int64, 0, len(slice))
	for _, val := range slice {
		valint, err := cfgconv.Int(val)
		if err != nil {
			return []int64{}, err
		}
		retval = append(retval, valint)
	}
	return retval, nil
}

func (c *config) GetObject(key string) (map[string]interface{}, error) {
	val, err := c.Get(key)
	if err != nil {
		return map[string]interface{}{}, err
	}
	return cfgconv.Object(val)
}

func (c *config) GetSlice(key string) ([]interface{}, error) {
	val, err := c.Get(key)
	if err != nil {
		return []interface{}{}, err
	}
	return cfgconv.Slice(val)
}

func (c *config) GetString(key string) (string, error) {
	val, err := c.Get(key)
	if err != nil {
		return "", err
	}
	return cfgconv.String(val)
}

func (c *config) GetStringDef(key string, def string) string {
	if val, err := c.GetString(key); err == nil {
		return val
	} else {
		return def
	}
}

func (c *config) GetStringSlice(key string) ([]string, error) {
	slice, err := c.GetSlice(key)
	if err != nil {
		return []string{}, fmt.Errorf("config key '%s' is not a slice", key)
	}
	retval := make([]string, 0, len(slice))
	for _, val := range slice {
		valstr, err := cfgconv.String(val)
		if err != nil {
			return []string{}, err
		}
		retval = append(retval, valstr)
	}
	return retval, nil
}

func (c *config) GetUint(key string) (uint64, error) {
	val, err := c.Get(key)
	if err != nil {
		return 0, err
	}
	return cfgconv.Uint(val)
}

func (c *config) GetUintDef(key string, def uint64) uint64 {
	if val, err := c.GetUint(key); err == nil {
		return val
	} else {
		return def
	}
}

func (c *config) GetUintSlice(key string) ([]uint64, error) {
	slice, err := c.GetSlice(key)
	if err != nil {
		return []uint64{}, fmt.Errorf("config key '%s' is not a slice", key)
	}
	retval := make([]uint64, 0, len(slice))
	for _, val := range slice {
		valuint, err := cfgconv.Uint(val)
		if err != nil {
			return []uint64{}, err
		}
		retval = append(retval, valuint)
	}
	return retval, nil
}
