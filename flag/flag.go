// Copyright © 2017 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package flag provides a POSIX/GNU style style command line parser/config getter.
//
// Parses command line flags and makes them available
// through a config.Getter interface.
//
// Handles:
//
//   - short form (-h)
//   - long form (--help)
//   - bool flags accessible as bool or int (the latter giving a count of the times the flag occurs)
//   - -- terminator to forcilby terminate flag parsing before remaining args.
//   - returning remaining args after parsing.
//
// Flags are assumed bool unless followed by a non-flag.
// Values following a flag are associated with the previous flag.
// In both short and long form, an '=' may be placed between the flag and value.
// e.g.
//
//   -c config and -c=config are equivalent
//
// The long form equivalents are:
//
//   --config-file config and --config-file=config
//
// Short flags may be grouped, e.g. "-abc" is equivalent to "-a -b -c".
// Grouped short flags may not have values following them.
// So "-abc foo" would stop parsing after the "-abc" and assume "foo" was
// the first non-flag an arg.
package flag

import (
	"os"
	"strings"

	"github.com/warthog618/config/keys"
)

// New creates a new Getter.
//
// By default the Getter will:
// - parse the command line os.Args[1:]
// - replace '-' in the flag space with '.' in the config space.
// - split list values with the ',' separator.
func New(options ...Option) (*Getter, error) {
	r := Getter{listSeparator: ","}
	for _, option := range options {
		option(&r)
	}
	if r.keyMapper == nil {
		r.keyMapper = keys.ReplaceMapper{From: "-", To: "."}
	}
	if r.cmdArgs == nil {
		r.cmdArgs = os.Args[1:]
	}
	r.parse()
	return &r, nil
}

// Getter provides the mapping from command line arguments to a config.Getter.
// The Getter scans the command line only at construction time, so its config state
// is effectively immutable.
type Getter struct {
	// The args to parse into config values.
	cmdArgs []string
	// residual args after flag parsing.
	args []string
	// config key=value
	config map[string]interface{}
	// map of short flag characters to long form flag name
	shortFlags map[byte]string
	// A replacer that maps from flag space to config space.
	keyMapper Mapper
	// The separator for slices stored in string values.
	listSeparator string
}

// Mapper maps a key from one space to another.
type Mapper interface {
	Map(string) string
}

// Option is a function which modifies a Getter at construction time.
type Option func(*Getter)

// WithCommandLine uses the provided command line as the source of config
// instead of os.Args[1:].
// The provided command line should NOT include the name of the executable
// (os.Args[0]).
func WithCommandLine(cmdArgs []string) Option {
	return func(r *Getter) {
		r.cmdArgs = cmdArgs
	}
}

// WithKeyMapper sets the mapper used to map from flag space to config space.
// The default replaces '-' in the flag space with '.' in the config space.
func WithKeyMapper(keyMapper Mapper) Option {
	return func(r *Getter) {
		r.keyMapper = keyMapper
	}
}

// WithListSeparator sets the separator between slice fields in the flag space.
// The default separator is ","
func WithListSeparator(separator string) Option {
	return func(r *Getter) {
		r.listSeparator = separator
	}
}

// WithShortFlags sets the set of short flags to be parsed from the command line.
// The shortFlags defines the mapping from single character short flags to
// long flag names.  Long names are within the flag space and so should
// use the appropriate tier separator. e.g. {'c':"config-file"}
func WithShortFlags(shortFlags map[byte]string) Option {
	return func(r *Getter) {
		r.shortFlags = shortFlags
	}
}

// Args returns the trailing arguments from the command line that are not flags,
// or flag values.
func (r *Getter) Args() []string {
	return r.args
}

// NArg returns the number of trailing args in the command line.
func (r *Getter) NArg() int {
	return len(r.args)
}

// NFlag returns the number of flags detected in the command line.
// Multiple instances of the same flag, in either short or long form, count
// as a single flag.
func (r *Getter) NFlag() int {
	return len(r.config)
}

// Get returns the value for a given key and true if found, or
// nil and false if not.
func (r *Getter) Get(key string) (interface{}, bool) {
	v, ok := r.config[key]
	if ok && len(r.listSeparator) > 0 {
		if vstr, sok := v.(string); sok {
			if strings.Contains(vstr, r.listSeparator) {
				return strings.Split(vstr, r.listSeparator), ok
			}
		}
	}
	return v, ok
}

func incrementFlag(config map[string]interface{}, key string) {
	if v, ok := config[key]; ok {
		if vint, ok := v.(int); ok {
			config[key] = vint + 1
			return
		}
	}
	config[key] = 1
}

func (r *Getter) parse() {
	config := map[string]interface{}{}
	for idx := 0; idx < len(r.cmdArgs); idx++ {
		arg := r.cmdArgs[idx]
		if strings.HasPrefix(arg, "--") {
			if len(arg) == 2 {
				// -- terminator
				r.args = r.cmdArgs[idx+1:]
				break
			}
			// long form
			arg = arg[2:]
			if strings.Contains(arg, "=") {
				// split on = and process complete in place
				s := strings.SplitN(arg, "=", 2)
				config[r.keyMapper.Map(s[0])] = s[1]
			} else {
				key := r.keyMapper.Map(arg)
				if idx < len(r.cmdArgs)-1 {
					val := r.cmdArgs[idx+1]
					if strings.HasPrefix(val, "-") {
						incrementFlag(config, key)
					} else {
						config[key] = val
						idx++
					}
				} else {
					incrementFlag(config, key)
				}
			}
		} else if strings.HasPrefix(arg, "-") {
			// short form
			arg = arg[1:]
			if len(arg) > 1 && !strings.Contains(arg, "=") {
				// grouped short flags
				for sidx := 0; sidx < len(arg); sidx++ {
					if flag, ok := r.shortFlags[arg[sidx]]; ok {
						incrementFlag(config, r.keyMapper.Map(flag))
					}
				}
				continue
			}
			val := ""
			if strings.Index(arg, "=") == 1 {
				val = arg[2:]
			} else if len(arg) != 1 {
				// ignore malformed flag
				continue
			} else {
				if idx < len(r.cmdArgs)-1 {
					v := r.cmdArgs[idx+1]
					if v[0] != '-' {
						val = v
						idx++
					}
				}
			}
			if flag, ok := r.shortFlags[arg[0]]; ok {
				key := r.keyMapper.Map(flag)
				if val == "" {
					incrementFlag(config, key)
				} else {
					config[key] = val
				}
			}
		} else {
			// non-flag terminator
			r.args = r.cmdArgs[idx:]
			break
		}
	}
	r.config = config
}
