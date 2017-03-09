// Environment reader for config.
package env

import (
	"os"
	"strings"
)

func New(prefix string) (*reader, error) {
	config := map[string]string(nil)
	nodes := map[string]bool(nil)
	r := reader{config, nodes, prefix, "_", "", ".", ":"}
	r.load()
	return &r, nil
}

type reader struct {
	// config key=value
	config map[string]string
	// set of nodes contained in config
	nodes map[string]bool
	// prefix in ENV space.
	// This must include any separator - the envSeparator does not separate the
	// prefix from the remainder of the key.
	envPrefix string
	// separator between key tiers in ENV space.
	envSeparator string
	// string that defines the env root in config space
	cfgPrefix string
	// separator between key tiers in config space.
	cfgSeparator string
	// The separator for slices stored in string values.
	listSeparator string
}

func (r *reader) Contains(key string) bool {
	if len(r.cfgPrefix) > 0 {
		if !strings.HasPrefix(key, r.cfgPrefix) {
			return false
		}
		key = key[len(r.cfgPrefix):]
	}
	if _, ok := r.config[key]; ok {
		return true
	}
	if _, ok := r.nodes[key]; ok {
		return true
	}
	return false
}

func (r *reader) Read(key string) (interface{}, bool) {
	if len(r.cfgPrefix) > 0 {
		if !strings.HasPrefix(key, r.cfgPrefix) {
			return nil, false
		}
		key = key[len(r.cfgPrefix):]
	}
	val, ok := r.config[key]
	if ok && len(r.listSeparator) > 0 && strings.Contains(val, r.listSeparator) {
		return strings.Split(val, r.listSeparator), ok
	}
	return val, ok
}

func (r *reader) SetCfgPrefix(prefix string) {
	r.cfgPrefix = strings.ToLower(prefix)
}

func (r *reader) SetCfgSeparator(separator string) {
	r.cfgSeparator = strings.ToLower(separator)
	r.load()
}

func (r *reader) SetEnvPrefix(prefix string) {
	r.envPrefix = prefix
	r.load()
}

func (r *reader) SetEnvSeparator(separator string) {
	r.envSeparator = separator
	r.load()
}

func (r *reader) SetListSeparator(separator string) {
	r.listSeparator = separator
}

func (r *reader) load() {
	config := map[string]string{}
	nodes := map[string]bool{}
	for _, env := range os.Environ() {
		if strings.HasPrefix(env, r.envPrefix) {
			keyValue := strings.SplitN(env, "=", 2)
			if len(keyValue) == 2 {
				envKey := keyValue[0][len(r.envPrefix):]
				path := strings.Split(envKey, r.envSeparator)
				cfgKey := strings.ToLower(strings.Join(path, r.cfgSeparator))
				config[cfgKey] = keyValue[1]
				nodePath := path[0]
				for idx := 1; idx < len(path); idx++ {
					nodes[strings.ToLower(nodePath)] = true
					nodePath = nodePath + r.cfgSeparator + path[idx]
				}
			}
		}
	}
	r.config, r.nodes = config, nodes
}
