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
	"github.com/warthog618/config/loader/file"
)

func TestNew(t *testing.T) {
	f := file.New("nosuch.file")
	assert.NotNil(t, f)
}

func TestFileLoad(t *testing.T) {
	// Existent
	f := file.New("file_test.go")
	require.NotNil(t, f)
	l, err := f.Load()
	assert.Nil(t, err)
	assert.NotNil(t, l)

	// Non-existent
	f = file.New("nosuch.file")
	require.NotNil(t, f)
	l, err = f.Load()
	assert.IsType(t, &os.PathError{}, err)
	assert.Nil(t, l)
}

func TestNewWatchedFile(t *testing.T) {
	// Non-existent
	f, err := file.NewWatchedFile("nosuch.file")
	assert.NotNil(t, err)
	assert.Nil(t, f)

	// Existing
	f, err = file.NewWatchedFile("file_test.go")
	assert.Nil(t, err)
	assert.NotNil(t, f)
}

func TestWatchedFileLoad(t *testing.T) {
	// Existent
	wf, err := file.NewWatchedFile("file_test.go")
	assert.Nil(t, err)
	require.NotNil(t, wf)
	l, err := wf.Load()
	assert.Nil(t, err)
	assert.NotNil(t, l)

	// Non-existent
	tf, err := ioutil.TempFile(".", "file_test_")
	assert.Nil(t, err)
	require.NotNil(t, tf)
	fname := tf.Name()
	wf, err = file.NewWatchedFile(fname)
	assert.Nil(t, err)
	require.NotNil(t, wf)
	tf.Close()
	os.Remove(fname)
	l, err = wf.Load()
	assert.NotNil(t, err)
	assert.Nil(t, l)
}

func TestWatchedFileClose(t *testing.T) {
	f, err := ioutil.TempFile(".", "file_test_")
	assert.Nil(t, err)
	require.NotNil(t, f)
	fname := f.Name()
	defer os.Remove(fname)
	wf, err := file.NewWatchedFile(fname)
	assert.Nil(t, err)
	require.NotNil(t, wf)
	wf.Close()
	err = wf.Watch(context.Background())
	assert.Equal(t, context.Canceled, err)
}

func testWatch(t *testing.T, w *file.WatchedFile) {
	t.Helper()
}

func TestWatchedFileWatch(t *testing.T) {
	// Write
	tf, err := ioutil.TempFile(".", "file_test_")
	assert.Nil(t, err)
	require.NotNil(t, tf)
	fname := tf.Name()
	wf, err := file.NewWatchedFile(fname)
	assert.Nil(t, err)
	require.NotNil(t, wf)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error)
	go func() {
		err := wf.Watch(ctx)
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
		err := wf.Watch(ctx)
		done <- err
	}()
	wf.Close()
	select {
	case err := <-done:
		assert.Equal(t, context.Canceled, err)
	case <-time.After(time.Second):
		assert.Fail(t, "watch didn't return")
	}

	// Remove
	wf, err = file.NewWatchedFile(fname)
	assert.Nil(t, err)
	require.NotNil(t, wf)
	done = make(chan error)
	go func() {
		err := wf.Watch(ctx)
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
	wf, err = file.NewWatchedFile(fname)
	assert.Nil(t, err)
	require.NotNil(t, wf)
	ctx, cancel = context.WithCancel(context.Background())
	done = make(chan error)
	go func() {
		err := wf.Watch(ctx)
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
		err := wf.Watch(ctx)
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
