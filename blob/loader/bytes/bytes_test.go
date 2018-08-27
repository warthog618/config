// Copyright © 2018 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package bytes_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/warthog618/config/blob/loader/bytes"
)

func TestNew(t *testing.T) {
	b := bytes.New([]byte{1, 2, 3, 4})
	assert.NotNil(t, b)
}

func TestFileLoad(t *testing.T) {
	b := bytes.New([]byte{1, 2, 3, 4})
	require.NotNil(t, b)
	l, err := b.Load()
	assert.Nil(t, err)
	assert.Equal(t, []byte{1, 2, 3, 4}, l)
}
