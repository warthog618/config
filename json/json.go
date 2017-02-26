// JSON format reader for config.
package json

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
)

type reader struct {
	config map[string]interface{}
}

func (r *reader) Contains(key string) bool {
	_, ok := r.Read(key)
	return ok
}

func (r *reader) Read(key string) (interface{}, bool) {
	if val, err := getFromMapTree(r.config, key, "."); err == nil {
		return val, true
	} else {
		return nil, false
	}
}

func NewBytes(cfg []byte) (*reader, error) {
	var config map[string]interface{}
	err := json.Unmarshal(cfg, &config)
	if err != nil {
		return &reader{}, err
	}
	return &reader{config}, nil
}

func NewFile(filename string) (*reader, error) {
	cfg, err := ioutil.ReadFile(filename)
	if err != nil {
		return &reader{}, err
	}
	return NewBytes(cfg)
}

func getFromMapTree(node map[string]interface{}, key string, pathSep string) (interface{}, error) {
	// full key match - also handles leaves
	if value, ok := node[key]; ok {
		return value, nil
	}
	// nested path match
	path := strings.Split(key, pathSep)
	if value, ok := node[path[0]]; ok {
		switch v := value.(type) {
		case map[string]interface{}:
			return getFromMapTree(v, strings.Join(path[1:], pathSep), pathSep)
		}
	}
	// no match
	return nil, fmt.Errorf("key '%v' not found", key)
}
