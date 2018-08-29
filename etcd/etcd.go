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
	"github.com/warthog618/config/keys"
	"github.com/warthog618/config/list"
	"github.com/warthog618/config/tree"
)

// New creates a new etcd getter for the given prefix, and loads the
// configuration.
// The prefix defines the root of the configuration in the etcd space.
// The ctx covers the initial loading of configuration.
func New(ctx context.Context, prefix string, options ...Option) (*Etcd, error) {
	e := Etcd{
		clientConfig: clientv3.Config{Endpoints: []string{"localhost:2379"}},
		prefix:       prefix,
	}
	for _, option := range options {
		option(&e)
	}
	if e.keyReplacer == nil {
		e.keyReplacer = keys.StringReplacer("/", ".")
	}
	if e.listSplitter == nil {
		e.listSplitter = list.NewSplitter(",")
	}
	if e.client == nil {
		client, err := clientv3.New(e.clientConfig)
		if err != nil {
			return nil, err
		}
		e.client = client
	}
	ctx = clientv3.WithRequireLeader(ctx)
	msi, err := e.load(ctx)
	if err != nil {
		e.Close()
		return nil, err
	}
	e.msi = msi
	return &e, nil
}

// Etcd represents a getter from an etcd v3 key/value store.
// It is assumed that the relevant configuration is located within a section
// of the etcd keyspace with a fixed key prefix, e.g. /my/app/config/.
type Etcd struct {
	mu sync.RWMutex
	// The current snapshot of configuration loaded from etcd.
	msi map[string]interface{}
	// lastest revision commited from etcd.
	msirev int64
	// updated config... might actually be a set of ops on msi??
	events []*clientv3.Event
	// lastest uncommitted revision seen from etcd.
	eventsrev int64
	// The configuration for the etcd client.
	clientConfig clientv3.Config
	// The etcd client.
	client *clientv3.Client
	// A replacer that maps from etcd space to config space.
	keyReplacer Replacer
	// The splitter for slices stored in string values.
	listSplitter list.Splitter
	// prefix defines the root of the configuration in the etcd space.
	// Is defined in the etcd space, and must include any trailing separator.
	prefix string
	// The channel providing update events from the etcd.
	wchan clientv3.WatchChan
}

// Replacer maps a key from one space to another.
type Replacer interface {
	Replace(string) string
}

// Close releases any resources allocated by the etcd.
// This implicitly closes any active Watches.
// After closing, the cached etcd configuration is still readable via Get, but
// will no longer be updated.
func (e *Etcd) Close() (err error) {
	return e.client.Close()
}

// Get implements the Getter API.
func (e *Etcd) Get(key string) (interface{}, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return tree.Get(e.msi, key, "")
}

// Watch blocks until the etcd configuration changes.
// The Watch returns nil if the source has changed and no error was encountered.
// Otherwise the Watch returns the error encountered.
// The Watch may be cancelled by providing a ctx WithCancel and
// calling its cancel function.
// It is assumed that Watch and CommitUpdate will only be called from a single
// goroutine, and with CommitUpdate only called after a successful return from
// Watch.
func (e *Etcd) Watch(ctx context.Context) error {
	if e.wchan == nil {
		ctx = clientv3.WithRequireLeader(ctx)
		e.wchan = e.client.Watch(
			context.Background(),
			e.prefix,
			clientv3.WithPrefix(),
			clientv3.WithRev(e.msirev+1),
		)
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case ev, ok := <-e.wchan:
		if !ok {
			return context.Canceled
		}
		if ev.Err() != nil {
			return ev.Err()
		}
		e.events = ev.Events
		e.eventsrev = ev.Header.Revision
		return nil
	}
}

// CommitUpdate commits a change to the configuration detected by Watch, making
// the change visible to Get.
// It is assumed that Watch and CommitUpdate will only be called from a single
// goroutine, and with CommitUpdate only called after a successful return from
// Watch.
func (e *Etcd) CommitUpdate() {
	e.mu.Lock()
	for _, ev := range e.events {
		key := string(ev.Kv.Key)
		if !strings.HasPrefix(key, e.prefix) {
			continue
		}
		key = e.keyReplacer.Replace(key[len(e.prefix):])
		switch ev.Type {
		case clientv3.EventTypeDelete:
			delete(e.msi, key)
		default:
			e.msi[key] = e.listSplitter.Split(string(ev.Kv.Value))
		}
	}
	e.msirev = e.eventsrev
	e.events = nil
	e.mu.Unlock()
}

func (e *Etcd) load(ctx context.Context) (map[string]interface{}, error) {
	x, err := e.client.Get(ctx, e.prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	e.msirev = x.Header.Revision
	msi := make(map[string]interface{})
	for _, kv := range x.Kvs {
		key := string(kv.Key)
		if !strings.HasPrefix(key, e.prefix) {
			continue
		}
		key = e.keyReplacer.Replace(key[len(e.prefix):])
		msi[key] = e.listSplitter.Split(string(kv.Value))
	}
	return msi, nil
}
