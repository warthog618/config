// Copyright © 2018 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package blob_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/warthog618/config/blob"
)

func TestNewWithSeparator(t *testing.T) {
	l := mockLoader{}
	d := mockDecoder{M: map[string]interface{}{
		"a": map[string]interface{}{"b.c_d": true}}}
	s := blob.New(&l, &d, blob.WithSeparator("-"))
	require.NotNil(t, s)
	v, ok := s.Get("a.b.c_d")
	assert.False(t, ok)
	assert.Nil(t, v)
	v, ok = s.Get("a-b.c_d")
	assert.True(t, ok)
	assert.Equal(t, true, v)
}
