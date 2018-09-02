// Copyright © 2018 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package blob_test

import (
	"context"
	"testing"
	"time"

	"github.com/pkg/errors"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/warthog618/config/blob"
)

func TestNew(t *testing.T) {
	l := mockLoader{}
	d := mockDecoder{}

	// all good
	s, err := blob.New(&l, &d)
	assert.Nil(t, err)
	require.NotNil(t, s)

	dcerr := errors.New("decode error")
	lderr := errors.New("load error")

	// load error
	l = mockLoader{LoadError: lderr}
	d = mockDecoder{DecodeError: dcerr}
	s, err = blob.New(&l, &d)
	assert.Equal(t, lderr, err)
	require.Nil(t, s)

	// decode error
	l = mockLoader{}
	s, err = blob.New(&l, &d)
	assert.Equal(t, dcerr, err)
	require.Nil(t, s)
}

func TestWatcher(t *testing.T) {
	l := mockLoader{}
	d := mockDecoder{}

	// all good
	s, err := blob.New(&l, &d)
	assert.Nil(t, err)
	require.NotNil(t, s)
	w, ok := s.Watcher()
	assert.True(t, ok)
	require.NotNil(t, w)

	// not watchable
	s, err = blob.New(&bareLoader{}, &d)
	assert.Nil(t, err)
	require.NotNil(t, s)
	w, ok = s.Watcher()
	assert.False(t, ok)
	require.Nil(t, w)
}

func TestWatcherClose(t *testing.T) {
	clerr := errors.New("close error")
	l := mockLoader{CloseError: clerr}
	d := mockDecoder{}
	s, err := blob.New(&l, &d)
	assert.Nil(t, err)
	require.NotNil(t, s)
	w, ok := s.Watcher()
	assert.True(t, ok)
	require.NotNil(t, w)
	err = w.Close()
	assert.Equal(t, clerr, err)
	assert.True(t, l.Closed)
}

func TestGet(t *testing.T) {
	l := mockLoader{}
	d := mockDecoder{M: map[string]interface{}{
		"a": map[string]interface{}{"b.c_d": true}}}
	s, err := blob.New(&l, &d)
	assert.Nil(t, err)
	require.NotNil(t, s)
	v, ok := s.Get("")
	assert.False(t, ok)
	assert.Nil(t, v)
	v, ok = s.Get("a.b.c_d")
	assert.True(t, ok)
	assert.Equal(t, true, v)
}

func TestWatch(t *testing.T) {
	l := mockLoader{N: make(chan struct{})}
	d := mockDecoder{M: map[string]interface{}{"a.b.c_d": "baseline"}}
	s, err := blob.New(&l, &d)
	assert.Nil(t, err)
	require.NotNil(t, s)
	w, ok := s.Watcher()
	assert.True(t, ok)
	require.NotNil(t, s)
	// baseline
	testWatcher(t, w, context.DeadlineExceeded)

	// update
	d.M = map[string]interface{}{"a.b.c_d": "updated"}
	l.Modify()
	testWatcher(t, w, nil)

	// no content change => no return
	d.M = map[string]interface{}{"a.b.c_d": "baseline"}
	l.Modify()
	testWatcher(t, w, context.DeadlineExceeded)

	// bad load
	d.M = map[string]interface{}{"a.b.c_d": "final"}
	d.DecodeError = errors.New("Decode error")
	l.Modify()
	testWatcher(t, w, d.DecodeError)

	// pathological decoder
	d.M = nil
	d.DecodeError = nil
	l.Modify()
	testWatcher(t, w, context.DeadlineExceeded)
}

func TestUpdate(t *testing.T) {
	l := mockLoader{N: make(chan struct{})}
	d := mockDecoder{M: map[string]interface{}{"a.b.c_d": "baseline"}}
	s, err := blob.New(&l, &d)
	assert.Nil(t, err)
	require.NotNil(t, s)
	w, ok := s.Watcher()
	assert.True(t, ok)
	require.NotNil(t, w)

	// baseline
	testWatcher(t, w, context.DeadlineExceeded)

	// update
	d.M = map[string]interface{}{"a.b.c_d": "updated"}
	l.Modify()
	testWatcher(t, w, nil)
	v, ok := s.Get("a.b.c_d")
	assert.True(t, ok)
	assert.Equal(t, "baseline", v)
	w.CommitUpdate()
	v, ok = s.Get("a.b.c_d")
	assert.True(t, ok)
	assert.Equal(t, "updated", v)
}

type bareLoader struct{}

func (l *bareLoader) Load() ([]byte, error) {
	return nil, nil
}

type mockLoader struct {
	B          []byte
	LoadError  error
	Closed     bool
	CloseError error
	N          chan struct{}
}

func (l *mockLoader) Load() ([]byte, error) {
	return l.B, l.LoadError
}

func (l *mockLoader) Watcher() (blob.WatcherCloser, bool) {
	return l, true
}

func (l *mockLoader) Close() error {
	l.Closed = true
	return l.CloseError
}

func (l *mockLoader) Watch(ctx context.Context) error {
	select {
	case <-l.N:
		l.N = make(chan struct{})
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (l *mockLoader) Modify() {
	close(l.N)
}

type mockDecoder struct {
	M           map[string]interface{}
	DecodeError error
}

func (d *mockDecoder) Decode(b []byte, v interface{}) error {
	m := v.(*map[string]interface{})
	*m = d.M
	return d.DecodeError
}

type watcher interface {
	Watch(context.Context) error
}

func testWatcher(t *testing.T, w watcher, xerr error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
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
