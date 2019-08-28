// Copyright © 2018 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package blob_test

import (
	"sync"
	"testing"
	"time"

	"github.com/pkg/errors"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/warthog618/config"
	"github.com/warthog618/config/blob"
	"github.com/warthog618/config/blob/decoder/json"
)

var defaultTimeout = 10 * time.Millisecond

func TestNew(t *testing.T) {
	l := newMockLoader(nil)
	d := mockDecoder{}

	// all good
	s := blob.New(l, &d)
	require.NotNil(t, s)
	assert.Implements(t, (*config.Getter)(nil), s)
	assert.Implements(t, (*config.WatchableGetter)(nil), s)

	dcerr := errors.New("decode error")
	lderr := errors.New("load error")

	// load error
	l = newMockLoader(nil)
	l.LoadError = lderr
	tl := func() {
		blob.New(l, &d, blob.MustLoad())
	}
	assert.PanicsWithValue(t, lderr, tl, "MustLoad - load error didn't panic")

	// decode error
	l = newMockLoader(nil)
	d = mockDecoder{DecodeError: dcerr}
	tl = func() {
		blob.New(l, &d, blob.MustLoad())
	}
	assert.PanicsWithValue(t, dcerr, tl, "MustLoad - decode error didn't panic")
}

func TestNewWatcher(t *testing.T) {
	l := newMockLoader(nil)
	d := mockDecoder{}

	// unwatchable
	s := blob.New(&bareLoader{}, &d)
	require.NotNil(t, s)
	assert.Implements(t, (*config.WatchableGetter)(nil), s)
	done := make(chan struct{})
	defer close(done)
	w := s.NewWatcher(done)
	assert.Nil(t, w)

	// watchable
	s = blob.New(l, &d)
	require.NotNil(t, s)
	w = s.NewWatcher(done)
	assert.NotNil(t, w)

	// watchable, but disabled
	s = blob.New(&mockLoader{}, &d)
	require.NotNil(t, s)
	w = s.NewWatcher(done)
	assert.Nil(t, w)
}

func TestGetterAsOption(t *testing.T) {
	l := newMockLoader(nil)
	d := mockDecoder{}
	c := config.New(blob.New(l, &d), blob.New(l, &d))
	c.Close()
}

func TestGet(t *testing.T) {
	l := newMockLoader(nil)
	d := mockDecoder{M: map[string]interface{}{
		"a": map[string]interface{}{"b.c_d": true}}}
	s := blob.New(l, &d)
	require.NotNil(t, s)
	v, ok := s.Get("")
	assert.False(t, ok)
	assert.Nil(t, v)
	v, ok = s.Get("a.b.c_d")
	assert.True(t, ok)
	assert.Equal(t, true, v)

	// bad load
	lderr := errors.New("load error")
	l.LoadError = lderr
	s = blob.New(l, &d)
	require.NotNil(t, s)
	v, ok = s.Get("")
	assert.False(t, ok)
	assert.Nil(t, v)
	v, ok = s.Get("a.b.c_d")
	assert.False(t, ok)
	assert.Nil(t, v)
}

func TestWatch(t *testing.T) {
	l := newMockLoader(nil)
	d := mockDecoder{M: map[string]interface{}{"a.b.c_d": "baseline"}}
	s := blob.New(l, &d)
	require.NotNil(t, s)
	done := make(chan struct{})
	defer close(done)
	w := s.NewWatcher(done)
	require.NotNil(t, w)

	// baseline
	testNotUpdated(t, w)

	// update
	d.SetM(map[string]interface{}{"a.b.c_d": "updated"})
	go l.Modify(nil)
	testUpdated(t, w, nil)

	// no content change => no return
	d.SetM(map[string]interface{}{"a.b.c_d": "baseline"})
	go l.Modify(nil)
	testNotUpdated(t, w)

	// load watch error
	loadError := errors.New("Load error")
	go l.Modify(loadError)
	update := testUpdated(t, w, loadError)
	require.NotNil(t, update)
	assert.NotPanics(t, update.Commit)

	// bad decode
	d.SetM(map[string]interface{}{"a.b.c_d": "final"})
	d.DecodeError = errors.New("Decode error")
	go l.Modify(nil)
	update = testUpdated(t, w, d.DecodeError)
	require.NotNil(t, update)
	assert.NotPanics(t, update.Commit)

	// pathological decoder
	d.SetM(nil)
	d.DecodeError = nil
	go l.Modify(nil)
	testNotUpdated(t, w)

	// loader closed
	l.Close()
	d.SetM(map[string]interface{}{"a.b.c_d": "closed"})
	testClosed(t, w.Update())
}

func TestUpdate(t *testing.T) {
	l := newMockLoader(nil)
	d := mockDecoder{M: map[string]interface{}{"a.b.c_d": "baseline"}}
	s := blob.New(l, &d)
	require.NotNil(t, s)
	done := make(chan struct{})
	w := s.NewWatcher(done)
	require.NotNil(t, w)

	// baseline
	testNotUpdated(t, w)

	// update
	d.SetM(map[string]interface{}{"a.b.c_d": "updated"})
	go l.Modify(nil)
	update := testUpdated(t, w, nil)
	require.NotNil(t, update)
	v, ok := s.Get("a.b.c_d")
	assert.True(t, ok)
	assert.Equal(t, "baseline", v)
	assert.NotPanics(t, update.Commit)
	v, ok = s.Get("a.b.c_d")
	assert.True(t, ok)
	assert.Equal(t, "updated", v)
	ge, ok := update.(Getter)
	assert.True(t, ok)
	assert.Equal(t, s, ge.Getter())
	te, ok := update.(TemporaryError)
	assert.True(t, ok)
	assert.False(t, te.TemporaryError())

	// closed while updating
	d.SetM(map[string]interface{}{"a.b.c_d": "final"})
	go l.Modify(nil)
	time.Sleep(defaultTimeout)
	close(done)
	time.Sleep(defaultTimeout)
	testClosed(t, w.Update())
}

func TestNewConfigFile(t *testing.T) {
	// specified
	l := newMockLoader(nil)
	d := mockDecoder{M: map[string]interface{}{
		"cfg": "blob_test.json",
		"go":  "blob_test.go"}}
	b := blob.New(l, &d)
	c := config.New(b)
	defer c.Close()
	jsondec := json.NewDecoder()
	f := blob.NewConfigFile(c, "cfg", "no_such_file.json", jsondec)
	assert.NotNil(t, f)
	a, ok := f.Get("a")
	assert.True(t, ok)
	assert.Equal(t, a, "from blob_test.json")

	// specified wont load
	p := func() { f = blob.NewConfigFile(c, "go", "no_such_file.json", jsondec) }
	assert.Panics(t, p)

	// default
	f = blob.NewConfigFile(c, "cfg2", "blob_test.json", jsondec)
	assert.NotNil(t, f)
	a, ok = f.Get("a")
	assert.True(t, ok)
	assert.Equal(t, a, "from blob_test.json")

	// default wont load
	p = func() { f = blob.NewConfigFile(c, "cfg2", "blob_test.go", jsondec) }
	assert.Panics(t, p)

	// neither
	f = blob.NewConfigFile(c, "cfg2", "no_such_file.json", jsondec)
	assert.NotNil(t, f)
	a, ok = f.Get("a")
	assert.False(t, ok)
	assert.Nil(t, a)
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
type bareLoader struct{}

func (l *bareLoader) Load() ([]byte, error) {
	return nil, nil
}

type mockLoader struct {
	mu        sync.Mutex
	B         []byte
	LoadError error
	update    chan error
}

func newMockLoader(b []byte) *mockLoader {
	return &mockLoader{B: b, update: make(chan error)}
}
func (l *mockLoader) Load() ([]byte, error) {
	return l.B, l.LoadError
}

func (l *mockLoader) NewWatcher(done <-chan struct{}) <-chan error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.update
}

func (l *mockLoader) Modify(err error) {
	l.mu.Lock()
	l.update <- err
	l.mu.Unlock()
}

func (l *mockLoader) Close() {
	l.mu.Lock()
	close(l.update)
	l.mu.Unlock()
}

type mockDecoder struct {
	mu          sync.Mutex
	M           map[string]interface{}
	DecodeError error
}

func (d *mockDecoder) Decode(b []byte, v interface{}) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	m := v.(*map[string]interface{})
	*m = d.M
	return d.DecodeError
}

func (d *mockDecoder) SetM(m map[string]interface{}) {
	d.mu.Lock()
	d.M = m
	d.mu.Unlock()
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
