// Copyright © 2018 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package etcd_test

import (
	"context"
	"testing"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/integration"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/warthog618/config/etcd"
)

var (
	defaultTimeout = 100 * time.Millisecond
	longTimeout    = 10 * time.Second
)

func TestNew(t *testing.T) {
	// no server
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	e, err := etcd.New(ctx, "/my/config/")
	cancel()
	assert.Equal(t, context.DeadlineExceeded, err)
	assert.Nil(t, e)

	// no endpoint
	ctx, cancel = context.WithTimeout(context.Background(), defaultTimeout)
	e, err = etcd.New(ctx, "/my/config/", etcd.WithEndpoint())
	cancel()
	assert.NotNil(t, err)
	assert.Nil(t, e)

	// real server
	addr, _, terminate := dummyEtcdServer(t, map[string]string{
		"/my/config/hello": "world",
	})
	defer terminate()
	ctx, cancel = context.WithTimeout(context.Background(), longTimeout)
	e, err = etcd.New(ctx, "/my/config/", etcd.WithEndpoint(addr))
	cancel()
	assert.Nil(t, err)
	require.NotNil(t, e)
	v, ok := e.Get("hello")
	assert.True(t, ok)

	assert.Equal(t, "world", v)
}

func TestWatcher(t *testing.T) {
	addr, _, terminate := dummyEtcdServer(t, map[string]string{
		"/my/config/hello": "world",
	})
	defer terminate()
	ctx, cancel := context.WithTimeout(context.Background(), longTimeout)
	e, err := etcd.New(ctx, "/my/config/", etcd.WithEndpoint(addr))
	cancel()
	assert.Nil(t, err)
	require.NotNil(t, e)
	w, ok := e.Watcher()
	assert.False(t, ok)
	require.Nil(t, w)
	ctx, cancel = context.WithTimeout(context.Background(), longTimeout)
	e, err = etcd.New(ctx, "/my/config/", etcd.WithEndpoint(addr), etcd.WithWatcher())
	cancel()
	assert.Nil(t, err)
	require.NotNil(t, e)
	w, ok = e.Watcher()
	assert.True(t, ok)
	require.NotNil(t, w)
}

func TestWatcherClose(t *testing.T) {
	addr, cl, terminate := dummyEtcdServer(t, map[string]string{
		"/my/config/hello": "world",
	})
	defer terminate()
	ctx, cancel := context.WithTimeout(context.Background(), longTimeout)
	e, err := etcd.New(ctx, "/my/config/", etcd.WithEndpoint(addr), etcd.WithWatcher())
	cancel()
	assert.Nil(t, err)
	require.NotNil(t, e)
	w, ok := e.Watcher()
	assert.True(t, ok)
	require.NotNil(t, w)
	v, ok := e.Get("hello")
	assert.True(t, ok)
	assert.Equal(t, "world", v)

	cl.Put(context.Background(), "/my/config/hello", "updated")
	testWatcher(t, w, nil)
	w.CommitUpdate()
	v, ok = e.Get("hello")
	assert.True(t, ok)
	assert.Equal(t, "updated", v)

	w.Close()

	cl.Put(context.Background(), "/my/config/hello", "final")
	testWatcher(t, w, context.Canceled)
	v, ok = e.Get("hello")
	assert.True(t, ok)
	assert.Equal(t, "updated", v)
}

func TestGet(t *testing.T) {
	patterns := []struct {
		name string
		k    string
		v    interface{}
		ok   bool
	}{
		{"leaf", "leaf", "42", true},
		{"nested leaf", "nested.leaf", "44", true},
		{"nested nonsense", "nested.nonsense", nil, false},
		{"nested slice", "nested.slice", []string{"c", "d"}, true},
		{"nested", "nested", nil, false},
		{"nonsense", "nonsense", nil, false},
		{"slice", "slice", []string{"a", "b"}, true},
		{"slice[]", "slice[]", 2, true},
		{"slice[1]", "slice[1]", "b", true},
		{"slice[3]", "slice[3]", nil, false},
	}
	cfg := map[string]string{
		"/my/config/leaf":         "42",
		"/my/config/slice":        "a,b",
		"/my/config/nested/leaf":  "44",
		"/my/config/nested/slice": "c,d",
	}
	addr, _, terminate := dummyEtcdServer(t, cfg)
	defer terminate()
	ctx, cancel := context.WithTimeout(context.Background(), longTimeout)
	e, err := etcd.New(ctx, "/my/config/", etcd.WithEndpoint(addr))
	cancel()
	assert.Nil(t, err)
	require.NotNil(t, e)

	for _, p := range patterns {
		f := func(t *testing.T) {
			v, ok := e.Get(p.k)
			assert.Equal(t, p.ok, ok)
			assert.Equal(t, p.v, v)
		}
		t.Run(p.name, f)
	}
}

func TestWatch(t *testing.T) {
	cfg := map[string]string{
		"/my/config/leaf": "42",
	}
	addr, cl, terminate := dummyEtcdServer(t, cfg)
	defer terminate()
	ctx, cancel := context.WithTimeout(context.Background(), longTimeout)
	s, err := etcd.New(ctx, "/my/config/", etcd.WithEndpoint(addr), etcd.WithWatcher())
	assert.Nil(t, err)
	require.NotNil(t, s)
	w, ok := s.Watcher()
	assert.True(t, ok)
	require.NotNil(t, w)
	cancel()

	// baseline
	testWatcher(t, w, context.DeadlineExceeded)

	// update
	cl.Put(context.Background(), "/my/config/leaf", "54")
	testWatcher(t, w, nil)

	// no content change, but still updated
	cl.Put(context.Background(), "/my/config/leaf", "54")
	testWatcher(t, w, nil)

	// closed so no update
	w.Close()
	cl.Put(context.Background(), "/my/config/leaf", "54")
	testWatcher(t, w, context.Canceled)
}

func TestUpdate(t *testing.T) {
	cfg := map[string]string{
		"/my/config/leaf": "baseline",
	}
	addr, cl, terminate := dummyEtcdServer(t, cfg)
	defer terminate()
	ctx, cancel := context.WithTimeout(context.Background(), longTimeout)
	s, err := etcd.New(ctx, "/my/config/", etcd.WithEndpoint(addr), etcd.WithWatcher())
	assert.Nil(t, err)
	require.NotNil(t, s)
	w, ok := s.Watcher()
	assert.True(t, ok)
	require.NotNil(t, w)
	cancel()

	// baseline
	testWatcher(t, w, context.DeadlineExceeded)

	// update - put
	cl.Put(context.Background(), "/my/config/leaf", "updated")
	testWatcher(t, w, nil)
	v, ok := s.Get("leaf")
	assert.True(t, ok)
	assert.Equal(t, "baseline", v)
	w.CommitUpdate()
	v, ok = s.Get("leaf")
	assert.True(t, ok)
	assert.Equal(t, "updated", v)

	// update - delete
	cl.Delete(context.Background(), "/my/config/leaf")
	testWatcher(t, w, nil)
	v, ok = s.Get("leaf")
	assert.True(t, ok)
	assert.Equal(t, "updated", v)
	w.CommitUpdate()
	v, ok = s.Get("leaf")
	assert.False(t, ok)
	assert.Nil(t, v)

}

func dummyEtcdServer(t *testing.T, mss map[string]string) (string, *clientv3.Client, func()) {
	clus := integration.NewClusterV3(t,
		&integration.ClusterConfig{Size: 1})
	c := clus.RandClient()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	for k, v := range mss {
		_, err := c.Put(ctx, k, v)
		if err != nil {
			t.Fatal(err)
		}
	}
	cancel()
	return clus.Members[0].GRPCAddr(),
		c,
		func() {
			clus.Terminate(t)
		}
}

type watcher interface {
	Watch(context.Context) error
}

func testWatcher(t *testing.T, w watcher, xerr error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	updated := make(chan error)
	go func() {
		err := w.Watch(ctx)
		updated <- err
	}()
	select {
	case err := <-updated:
		assert.Equal(t, xerr, errors.Cause(err))
	case <-time.After(time.Second):
		assert.Fail(t, "watch failed to return")
	}
	cancel()
}
