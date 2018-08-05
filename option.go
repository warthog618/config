// Copyright © 2018 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package config

// ConfigOption is a construction option for a Config.
type ConfigOption interface {
	applyConfigOption(c *Config)
}

// SeparatorOption defines the string that separates tiers in keys.
type SeparatorOption struct {
	s string
}

// ValueOption is a construction option for a Value.
type ValueOption interface {
	applyValueOption(v *Value)
}

func (s SeparatorOption) applyAliasOption(a *Alias) {
	a.pathSep = s.s
}

func (s SeparatorOption) applyConfigOption(c *Config) {
	c.pathSep = s.s
}

// WithSeparator is an Option that sets the config namespace separator.
// This is an option to ensure it can only set at construction time,
// as changing it at runtime makes no sense.
func WithSeparator(s string) SeparatorOption {
	return SeparatorOption{s}
}

// SignalOption defines the signal the config will use to indicate updates.
type SignalOption struct {
	s *Signal
}

func (s SignalOption) applyConfigOption(c *Config) {
	c.signal = s.s
}

// WithUpdateSignal is an Option that sets the signal the config
// uses to indicate it has been updated.
func WithUpdateSignal(s *Signal) SignalOption {
	return SignalOption{s}
}

// WithTag is an Option that sets the config unmarshalling tag.
// The default tag is "config".
func WithTag(t string) TagOption {
	return TagOption{t}
}

// TagOption defines the string that identies field names during unmarshaling.
type TagOption struct {
	t string
}

func (t TagOption) applyConfigOption(c *Config) {
	c.tag = t.t
}

// ErrorHandlerOption defines the handler for errors.
type ErrorHandlerOption struct {
	e ErrorHandler
}

func (e ErrorHandlerOption) applyConfigOption(c *Config) {
	c.eh = e.e
}

func (e ErrorHandlerOption) applyValueOption(v *Value) {
	v.eh = e.e
}

// WithErrorHandler is an Option that sets the error handling for Must.
// The default is to ignore errors.
func WithErrorHandler(e ErrorHandler) ErrorHandlerOption {
	return ErrorHandlerOption{e}
}

// WithPanic makes an object panic on error.
func WithPanic() ErrorHandlerOption {
	return ErrorHandlerOption{func(err error) {
		panic(err)
	}}
}
