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
	f := file.New("file_test.go")
	require.NotNil(t, f)

	// Non-existent
	f = file.New("nosuch.file")
	require.NotNil(t, f)
	assert.Implements(t, (*blob.WatchableLoader)(nil), f)
}

func TestLoad(t *testing.T) {
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

var defaultTimeout = time.Millisecond

func TestWatcherClose(t *testing.T) {
	f, err := ioutil.TempFile(".", "file_test_")
	defer f.Close()
	assert.Nil(t, err)
	require.NotNil(t, f)
	fname := f.Name()
	defer os.Remove(fname)
	wf := file.New(fname, file.WithWatcher())
	require.NotNil(t, wf)
	done := make(chan struct{})
	wchan := wf.NewWatcher(done)
	require.NotNil(t, wchan)
	// immediate update to trigger load
	select {
	case err, ok := <-wchan:
		assert.True(t, ok)
		assert.Nil(t, err)
	case <-time.After(time.Second):
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
	wf := file.New(fname)
	require.NotNil(t, wf)
	done := make(chan struct{})
	defer close(done)
	wchan := wf.NewWatcher(done)
	require.Nil(t, wchan)

	// Watched
	wf = file.New(fname, file.WithWatcher())
	require.NotNil(t, wf)
	wchan = wf.NewWatcher(done)
	require.NotNil(t, wchan)
}

func TestWatcherWatch(t *testing.T) {
	// Write
	tf, err := ioutil.TempFile("", "file_test_")
	assert.Nil(t, err)
	require.NotNil(t, tf)
	fname := tf.Name()
	wf := file.New(fname, file.WithWatcher())
	require.NotNil(t, wf)
	done := make(chan struct{})
	wchan := wf.NewWatcher(done)
	require.NotNil(t, wchan)
	// immediate update to trigger load
	testUpdated(t, wchan)
	// then block until update
	testNotUpdated(t, wchan)
	// but trigger on update
	tf.Write([]byte("test pattern"))
	testUpdated(t, wchan)

	// Close
	close(done)
	testCanceled(t, wchan)

	// Remove
	wf = file.New(fname, file.WithWatcher())
	require.NotNil(t, wf)
	done = make(chan struct{})
	wchan = wf.NewWatcher(done)
	require.NotNil(t, wchan)
	testUpdated(t, wchan)
	tf.Close()
	os.Remove(fname)
	testUpdated(t, wchan)
	testErrored(t, wchan)

	close(done)
	testCanceled(t, wchan)

	// Overwrite
	tf, err = ioutil.TempFile("", "file_test_")
	assert.Nil(t, err)
	require.NotNil(t, tf)
	defer tf.Close()
	fname = tf.Name()
	wf = file.New(fname, file.WithWatcher())
	require.NotNil(t, wf)
	done = make(chan struct{})
	defer close(done)
	wchan = wf.NewWatcher(done)
	require.NotNil(t, wchan)
	// immediate update to trigger load
	testUpdated(t, wchan)
	tf, err = ioutil.TempFile("", "file_test_")
	assert.Nil(t, err)
	require.NotNil(t, tf)
	defer tf.Close()
	fname2 := tf.Name()
	os.Rename(fname2, fname)
	testUpdated(t, wchan)
	defer os.Remove(fname)

	// Rename
	tf, err = ioutil.TempFile("", "file_test_")
	assert.Nil(t, err)
	require.NotNil(t, tf)
	defer tf.Close()
	fname = tf.Name()
	wf = file.New(fname, file.WithWatcher())
	require.NotNil(t, wf)
	done = make(chan struct{})
	defer close(done)
	wchan = wf.NewWatcher(done)
	require.NotNil(t, wchan)
	// immediate update to trigger load
	testUpdated(t, wchan)
	os.Rename(fname, fname+"r")
	testUpdated(t, wchan)
	defer os.Remove(fname + "r")
}

func testErrored(t *testing.T, wchan <-chan error) {
	t.Helper()
	select {
	case err, ok := <-wchan:
		assert.True(t, ok)
		assert.NotNil(t, err)
	case <-time.After(time.Second):
		assert.Fail(t, "watch didn't return")
	}
}

func testUpdated(t *testing.T, wchan <-chan error) {
	t.Helper()
	select {
	case err, ok := <-wchan:
		assert.True(t, ok)
		assert.Nil(t, err)
	case <-time.After(time.Second):
		assert.Fail(t, "watch didn't return")
	}
}

func testCanceled(t *testing.T, wchan <-chan error) {
	t.Helper()
	select {
	case err, ok := <-wchan:
		assert.False(t, ok)
		assert.Nil(t, err)
	case <-time.After(time.Second):
		assert.Fail(t, "watch didn't return")
	}
}
func testNotUpdated(t *testing.T, wchan <-chan error) {
	t.Helper()
	select {
	case err, ok := <-wchan:
		assert.False(t, ok)
		assert.Fail(t, "unexpected update", err)
	case <-time.After(5 * defaultTimeout):
	}

}
