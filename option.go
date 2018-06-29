// Copyright © 2018 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package config

// ConfigOption is a construction option for a Config.
type ConfigOption interface {
	applyConfigOption(c *Config)
}

// MustOption is a construction option for a Must.
type MustOption interface {
	applyMustOption(m *Must)
}

// SeparatorOption defines the string that separates tiers in keys.
type SeparatorOption struct {
	s string
}

func (s SeparatorOption) applyAliasOption(a *Alias) {
	a.pathSep = s.s
}

func (s SeparatorOption) applyConfigOption(c *Config) {
	c.pathSep = s.s
}

func (s SeparatorOption) applyMustOption(m *Must) {
	m.c.pathSep = s.s
}

// WithSeparator is an Option that sets the config namespace separator.
// This is an option to ensure it can only set at construction time,
// as changing it at runtime makes no sense.
func WithSeparator(s string) SeparatorOption {
	return SeparatorOption{s}
}

// TagOption defines the string that identies field names during unmarshaling.
type TagOption struct {
	t string
}

func (t TagOption) applyConfigOption(c *Config) {
	c.tag = t.t
}

func (t TagOption) applyMustOption(m *Must) {
	m.c.tag = t.t
}

// WithTag is an Option that sets the config unmarshalling tag.
// The default tag is "config".
func WithTag(t string) TagOption {
	return TagOption{t}
}

// ErrorHandlerOption defines the handler for Must errors.
type ErrorHandlerOption struct {
	e ErrorHandler
}

func (e ErrorHandlerOption) applyMustOption(m *Must) {
	m.e = e.e
}

// WithErrorHandler is an Option that sets the error handling for Must.
// The default is to ignore errors.
func WithErrorHandler(e ErrorHandler) ErrorHandlerOption {
	return ErrorHandlerOption{e}
}

// WithPanic makes a Must panic on error.
func WithPanic() ErrorHandlerOption {
	return ErrorHandlerOption{func(err error) {
		panic(err)
	}}
}
