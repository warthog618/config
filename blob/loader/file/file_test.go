// Copyright © 2018 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package file_test

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/warthog618/config/blob"
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
	assert.Implements(t, (*blob.WatchableLoader)(nil), f)
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

var defaultTimeout = time.Millisecond

func TestWatcherClose(t *testing.T) {
	f, err := ioutil.TempFile(".", "file_test_")
	assert.Nil(t, err)
	require.NotNil(t, f)
	fname := f.Name()
	defer os.Remove(fname)
	wf, err := file.New(fname, file.WithWatcher())
	assert.Nil(t, err)
	require.NotNil(t, wf)
	done := make(chan struct{})
	wchan := wf.NewWatcher(done)
	require.NotNil(t, wchan)
	// immediate update to trigger load
	select {
	case err, ok := <-wchan:
		assert.True(t, ok)
		assert.Nil(t, err)
	case <-time.After(defaultTimeout):
		assert.Fail(t, "watch didn't return")
	}
	close(done)
	select {
	case err, ok := <-wchan:
		assert.False(t, ok)
		assert.Nil(t, err)
	case <-time.After(time.Second):
		assert.Fail(t, "watch did not terminate")
	}
}

func TestWatcher(t *testing.T) {
	tf, err := ioutil.TempFile(".", "file_test_")
	assert.Nil(t, err)
	require.NotNil(t, tf)
	fname := tf.Name()
	defer os.Remove(fname)

	// Not watched
	wf, err := file.New(fname)
	assert.Nil(t, err)
	require.NotNil(t, wf)
	done := make(chan struct{})
	defer close(done)
	wchan := wf.NewWatcher(done)
	require.Nil(t, wchan)

	// Watched
	wf, err = file.New(fname, file.WithWatcher())
	assert.Nil(t, err)
	require.NotNil(t, wf)
	wchan = wf.NewWatcher(done)
	require.NotNil(t, wchan)
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
	done := make(chan struct{})
	wchan := wf.NewWatcher(done)
	require.NotNil(t, wchan)
	// immediate update to trigger load
	select {
	case err, ok := <-wchan:
		assert.True(t, ok)
		assert.Nil(t, err)
	case <-time.After(defaultTimeout):
		assert.Fail(t, "watch didn't return")
	}
	// then block until update
	select {
	case err, ok := <-wchan:
		assert.False(t, ok)
		assert.Fail(t, "unexpected update", err)
	case <-time.After(5 * defaultTimeout):
	}
	// but trigger on update
	tf.Write([]byte("test pattern"))
	select {
	case err, ok := <-wchan:
		assert.True(t, ok)
		assert.Nil(t, err)
	case <-time.After(time.Second):
		assert.Fail(t, "watch didn't return")
	}

	// Close
	close(done)
	select {
	case err, ok := <-wchan:
		assert.False(t, ok)
		assert.Nil(t, err)
	case <-time.After(time.Second):
		assert.Fail(t, "watch didn't exit")
	}

	// Remove
	wf, err = file.New(fname, file.WithWatcher())
	assert.Nil(t, err)
	require.NotNil(t, wf)
	done = make(chan struct{})
	defer close(done)
	wchan = wf.NewWatcher(done)
	require.NotNil(t, wchan)
	// immediate update to trigger load
	select {
	case err, ok := <-wchan:
		assert.True(t, ok)
		assert.Nil(t, err)
	case <-time.After(defaultTimeout):
		assert.Fail(t, "watch didn't return")
	}
	tf.Close()
	os.Remove(fname)
	select {
	case err, ok := <-wchan:
		assert.True(t, ok)
		assert.Nil(t, err)
	case <-time.After(time.Second):
		assert.Fail(t, "watch didn't return")
	}

	// Rename
	tf, err = ioutil.TempFile(".", "file_test_")
	assert.Nil(t, err)
	require.NotNil(t, tf)
	fname = tf.Name()
	wf, err = file.New(fname, file.WithWatcher())
	assert.Nil(t, err)
	require.NotNil(t, wf)
	done = make(chan struct{})
	wchan = wf.NewWatcher(done)
	require.NotNil(t, wchan)
	// immediate update to trigger load
	select {
	case err, ok := <-wchan:
		assert.True(t, ok)
		assert.Nil(t, err)
	case <-time.After(defaultTimeout):
		assert.Fail(t, "watch didn't return")
	}
	os.Rename(fname, fname+"r")
	select {
	case err, ok := <-wchan:
		assert.True(t, ok)
		assert.Nil(t, err)
	case <-time.After(time.Second):
		assert.Fail(t, "watch didn't return")
	}
	os.Remove(fname + "r")
}
