// Copyright © 2018 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package config

import (
	"regexp"
	"strings"
	"sync"
)

// NewAlias creates an Alias.
func NewAlias(options ...aliasOption) *Alias {
	a := &Alias{
		aa:      map[string][]string{},
		pathSep: "."}
	for _, option := range options {
		option.applyAliasOption(a)
	}
	return a
}

// WithAlias provides a decorator that calls the Getter, and falls back
// to a set of aliases if the lookup of the key fails.
func WithAlias(a *Alias) Decorator {
	return func(g Getter) Getter {
		return GetterFunc(func(key string) (interface{}, bool) {
			return a.Get(g, key)
		})
	}
}

// Alias provides a mapping from a key to a set of old or alternate keys.
type Alias struct {
	// mutex lock covering aa and the arrays it contains.
	mu      sync.RWMutex
	aa      map[string][]string
	pathSep string
}

// Get calls the Getter, and if that fails tries aliases to other keys.
func (a *Alias) Get(g Getter, key string) (interface{}, bool) {
	if v, ok := g.Get(key); ok {
		return v, true
	}
	if v, ok := a.getLeaf(g, key); ok {
		return v, true
	}
	if v, ok := a.getBranch(g, key); ok {
		return v, true
	}
	return nil, false
}

// Append adds an alias from the new key to the old.
// If aliases already exist for the new key then this appended to the end
// of the existing list.
func (a *Alias) Append(new, old string) {
	a.mu.Lock()
	a.aa[new] = append(a.aa[new], old)
	a.mu.Unlock()
}

// Insert adds an alias from the new key to the old.
// If aliases already exist for the new key then this inserted to the
// beginning of the existing list.
func (a *Alias) Insert(new, old string) {
	a.mu.Lock()
	a.aa[new] = append([]string{old}, a.aa[new]...)
	a.mu.Unlock()
}

// NewRegexAlias creates a RegexAlias.
func NewRegexAlias() *RegexAlias {
	return &RegexAlias{ra: []regex{}}
}

// WithRegexAlias provides a decorator that calls the Getter, and falls back
// to a set of regular expression aliases if the lookup of the key fails.
func WithRegexAlias(r *RegexAlias) Decorator {
	return func(g Getter) Getter {
		return GetterFunc(func(key string) (interface{}, bool) {
			return r.Get(g, key)
		})
	}
}

type regex struct {
	re  *regexp.Regexp
	old string
}

// RegexAlias provides a mapping from a key to an old key.
// New keys are expected to contain regular expressions.
type RegexAlias struct {
	mu sync.RWMutex
	ra []regex
}

// Append adds an alias from a regular expression matching a new key to the old.
func (r *RegexAlias) Append(new, old string) error {
	re, err := regexp.Compile(new)
	if err != nil {
		return err
	}
	r.mu.Lock()
	r.ra = append(r.ra, regex{re, old})
	r.mu.Unlock()
	return nil
}

// Get calls the Getter, and if that fails tries aliases to other keys.
func (r *RegexAlias) Get(g Getter, key string) (interface{}, bool) {
	if v, ok := g.Get(key); ok {
		return v, true
	}
	r.mu.RLock()
	for _, ra := range r.ra {
		if ra.re.MatchString(key) {
			k := ra.re.ReplaceAllString(key, ra.old)
			if v, ok := g.Get(k); ok {
				r.mu.RUnlock()
				return v, ok
			}
		}
	}
	r.mu.RUnlock()
	return nil, false
}

func (a *Alias) getLeaf(g Getter, key string) (interface{}, bool) {
	a.mu.RLock()
	if aliases, ok := a.aa[key]; ok {
		for _, alias := range aliases {
			if v, ok := g.Get(alias); ok {
				a.mu.RUnlock()
				return v, true
			}
		}
	}
	a.mu.RUnlock()
	return nil, false
}

func (a *Alias) getBranch(g Getter, key string) (interface{}, bool) {
	path := strings.Split(key, a.pathSep)
	for plen := len(path) - 1; plen >= 0; plen-- {
		nodeKey := strings.Join(path[:plen], a.pathSep)
		if aliases, ok := a.aa[nodeKey]; ok {
			for _, alias := range aliases {
				if len(alias) > 0 {
					alias = alias + a.pathSep
				}
				idx := len(nodeKey)
				if idx > 0 {
					idx += len(a.pathSep)
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

// aliasOption is a construction option for an Alias.
type aliasOption interface {
	applyAliasOption(c *Alias)
}
