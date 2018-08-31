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
	if g.w == nil {
		g.client.Close()
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
	// The watcher.
	w *watcher
}

type watcher struct {
	g *Getter
	// updated config... might actually be a set of ops on msi??
	events []*clientv3.Event
	// lastest uncommitted revision seen from etcd.
	rev int64
	// The channel providing update events from the etcd.
	wchan clientv3.WatchChan
}

// Close releases any resources allocated by the watcher.
// This implicitly closes any active Watches.
// After closing, the cached etcd configuration is still readable via Get, but
// will no longer be updated.
func (w *watcher) Close() (err error) {
	return w.g.client.Close()
}

// Get implements the Getter API.
func (g *Getter) Get(key string) (interface{}, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return tree.Get(g.msi, key, "")
}

// Watcher returns the watcher for the getter.
// The watcher is created at construction time if the Loader has a watcher.
// Returns nil if the getter is not watchable.
func (g *Getter) Watcher() config.GetterWatcher {
	return g.w
}

// Watch blocks until the etcd configuration changes.
// The Watch returns nil if the source has changed and no error was encountered.
// Otherwise the Watch returns the error encountered.
// The Watch may be cancelled by providing a ctx WithCancel and
// calling its cancel function.
// It is assumed that Watch and CommitUpdate will only be called from a single
// goroutine, and with CommitUpdate only called after a successful return from
// Watch.
func (w *watcher) Watch(ctx context.Context) error {
	if w.wchan == nil {
		ctx = clientv3.WithRequireLeader(ctx)
		w.wchan = w.g.client.Watch(
			context.Background(),
			w.g.prefix,
			clientv3.WithPrefix(),
			clientv3.WithRev(w.g.rev+1),
		)
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case ev, ok := <-w.wchan:
		if !ok {
			return context.Canceled
		}
		if ev.Err() != nil {
			return ev.Err()
		}
		w.events = ev.Events
		w.rev = ev.Header.Revision
		return nil
	}
}

func (w *watcher) CommitUpdate() {
	w.g.update(w.events, w.rev)
	w.events = nil
}

// update commits a change to the configuration detected by Watch, making
// the change visible to Get.
func (g *Getter) update(events []*clientv3.Event, rev int64) {
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
