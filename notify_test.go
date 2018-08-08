// Copyright © 2018 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/warthog618/config"
)

func TestNotify(t *testing.T) {
	s := config.NewNotifier()
	d := s.Notified()
	select {
	case <-d:
		assert.Fail(t, "already notified")
	default:
	}
	s.Notify()
	select {
	case <-d:
	default:
		assert.Fail(t, "not notified")
	}
	d = s.Notified()
	select {
	case <-d:
		assert.Fail(t, "already notified")
	default:
	}
}
