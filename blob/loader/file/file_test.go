// Copyright © 2018 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package file_test

import (
	"context"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/warthog618/config/blob/loader/file"
)

func TestNew(t *testing.T) {
	// Existent
	f, err := file.New("file_test.go")
	assert.Nil(t, err)
	require.NotNil(t, f)

	// Non-existent
	f, err = file.New("nosuch.file")
	assert.Nil(t, err)
	require.NotNil(t, f)
}

func TestLoad(t *testing.T) {
	// Existent
	f, err := file.New("file_test.go")
	assert.Nil(t, err)
	require.NotNil(t, f)
	l, err := f.Load()
	assert.Nil(t, err)
	assert.NotNil(t, l)

	// Non-existent
	f, err = file.New("nosuch.file")
	assert.Nil(t, err)
	require.NotNil(t, f)
	l, err = f.Load()
	assert.IsType(t, &os.PathError{}, err)
	assert.Nil(t, l)
}

func TestWatcher(t *testing.T) {
	// Non-existent
	f, err := file.New("nosuch.file", file.WithWatcher())
	assert.NotNil(t, err)
	assert.Nil(t, f)

	// Existent
	f, err = file.New("file_test.go", file.WithWatcher())
	assert.Nil(t, err)
	require.NotNil(t, f)
	w := f.Watcher()
	assert.NotNil(t, w)
}

func TestClose(t *testing.T) {
	f, err := ioutil.TempFile(".", "file_test_")
	assert.Nil(t, err)
	require.NotNil(t, f)
	fname := f.Name()
	defer os.Remove(fname)
	wf, err := file.New(fname, file.WithWatcher())
	assert.Nil(t, err)
	require.NotNil(t, wf)
	w := wf.Watcher()
	require.NotNil(t, w)
	w.Close()
	err = w.Watch(context.Background())
	assert.Equal(t, context.Canceled, err)
}

func TestWatcherWatch(t *testing.T) {
	// Write
	tf, err := ioutil.TempFile(".", "file_test_")
	assert.Nil(t, err)
	require.NotNil(t, tf)
	fname := tf.Name()
	wf, err := file.New(fname, file.WithWatcher())
	assert.Nil(t, err)
	require.NotNil(t, wf)
	w := wf.Watcher()
	require.NotNil(t, w)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error)
	go func() {
		err := w.Watch(ctx)
		done <- err
	}()
	tf.Write([]byte("test pattern"))
	select {
	case err := <-done:
		assert.Nil(t, err)
	case <-time.After(time.Second):
		assert.Fail(t, "watch didn't return")
	}

	// Close
	done = make(chan error)
	go func() {
		err := w.Watch(ctx)
		done <- err
	}()
	w.Close()
	select {
	case err := <-done:
		assert.Equal(t, context.Canceled, err)
	case <-time.After(time.Second):
		assert.Fail(t, "watch didn't return")
	}

	// Remove
	wf, err = file.New(fname, file.WithWatcher())
	assert.Nil(t, err)
	require.NotNil(t, wf)
	w = wf.Watcher()
	require.NotNil(t, w)
	done = make(chan error)
	go func() {
		err := w.Watch(ctx)
		done <- err
	}()
	tf.Close()
	os.Remove(fname)
	select {
	case err := <-done:
		assert.Nil(t, err)
	case <-time.After(time.Second):
		assert.Fail(t, "watch didn't return")
	}
	cancel()

	// Rename
	tf, err = ioutil.TempFile(".", "file_test_")
	assert.Nil(t, err)
	require.NotNil(t, tf)
	fname = tf.Name()
	wf, err = file.New(fname, file.WithWatcher())
	assert.Nil(t, err)
	require.NotNil(t, wf)
	w = wf.Watcher()
	require.NotNil(t, w)
	ctx, cancel = context.WithCancel(context.Background())
	done = make(chan error)
	go func() {
		err := w.Watch(ctx)
		done <- err
	}()
	os.Rename(fname, fname+"r")
	select {
	case err := <-done:
		assert.Nil(t, err)
	case <-time.After(time.Second):
		assert.Fail(t, "watch didn't return")
	}

	// Cancel
	done = make(chan error)
	go func() {
		err := w.Watch(ctx)
		done <- err
	}()
	cancel()
	select {
	case err := <-done:
		assert.Equal(t, context.Canceled, err)
	case <-time.After(time.Second):
		assert.Fail(t, "watch didn't return")
	}
	os.Remove(fname + "r")
}
