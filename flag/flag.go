// POSIX/GNU style style command line parser/config reader.
//
// Parses command line flags and makes them available
// through a config.Reader interface.
//
// Handles:
// - short form (-h)
// - long form (--help)
// - bool flags accessible as bool or int (the latter giving a count of the times the flag occurs)
// - -- terminator to foriclby terminate flag parsing before remaining args.
// - returning remaining args after parsing.
//
// Flags are assumed bool unless followed by a non-flag.
// Values following a flag are associated with the previous flag.
// In both short and long form, an '=' may be placed between the flag and value.
// e.g.
// -c config and -c=config are equivalent
// The long form equivalents are:
// --config-file config and --config-file=config
//
// Short flags may be grouped, e.g. -abc is equivalent to -a -b -c.
// Grouped short flags may not have values following them.
// So -abc config would stop parsing after the -abc and assume config was
// the first non-flag an arg.
package flag

import (
	"os"
	"strings"
)

func New(cmdArgs []string, shortFlags map[byte]string) (*reader, error) {
	if len(cmdArgs) == 0 {
		cmdArgs = os.Args[1:]
	}
	args := []string{}
	config := map[string]interface{}(nil)
	nodes := map[string]bool(nil)
	r := reader{cmdArgs, args, config, nodes, shortFlags, "-", ".", ","}
	r.parse()
	return &r, nil
}

type reader struct {
	// The args to parse into config values.
	cmdArgs []string
	// residual args after flag parsing.
	args []string
	// config key=value
	config map[string]interface{}
	// set of nodes contained in config
	nodes map[string]bool
	// map of short flag characters to long form flag name
	shortFlags map[byte]string
	// separator between key tiers in flag space.
	flagSeparator string
	// separator between key tiers in config space.
	cfgSeparator string
	// The separator for slices stored in string values.
	listSeparator string
}

func (r *reader) Args() []string {
	return r.args
}

func (r *reader) NArg() int {
	return len(r.args)
}

func (r *reader) NFlag() int {
	return len(r.config)
}

func (r *reader) SetShortFlag(shortFlag byte, longFlag string) {
	r.shortFlags[shortFlag] = longFlag
	r.parse()
}

func (r *reader) Contains(key string) bool {
	if _, ok := r.config[key]; ok {
		return true
	}
	if _, ok := r.nodes[key]; ok {
		return true
	}
	return false
}

func (r *reader) Read(key string) (interface{}, bool) {
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

func (r *reader) SetCfgSeparator(separator string) {
	r.cfgSeparator = strings.ToLower(separator)
	r.parse()
}

func (r *reader) SetFlagSeparator(separator string) {
	r.flagSeparator = separator
	r.parse()
}

func (r *reader) SetListSeparator(separator string) {
	r.listSeparator = separator
}

func (r *reader) cfgKey(flag string) string {
	path := strings.Split(flag, r.flagSeparator)
	return strings.ToLower(strings.Join(path, r.cfgSeparator))
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

func (r *reader) parse() {
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
				config[r.cfgKey(s[0])] = s[1]
			} else {
				key := r.cfgKey(arg)
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
			arg := arg[1:]
			if len(arg) > 1 && !strings.Contains(arg, "=") {
				// grouped short flags
				for sidx := 0; sidx < len(arg); sidx++ {
					if flag, ok := r.shortFlags[arg[sidx]]; ok {
						incrementFlag(config, r.cfgKey(flag))
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
				key := r.cfgKey(flag)
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
	nodes := map[string]bool{}
	for key := range config {
		path := strings.Split(key, r.cfgSeparator)
		nodePath := path[0]
		for idx := 1; idx < len(path); idx++ {
			nodes[nodePath] = true
			nodePath = nodePath + r.cfgSeparator + path[idx]
		}
	}
	r.config, r.nodes = config, nodes
}
