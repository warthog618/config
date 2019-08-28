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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/warthog618/config"
	"github.com/warthog618/config/etcd"
)

var (
	defaultTimeout = 500 * time.Millisecond
	longTimeout    = 10 * time.Second
)

func TestNew(t *testing.T) {
	// no server
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	e, err := etcd.New(ctx, "/my/config/")
	cancel()
	assert.Equal(t, context.DeadlineExceeded, err)
	assert.Nil(t, e)
	assert.Implements(t, (*config.Getter)(nil), e)
	assert.Implements(t, (*config.WatchableGetter)(nil), e)

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

func TestNewWatcher(t *testing.T) {
	// also tests WithWatcher
	addr, cl, terminate := dummyEtcdServer(t, map[string]string{
		"/my/config/hello": "world",
	})
	defer terminate()
	ctx, cancel := context.WithTimeout(context.Background(), longTimeout)
	e, err := etcd.New(ctx, "/my/config/", etcd.WithEndpoint(addr))
	cancel()
	assert.Nil(t, err)
	require.NotNil(t, e)
	done := make(chan struct{})
	defer close(done)
	w := e.NewWatcher(done)
	assert.Nil(t, w)

	ctx, cancel = context.WithTimeout(context.Background(), longTimeout)
	e, err = etcd.New(ctx, "/my/config/", etcd.WithEndpoint(addr), etcd.WithWatcher())
	cancel()
	assert.Nil(t, err)
	require.NotNil(t, e)
	w = e.NewWatcher(done)
	testNotUpdated(t, w)
	cl.Put(context.Background(), "/my/config/hello", "final")
	testUpdated(t, w, nil)
}

func TestClose(t *testing.T) {
	addr, cl, terminate := dummyEtcdServer(t, map[string]string{
		"/my/config/hello": "world",
	})
	defer terminate()
	ctx, cancel := context.WithTimeout(context.Background(), longTimeout)
	e, err := etcd.New(ctx, "/my/config/", etcd.WithEndpoint(addr))
	cancel()
	assert.Nil(t, err)
	require.NotNil(t, e)
	done := make(chan struct{})
	defer close(done)
	w := e.NewWatcher(done)
	assert.Nil(t, w)
	err = e.Close()
	assert.Nil(t, err)

	ctx, cancel = context.WithTimeout(context.Background(), longTimeout)
	e, err = etcd.New(ctx, "/my/config/", etcd.WithEndpoint(addr), etcd.WithWatcher())
	cancel()
	assert.Nil(t, err)
	require.NotNil(t, e)
	w = e.NewWatcher(done)
	testNotUpdated(t, w)
	cl.Put(context.Background(), "/my/config/hello", "final")
	testUpdated(t, w, nil)
	err = e.Close()
	assert.Nil(t, err)
	testClosed(t, w.Update())
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

func TestGetterAsOption(t *testing.T) {
	addr, _, terminate := dummyEtcdServer(t, map[string]string{})
	defer terminate()
	ctx, cancel := context.WithTimeout(context.Background(), longTimeout)
	e, err := etcd.New(ctx, "/my/config/", etcd.WithEndpoint(addr))
	cancel()
	assert.Nil(t, err)
	c := config.NewConfig(e, e)
	c.Close()
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
	done := make(chan struct{})
	w := s.NewWatcher(done)
	cancel()

	// baseline
	testNotUpdated(t, w)

	// update
	cl.Put(context.Background(), "/my/config/leaf", "54")
	testUpdated(t, w, nil)

	// no content change, but still updated
	cl.Put(context.Background(), "/my/config/leaf", "54")
	testUpdated(t, w, nil)

	// closed so no update
	close(done)
	cl.Put(context.Background(), "/my/config/leaf", "54")
	testClosed(t, w.Update())
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
	done := make(chan struct{})
	w := s.NewWatcher(done)
	cancel()

	// baseline
	testNotUpdated(t, w)

	// update - put
	cl.Put(context.Background(), "/my/config/leaf", "updated")
	update := testUpdated(t, w, nil)
	require.NotNil(t, update)
	v, ok := s.Get("leaf")
	assert.True(t, ok)
	assert.Equal(t, "baseline", v)
	update.Commit()
	v, ok = s.Get("leaf")
	assert.True(t, ok)
	assert.Equal(t, "updated", v)
	ge, ok := update.(Getter)
	assert.True(t, ok)
	assert.Equal(t, s, ge.Getter())
	te, ok := update.(TemporaryError)
	assert.True(t, ok)
	assert.False(t, te.TemporaryError())

	// update - delete
	cl.Delete(context.Background(), "/my/config/leaf")
	update = testUpdated(t, w, nil)
	require.NotNil(t, update)
	v, ok = s.Get("leaf")
	assert.True(t, ok)
	assert.Equal(t, "updated", v)
	update.Commit()
	v, ok = s.Get("leaf")
	assert.False(t, ok)
	assert.Nil(t, v)

	// closed while updating
	cl.Put(context.Background(), "/my/config/leaf", "readded")
	time.Sleep(defaultTimeout)
	close(done)
	time.Sleep(defaultTimeout)
	testClosed(t, w.Update())
}

type Error interface {
	Err() error
}

type Getter interface {
	Getter() config.Getter
}
type TemporaryError interface {
	TemporaryError() bool
}

func dummyEtcdServer(t *testing.T, mss map[string]string) (string, *clientv3.Client, func()) {
	clus := integration.NewClusterV3(t,
		&integration.ClusterConfig{Size: 1})
	c := clus.RandClient()
	ctx, cancel := context.WithTimeout(context.Background(), longTimeout)
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

func testUpdated(t *testing.T, w config.GetterWatcher, xerr error) config.GetterUpdate {
	t.Helper()
	select {
	case update := <-w.Update():
		ue := update.(Error)
		assert.Equal(t, xerr, ue.Err())
		return update
	case <-time.After(time.Second):
		assert.Fail(t, "watch failed to return")
	}
	return nil
}

func testNotUpdated(t *testing.T, w config.GetterWatcher) {
	t.Helper()
	select {
	case update := <-w.Update():
		assert.Fail(t, "unexpected update", update)
	case <-time.After(defaultTimeout):
	}
}

func testClosed(t *testing.T, u <-chan config.GetterUpdate) {
	t.Helper()
	select {
	case update, ok := <-u:
		assert.False(t, ok)
		assert.Nil(t, update)
	case <-time.After(defaultTimeout):
	}
}
