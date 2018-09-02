// Copyright © 2018 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package config

import (
	"context"
	"sync"
)

// NewStack creates a Stack.
func NewStack(gg ...Getter) *Stack {
	s := Stack{}
	for _, g := range gg {
		if g != nil {
			s.gg = append(s.gg, g)
		}
	}
	return &s
}

// Stack attempts a get using a list of Getters, in the order provided,
// returning the first result found.
// Additional layers may be added to the stack at runtime.
type Stack struct {
	// RWLock covering gg.
	// It does not prevent concurrent access to the getters themselves,
	// only to the gg array itself.
	mu sync.RWMutex
	// A list of Getters providing config key/value pairs.
	gg []Getter
	// watcher of the getters
	w *stackWatcher
}

// Append appends a getter to the set of getters for the Stack.
// This means this getter is only used as a last resort, relative to
// the existing getters.
func (s *Stack) Append(g Getter) {
	if g == nil {
		return
	}
	s.mu.Lock()
	s.gg = append(s.gg, g)
	if s.w != nil {
		s.w.append(g)
	}
	s.mu.Unlock()
}

// Get gets the raw value corresponding to the key.
// It iterates through the list of getters, searching for a matching key.
// Returns the first match found, or an error if none is found.
func (s *Stack) Get(key string) (interface{}, bool) {
	s.mu.RLock()
	for _, g := range s.gg {
		if v, ok := g.Get(key); ok {
			s.mu.RUnlock()
			return v, ok
		}
	}
	s.mu.RUnlock()
	return nil, false
}

// Insert inserts a getter to the set of getters for the Stack.
// This means this getter is used before the existing getters.
func (s *Stack) Insert(g Getter) {
	if g == nil {
		return
	}
	s.mu.Lock()
	s.gg = append([]Getter{g}, s.gg...)
	if s.w != nil {
		s.w.append(g)
	}
	s.mu.Unlock()
}

// Watcher implements the WatchableGetter interface.
func (s *Stack) Watcher() (GetterWatcher, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.w != nil {
		return s.w, true
	}
	ww := []GetterWatcher{}
	for _, g := range s.gg {
		if wg, ok := g.(WatchableGetter); ok {
			if w, ok := wg.Watcher(); ok {
				ww = append(ww, w)
			}
		}
	}
	s.w = &stackWatcher{
		mu:    &sync.Mutex{},
		gg:    ww,
		uchan: make(chan GetterWatcher),
		cchan: make(chan GetterWatcher, 1)}
	return s.w, true
}

type stackWatcher struct {
	mu      sync.Locker
	gg      []GetterWatcher
	wctx    context.Context
	wcancel func()
	uchan   chan GetterWatcher
	cchan   chan GetterWatcher
}

func (s *stackWatcher) Close() (rerr error) {
	s.mu.Lock()
	if s.wcancel != nil {
		s.wcancel()
	}
	gg := s.gg
	s.mu.Unlock()
	for _, g := range gg {
		err := g.Close()
		if rerr == nil {
			rerr = err
		}
	}
	return
}

func (s *stackWatcher) CommitUpdate() {
	for {
		select {
		case g := <-s.cchan:
			g.CommitUpdate()
			go s.watchGetter(s.wctx, g)
		default:
			return
		}
	}
}

func (s *stackWatcher) Watch(ctx context.Context) error {
	s.mu.Lock()
	if s.wctx == nil {
		ctx, cancel := context.WithCancel(context.Background())
		for _, g := range s.gg {
			go s.watchGetter(ctx, g)
		}
		s.wctx = ctx
		s.wcancel = cancel
	}
	s.mu.Unlock()
	select {
	case g := <-s.uchan:
		s.cchan <- g
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *stackWatcher) watchGetter(ctx context.Context, gw GetterWatcher) {
	for {
		if err := gw.Watch(ctx); err != nil {
			if IsTemporary(err) {
				continue
			}
			return
		}
		break
	}
	select {
	case <-ctx.Done():
		return
	case s.uchan <- gw:
	}
}

func (s *stackWatcher) append(g Getter) {
	wg, ok := g.(WatchableGetter)
	if !ok {
		return
	}
	w, ok := wg.Watcher()
	if !ok {
		return
	}
	s.mu.Lock()
	s.gg = append(s.gg, w)
	if s.wctx != nil {
		go s.watchGetter(s.wctx, w)
	}
	s.mu.Unlock()
}
