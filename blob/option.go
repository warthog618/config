// Copyright © 2018 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package blob

// Option is a construction option for a Blob.
type Option interface {
	applyOption(s *Getter)
}

// SeparatorOption defines the string that separates tiers in keys.
type SeparatorOption struct {
	s string
}

func (s SeparatorOption) applyOption(x *Getter) {
	x.pathSep = s.s
}

// WithSeparator is an Option that sets the config namespace separator.
// This is an option to ensure it can only set at construction time,
// as changing it at runtime makes no sense.
func WithSeparator(s string) SeparatorOption {
	return SeparatorOption{s}
}
