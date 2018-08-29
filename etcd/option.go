// Copyright © 2018 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package etcd

import (
	"github.com/coreos/etcd/clientv3"
	"github.com/warthog618/config/keys"
	"github.com/warthog618/config/list"
)

// Option is a function which modifies an Etcd getter at construction time.
type Option func(*Getter)

// WithClient sets the etcd client.
// The default is constructed from the clientConfig.
func WithClient(client *clientv3.Client) Option {
	return func(e *Getter) {
		e.client = client
	}
}

// WithClientConfig sets the configuration of the etcd client.
// The default is an empty config with one endpoint - "localhost:2379".
func WithClientConfig(cfg clientv3.Config) Option {
	return func(e *Getter) {
		e.clientConfig = cfg
	}
}

// WithEndpoint sets the endpoint URL of the client port of the etcd server.
// The default endpoint is "localhost:2379".
func WithEndpoint(endpoints ...string) Option {
	return func(e *Getter) {
		e.clientConfig.Endpoints = endpoints
	}
}

// WithKeyReplacer sets the replacer used to map from etcd space to config space.
// The default replaces '/' in the etcd space with '.' in the config space.
func WithKeyReplacer(keyReplacer keys.Replacer) Option {
	return func(e *Getter) {
		e.keyReplacer = keyReplacer
	}
}

// WithListSplitter splits slice fields stored as strings in the etcd space.
// The default splitter separates on ",".
func WithListSplitter(splitter list.Splitter) Option {
	return func(e *Getter) {
		e.listSplitter = splitter
	}
}
