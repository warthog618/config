// Copyright © 2018 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package etcd provides a config getter for etcd v3 key/value stores.
package etcd

import (
	"context"
	"strings"
	"sync"

	"github.com/coreos/etcd/clientv3"
	"github.com/warthog618/config"
	"github.com/warthog618/config/keys"
	"github.com/warthog618/config/list"
	"github.com/warthog618/config/tree"
)

// New creates a new etcd getter for the given prefix, and loads the
// configuration.
// The prefix defines the root of the configuration in the etcd space.
// The ctx covers the initial loading of configuration.
func New(ctx context.Context, prefix string, options ...Option) (*Getter, error) {
	g := Getter{
		clientConfig: clientv3.Config{Endpoints: []string{"localhost:2379"}},
		prefix:       prefix,
	}
	for _, option := range options {
		option(&g)
	}
	if g.keyReplacer == nil {
		g.keyReplacer = keys.StringReplacer("/", ".")
	}
	if g.listSplitter == nil {
		g.listSplitter = list.NewSplitter(",")
	}
	if g.client == nil {
		client, err := clientv3.New(g.clientConfig)
		if err != nil {
			return nil, err
		}
		g.client = client
	}
	ctx = clientv3.WithRequireLeader(ctx)
	msi, err := g.load(ctx)
	if err != nil {
		g.client.Close()
		return nil, err
	}
	g.msi = msi
	if !g.watcher {
		g.client.Close()
		g.client = nil
	}
	return &g, nil
}

// Getter represents a getter from an etcd v3 key/value store.
// It is assumed that the relevant configuration is located within a section
// of the etcd keyspace with a fixed key prefix, e.g. /my/app/config/.
type Getter struct {
	mu sync.RWMutex
	// The current snapshot of configuration loaded from etcd.
	msi map[string]interface{}
	// lastest revision committed from etcd.
	rev int64
	// The configuration for the etcd client.
	clientConfig clientv3.Config
	// The etcd client.
	client *clientv3.Client
	// A replacer that maps from etcd space to config space.
	keyReplacer keys.Replacer
	// The splitter for slices stored in string values.
	listSplitter list.Splitter
	// prefix defines the root of the configuration in the etcd space.
	// Is defined in the etcd space, and must include any trailing separator.
	prefix string
	// watcher enable flag.
	watcher bool
}

// Close closes the getters connection to the server.
// This automatically closes any active watchers.
// The current configuration snapshot can still be accessed via Get.
func (g *Getter) Close() error {
	g.mu.Lock()
	if g.client != nil {
		g.client.Close()
	}
	g.mu.Unlock()
	return nil
}

// Get implements the Getter API.
func (g *Getter) Get(key string) (interface{}, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return tree.Get(g.msi, key, "")
}

// NewWatcher creates a watcher goroutine if watcher was enabled during construction.
func (g *Getter) NewWatcher(done <-chan struct{}) config.GetterWatcher {
	if g.watcher {
		ctx := clientv3.WithRequireLeader(context.Background())
		g.mu.RLock()
		wchan := g.client.Watch(
			ctx,
			g.prefix,
			clientv3.WithPrefix(),
			clientv3.WithRev(g.rev+1),
		)
		g.mu.RUnlock()
		gw := &getterWatcher{uch: make(chan config.GetterUpdate)}
		go g.watch(done, wchan, gw)
		return gw
	}
	return nil
}

func (g *Getter) watch(done <-chan struct{}, wchan clientv3.WatchChan, gw *getterWatcher) {
	defer close(gw.uch)
	send := func(u getterUpdate) {
		select {
		case gw.uch <- u:
		case <-done:
			return
		}
	}
	for {
		select {
		case <-done:
			return
		case ev, ok := <-wchan:
			if !ok {
				return
			}
			if ev.Err() != nil {
				// need more info on event types...
				send(getterUpdate{g: g, err: ev.Err()})
				continue
			}
			send(getterUpdate{g: g, commit: func() { g.commit(ev.Events, ev.Header.Revision) }})
		}
	}
}

// commit commits a change to the configuration detected by Watch, making
// the change visible to Get.
func (g *Getter) commit(events []*clientv3.Event, rev int64) {
	g.mu.Lock()
	for _, ev := range events {
		key := string(ev.Kv.Key)
		if !strings.HasPrefix(key, g.prefix) {
			continue
		}
		key = g.keyReplacer.Replace(key[len(g.prefix):])
		switch ev.Type {
		case clientv3.EventTypeDelete:
			delete(g.msi, key)
		default:
			g.msi[key] = g.listSplitter.Split(string(ev.Kv.Value))
		}
	}
	g.rev = rev
	g.mu.Unlock()
}

func (g *Getter) load(ctx context.Context) (map[string]interface{}, error) {
	x, err := g.client.Get(ctx, g.prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	g.rev = x.Header.Revision
	msi := make(map[string]interface{})
	for _, kv := range x.Kvs {
		key := string(kv.Key)
		if !strings.HasPrefix(key, g.prefix) {
			continue
		}
		key = g.keyReplacer.Replace(key[len(g.prefix):])
		msi[key] = g.listSplitter.Split(string(kv.Value))
	}
	return msi, nil
}

type getterWatcher struct {
	uch chan config.GetterUpdate
}

func (g *getterWatcher) Update() <-chan config.GetterUpdate {
	return g.uch
}

type getterUpdate struct {
	g       config.Getter
	err     error
	commit  func()
	temperr bool
}

func (g getterUpdate) Getter() config.Getter {
	return g.g
}

func (g getterUpdate) Err() error {
	return g.err
}

func (g getterUpdate) TemporaryError() bool {
	return g.temperr
}

func (g getterUpdate) Commit() {
	if g.commit == nil {
		return
	}
	g.commit()
}
