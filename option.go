// Copyright © 2018 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package config

// Option is a construction option for a Config.
type Option interface {
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

// WithTag is an Option that sets the config unmarshalling tag.
// The default tag is "config".
func WithTag(t string) TagOption {
	return TagOption{t}
}

// TagOption defines the string that identifies field names during unmarshaling.
type TagOption struct {
	t string
}

func (t TagOption) applyConfigOption(c *Config) {
	c.tag = t.t
}

// DefaultOption defines default configuration uses as a fall back if a field is not returned by the main getter.
type DefaultOption struct {
	d Getter
}

func (o DefaultOption) applyConfigOption(c *Config) {
	c.defg = o.d
}

// WithDefault is an Option that sets the default configuration.
// If applied multiple times, the earlier defaults are ignored.
func WithDefault(d Getter) DefaultOption {
	return DefaultOption{d}
}

// ErrorHandlerOption defines the handler for errors.
type ErrorHandlerOption struct {
	e ErrorHandler
}

func (o ErrorHandlerOption) applyConfigOption(c *Config) {
	c.geh = o.e
	c.veh = o.e
}

func (o ErrorHandlerOption) applyValueOption(v *Value) {
	v.eh = o.e
}

// WithErrorHandler is an Option that sets the error handling for a Config or Value.
// For Config this applies to Get and is propagated to returned Values.
// For Value this applies to all type conversions.
func WithErrorHandler(e ErrorHandler) ErrorHandlerOption {
	return ErrorHandlerOption{e}
}

// WithGetErrorHandler is an Option that sets the error handling for a Config Gets.
func WithGetErrorHandler(e ErrorHandler) GetErrorHandlerOption {
	return GetErrorHandlerOption{e}
}

// GetErrorHandlerOption defines the handler for errors returned by Gets.
type GetErrorHandlerOption struct {
	e ErrorHandler
}

func (o GetErrorHandlerOption) applyConfigOption(c *Config) {
	c.geh = o.e
}

// WithValueErrorHandler is an Option that sets the error handling for a Config Gets.
func WithValueErrorHandler(e ErrorHandler) ValueErrorHandlerOption {
	return ValueErrorHandlerOption{e}
}

// ValueErrorHandlerOption defines the error handler added to Values returned by
// Config Gets.
// These may be overridden by ValueOptions.
type ValueErrorHandlerOption struct {
	e ErrorHandler
}

func (o ValueErrorHandlerOption) applyConfigOption(c *Config) {
	c.veh = o.e
}

// WithMust makes an object panic on error.
// For Config this applies to Get and is propagated to returned Values.
// For Value this applies to all type conversions.
func WithMust() ErrorHandlerOption {
	return ErrorHandlerOption{func(err error) error {
		panic(err)
	}}
}

// WithZeroDefaults makes an object ignore errors and instead return zeroed
// default values.
// For Config this applies to Get and is propagated to returned Values.
// For Value this applies to all type conversions.
func WithZeroDefaults() ErrorHandlerOption {
	return ErrorHandlerOption{func(err error) error {
		return nil
	}}
}
