// Copyright © 2018 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package blob_test

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/warthog618/config"
	"github.com/warthog618/config/blob"
)

func TestWithTemporary(t *testing.T) {
	te := blob.WithTemporary(nil)
	assert.Nil(t, te)

	e := errors.New("base error")
	te = blob.WithTemporary(e)
	assert.Equal(t, e, errors.Cause(te))
	assert.Equal(t, e.Error(), te.Error())
	assert.False(t, config.IsTemporary(e))
	assert.True(t, config.IsTemporary(te))
}
