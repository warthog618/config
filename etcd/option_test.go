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
	"github.com/warthog618/config/keys"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/warthog618/config/etcd"
	"github.com/warthog618/config/list"
)

func TestNewWithClient(t *testing.T) {
	_, cl, terminate := dummyEtcdServer(t, map[string]string{
		"/my/config/hello": "world",
	})
	defer terminate()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	e, err := etcd.New(ctx, "/my/config/", etcd.WithClient(cl), etcd.WithWatcher())
	cancel()
	assert.Nil(t, err)
	require.NotNil(t, e)
	v, ok := e.Get("hello")
	assert.True(t, ok)
	assert.Equal(t, "world", v)
}

func TestNewWithClientConfig(t *testing.T) {
	clus := integration.NewClusterV3(t,
		&integration.ClusterConfig{Size: 1})
	defer clus.Terminate(t)
	c := clus.RandClient()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	_, err := c.Put(ctx, "/my/config/hello", "world")
	assert.Nil(t, err)
	e, err := etcd.New(ctx, "/my/config/",
		etcd.WithClientConfig(clientv3.Config{
			Endpoints:        []string{clus.Members[0].GRPCAddr()},
			AutoSyncInterval: 5 * time.Minute,
		}))
	cancel()
	assert.Nil(t, err)
	require.NotNil(t, e)
	v, ok := e.Get("hello")
	assert.True(t, ok)
	assert.Equal(t, "world", v)
}

func TestNewWithEndpoint(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	e, err := etcd.New(ctx, "/my/config/", etcd.WithEndpoint())
	cancel()
	assert.NotNil(t, err)
	assert.Nil(t, e)
}

func TestNewWithKeyReplacer(t *testing.T) {
	patterns := []struct {
		name string
		k    string
		v    interface{}
		ok   bool
	}{
		{"leaf", "Leaf", "42", true},
		{"nested leaf", "Nested.Leaf", "44", true},
		{"nested nonsense", "Nested.nonsense", nil, false},
		{"nested slice", "Nested.Slice", []string{"c", "d"}, true},
		{"nested", "Nested", nil, false},
		{"nonsense", "nonsense", nil, false},
		{"slice", "Slice", []string{"a", "b"}, true},
		{"slice[]", "Slice[]", 2, true},
		{"slice[1]", "Slice[1]", "b", true},
		{"slice[3]", "Slice[3]", nil, false},
	}
	cfg := map[string]string{
		"/my/config/leaf":         "42",
		"/my/config/slice":        "a,b",
		"/my/config/nested.leaf":  "44",
		"/my/config/nested.slice": "c,d",
	}
	ep, _, terminate := dummyEtcdServer(t, cfg)
	defer terminate()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	e, err := etcd.New(ctx, "/my/config/", etcd.WithEndpoint(ep),
		etcd.WithKeyReplacer(keys.CamelCaseReplacer()))
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

func TestNewWithListSplitter(t *testing.T) {
	patterns := []struct {
		name string
		sep  string
		k    string
		v    interface{}
		ok   bool
	}{
		{"colon", ":", "colon", []string{"4", "2"}, true},
		{"comma", ",", "comma", []string{"a", "b"}, true},
		{"nested hash", "#", "nested.hash", []string{"4", "3"}, true},
		{"nested comma", ",", "nested.comma", []string{"c", "d"}, true},
		{"not comma", ":", "nested.comma", "c,d", true},
	}
	cfg := map[string]string{
		"/my/config/colon":        "4:2",
		"/my/config/comma":        "a,b",
		"/my/config/nested/hash":  "4#3",
		"/my/config/nested/comma": "c,d",
	}
	ep, _, terminate := dummyEtcdServer(t, cfg)
	defer terminate()
	for _, p := range patterns {
		f := func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			e, err := etcd.New(ctx, "/my/config/", etcd.WithEndpoint(ep),
				etcd.WithListSplitter(list.NewSplitter(p.sep)))
			cancel()
			assert.Nil(t, err)
			require.NotNil(t, e)
			v, ok := e.Get(p.k)
			assert.Equal(t, p.ok, ok)
			assert.Equal(t, p.v, v)
		}
		t.Run(p.name, f)
	}
}

func TestNewWithWatcher(t *testing.T) {
	addr, _, terminate := dummyEtcdServer(t, map[string]string{
		"/my/config/hello": "world",
	})
	defer terminate()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	e, err := etcd.New(ctx, "/my/config/", etcd.WithEndpoint(addr), etcd.WithWatcher())
	cancel()
	assert.Nil(t, err)
	require.NotNil(t, e)
	w, ok := e.Watcher()
	assert.True(t, ok)
	require.NotNil(t, w)
}
