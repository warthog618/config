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
	return func(g *Getter) {
		g.client = client
	}
}

// WithClientConfig sets the configuration of the etcd client.
// The default is an empty config with one endpoint - "localhost:2379".
func WithClientConfig(cfg clientv3.Config) Option {
	return func(g *Getter) {
		g.clientConfig = cfg
	}
}

// WithEndpoint sets the endpoint URL of the client port of the etcd server.
// The default endpoint is "localhost:2379".
func WithEndpoint(endpoints ...string) Option {
	return func(g *Getter) {
		g.clientConfig.Endpoints = endpoints
	}
}

// WithKeyReplacer sets the replacer used to map from etcd space to config space.
// The default replaces '/' in the etcd space with '.' in the config space.
func WithKeyReplacer(keyReplacer keys.Replacer) Option {
	return func(g *Getter) {
		g.keyReplacer = keyReplacer
	}
}

// WithListSplitter splits slice fields stored as strings in the etcd space.
// The default splitter separates on ",".
func WithListSplitter(splitter list.Splitter) Option {
	return func(g *Getter) {
		g.listSplitter = splitter
	}
}

// WithWatcher is an Option that enables watching of the etcd.
// This is an option to ensure it can only set at construction time, so the
// watcher is a singleton.
func WithWatcher() Option {
	return func(g *Getter) {
		g.w = &watcher{g: g}
	}
}
