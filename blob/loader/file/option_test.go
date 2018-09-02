// Copyright © 2018 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package file_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/warthog618/config/blob/loader/file"
)

func TestNewWithWatcher(t *testing.T) {
	// Non-existent
	f, err := file.New("nosuch.file", file.WithWatcher())
	assert.NotNil(t, err)
	assert.Nil(t, f)

	// Existent
	f, err = file.New("file_test.go", file.WithWatcher())
	assert.Nil(t, err)
	require.NotNil(t, f)
	w, ok := f.Watcher()
	assert.True(t, ok)
	assert.NotNil(t, w)
}
