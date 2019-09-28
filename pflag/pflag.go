// Copyright © 2017 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package pflag provides a POSIX/GNU style style command line parser/config Getter.
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
// the first non-flag argument.
package pflag

import (
	"os"
	"strings"

	"github.com/warthog618/config"
	"github.com/warthog618/config/keys"
	"github.com/warthog618/config/list"
	"github.com/warthog618/config/tree"
)

// New creates a new Getter.
//
// By default the Getter will:
// - parse the command line os.Args[1:]
// - replace '-' in the flag space with '.' in the config space.
// - split list values with the ',' separator.
func New(options ...Option) *Getter {
	g := Getter{}
	for _, option := range options {
		option(&g)
	}
	if g.keyReplacer == nil {
		g.keyReplacer = keys.StringReplacer("-", ".")
	}
	if g.listSplitter == nil {
		g.listSplitter = list.NewSplitter(",")
	}
	if g.cmdArgs == nil {
		g.cmdArgs = os.Args[1:]
	}
	if len(g.flags) != 0 {
		g.shortFlags = make(map[rune]string)
		g.boolFlags = make(map[string]bool)
		for _, f := range g.flags {
			if len(f.Name) == 0 {
				// ignore unnamed flags
				continue
			}
			if f.Short != 0 {
				g.shortFlags[f.Short] = f.Name
			}
			key := g.keyReplacer.Replace(f.Name)
			if f.Options&IsBool != 0 {
				g.boolFlags[key] = true
			}
		}
	}
	g.parse()
	return &g
}

// Flag represents one of the command line flags.
type Flag struct {
	Name    string
	Short   rune
	Options FlagOptions
}

// FlagOptions is a set of boolean options for a flag.
type FlagOptions int

const (
	// IsBool indicates the flag is only allowed to be a bool
	IsBool FlagOptions = 1 << iota
)

// Getter provides the mapping from command line arguments to a config.Getter.
//
// The Getter scans the command line only at construction time, so its config state
// is effectively immutable.
type Getter struct {
	config.GetterAsOption

	// The args to parse into config values.
	cmdArgs []string

	// residual args after flag parsing.
	args []string

	// config key=value
	config map[string]interface{}

	// set of flags that get special treatment
	flags []Flag

	// map of short flag characters to long form flag name
	shortFlags map[rune]string

	// set of flag names that are boolean
	boolFlags map[string]bool

	// A replacer that maps from flag space to config space.
	keyReplacer keys.Replacer

	// The splitter for slices stored in string values.
	listSplitter list.Splitter
}

// Option is a function which modifies a Getter at construction time.
type Option func(*Getter)

// WithCommandLine uses the provided command line as the source of config
// instead of os.Args[1:].
//
// The provided command line should NOT include the name of the executable
// (os.Args[0]).
func WithCommandLine(cmdArgs []string) Option {
	return func(g *Getter) {
		g.cmdArgs = cmdArgs
	}
}

// WithKeyReplacer sets the replacer used to map from flag space to config space.
// The default replaces '-' in the flag space with '.' in the config space.
func WithKeyReplacer(keyReplacer keys.Replacer) Option {
	return func(g *Getter) {
		g.keyReplacer = keyReplacer
	}
}

// WithListSplitter splits slice fields stored as strings in the pflag space.
//
// The default splitter separates on ",".
func WithListSplitter(splitter list.Splitter) Option {
	return func(g *Getter) {
		g.listSplitter = splitter
	}
}

// WithFlags places constraints on some flags as they are parsed.
//
// This option is not necessary to specify flags - only those that require
// special treatment.
func WithFlags(flags []Flag) Option {
	return func(r *Getter) {
		r.flags = flags
	}
}

// Args returns the trailing arguments from the command line that are not flags,
// or flag values.
func (g *Getter) Args() []string {
	return g.args
}

// NArg returns the number of trailing args in the command line.
func (g *Getter) NArg() int {
	return len(g.args)
}

// NFlag returns the number of flags detected in the command line.
//
// Multiple instances of the same flag, in either short or long form, count
// as a single flag.
func (g *Getter) NFlag() int {
	return len(g.config)
}

// Get returns the value for a given key and true if found, or nil and false if
// not.
func (g *Getter) Get(key string) (interface{}, bool) {
	return tree.Get(g.config, key, "")
}

func (g *Getter) parse() {
	config := map[string]interface{}{}
	for idx := 0; idx < len(g.cmdArgs); idx++ {
		arg := g.cmdArgs[idx]
		nxarg := ""
		if idx < len(g.cmdArgs)-1 {
			if !strings.HasPrefix(g.cmdArgs[idx+1], "-") {
				nxarg = g.cmdArgs[idx+1]
			}
		}
		if strings.HasPrefix(arg, "--") {
			if len(arg) == 2 {
				// -- terminator
				g.args = g.cmdArgs[idx+1:]
				break
			}
			// long form
			arg = arg[2:]
			idx += g.parseLongForm(config, arg, nxarg)
		} else if strings.HasPrefix(arg, "-") {
			// short form
			arg = arg[1:]
			idx += g.parseShortForm(config, arg, nxarg)
		} else {
			// non-flag terminator
			g.args = g.cmdArgs[idx:]
			break
		}
	}
	g.config = config
}

// parses the short form flags.
// Returns 1 if it absorbs the nxarg.
func (g *Getter) parseShortForm(config map[string]interface{}, arg, nxarg string) int {
	if len(arg) > 1 && !strings.Contains(arg, "=") {
		// grouped short flags
		for _, ch := range arg {
			if flag, ok := g.shortFlags[ch]; ok {
				incrementFlag(config, g.keyReplacer.Replace(flag))
			}
		}
		return 0
	}
	if flag, ok := g.shortFlags[rune(arg[0])]; ok {
		key := g.keyReplacer.Replace(flag)
		val := ""
		switch {
		case strings.Index(arg, "=") == 1:
			val = arg[2:]
			config[key] = g.listSplitter.Split(val)
		case len(arg) != 1:
			// ignore malformed flag
			return 0
		case g.boolFlags[key] == true:
			// ignore any trailing value
			incrementFlag(config, key)
		case len(nxarg) != 0:
			config[key] = g.listSplitter.Split(nxarg)
			return 1
		default:
			incrementFlag(config, key)
		}
	}
	return 0
}

// parses the long form flags.
// Returns 1 if it absorbs the nxarg.
func (g *Getter) parseLongForm(config map[string]interface{}, arg, nxarg string) int {
	if strings.Contains(arg, "=") {
		// split on = and process complete in place
		s := strings.SplitN(arg, "=", 2)
		key := g.keyReplacer.Replace(s[0])
		config[key] = g.listSplitter.Split(s[1])
	} else {
		key := g.keyReplacer.Replace(arg)
		switch {
		case g.boolFlags[key] == true:
			incrementFlag(config, key)
		case len(nxarg) > 0:
			config[key] = g.listSplitter.Split(nxarg)
			return 1
		default:
			incrementFlag(config, key)
		}
	}
	return 0
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
