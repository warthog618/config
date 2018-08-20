// Copyright © 2018 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package config_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/warthog618/config"
)

func TestNewSource(t *testing.T) {
	l := mockLoader{}
	d := mockDecoder{}

	// all good
	s, err := config.NewSource(&l, &d)
	assert.Nil(t, err)
	require.NotNil(t, s)

	dcerr := errors.New("decode error")
	lderr := errors.New("load error")

	// load error
	l = mockLoader{LoadError: lderr}
	d = mockDecoder{DecodeError: dcerr}
	s, err = config.NewSource(&l, &d)
	assert.Equal(t, lderr, err)
	require.Nil(t, s)

	// decode error
	l = mockLoader{}
	s, err = config.NewSource(&l, &d)
	assert.Equal(t, dcerr, err)
	require.Nil(t, s)
}

func TestNewSourceWithSeparator(t *testing.T) {
	l := mockLoader{}
	d := mockDecoder{M: map[string]interface{}{
		"a": map[string]interface{}{"b.c_d": true}}}
	s, err := config.NewSource(&l, &d, config.WithSeparator("-"))
	assert.Nil(t, err)
	require.NotNil(t, s)
	v, ok := s.Get("a.b.c_d")
	assert.False(t, ok)
	assert.Nil(t, v)
	v, ok = s.Get("a-b.c_d")
	assert.True(t, ok)
	assert.Equal(t, true, v)
}

func TestNewSourceClose(t *testing.T) {
	clerr := errors.New("close error")
	l := mockLoader{CloseError: clerr}
	d := mockDecoder{}
	s, err := config.NewSource(&l, &d)
	assert.Nil(t, err)
	require.NotNil(t, s)
	err = s.Close()
	assert.Equal(t, clerr, err)
	assert.True(t, l.Closed)
}

func TestSourceGet(t *testing.T) {
	l := mockLoader{}
	d := mockDecoder{M: map[string]interface{}{
		"a": map[string]interface{}{"b.c_d": true}}}
	s, err := config.NewSource(&l, &d)
	assert.Nil(t, err)
	require.NotNil(t, s)
	v, ok := s.Get("")
	assert.False(t, ok)
	assert.Nil(t, v)
	v, ok = s.Get("a.b.c_d")
	assert.True(t, ok)
	assert.Equal(t, true, v)
}

func TestSourceWatch(t *testing.T) {
	l := mockLoader{N: make(chan struct{})}
	d := mockDecoder{M: map[string]interface{}{"a.b.c_d": "baseline"}}
	s, err := config.NewSource(&bareLoader{}, &d)
	assert.Nil(t, err)
	require.NotNil(t, s)
	// unwatchable
	testWatcher(t, s, config.ErrUnwatchable)

	s, err = config.NewSource(&l, &d)
	assert.Nil(t, err)
	require.NotNil(t, s)
	// baseline
	testWatcher(t, s, context.DeadlineExceeded)

	// update
	d.M = map[string]interface{}{"a.b.c_d": "updated"}
	l.Update()
	testWatcher(t, s, nil)

	// no content change => no return
	d.M = map[string]interface{}{"a.b.c_d": "baseline"}
	l.Update()
	testWatcher(t, s, context.DeadlineExceeded)

	// bad load
	d.M = map[string]interface{}{"a.b.c_d": "final"}
	d.DecodeError = errors.New("Decode error")
	l.Update()
	testWatcher(t, s, context.DeadlineExceeded)
}

func TestSourceCommitUpdate(t *testing.T) {
	l := mockLoader{N: make(chan struct{})}
	d := mockDecoder{M: map[string]interface{}{"a.b.c_d": "baseline"}}
	s, err := config.NewSource(&l, &d)
	assert.Nil(t, err)
	require.NotNil(t, s)

	// baseline
	testWatcher(t, s, context.DeadlineExceeded)

	// update
	d.M = map[string]interface{}{"a.b.c_d": "updated"}
	l.Update()
	testWatcher(t, s, nil)
	v, ok := s.Get("a.b.c_d")
	assert.True(t, ok)
	assert.Equal(t, "baseline", v)
	s.CommitUpdate()
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

func (l *mockLoader) Close() error {
	l.Closed = true
	return l.CloseError
}

func (l *mockLoader) Load() ([]byte, error) {
	return l.B, l.LoadError
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

func (l *mockLoader) Update() {
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
