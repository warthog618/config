# config

A lightweight and versatile configuration toolkit for Go.

[![Build Status](https://travis-ci.org/warthog618/config.svg)](https://travis-ci.org/warthog618/config)
[![Coverage Status](https://coveralls.io/repos/github/warthog618/config/badge.svg?branch=master)](https://coveralls.io/github/warthog618/config?branch=master)
[![GoDoc](https://godoc.org/github.com/warthog618/config/sar?status.svg)](https://godoc.org/github.com/warthog618/config)
[![Go Report Card](https://goreportcard.com/badge/github.com/warthog618/config)](https://goreportcard.com/report/github.com/warthog618/config)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://github.com/warthog618/config/blob/master/LICENSE)

## Overview

**config** presents configuration as a unified key/value store, providing a
single consistent API to access configuration parameters, independent of the
underlying configuration storage formats, locations or technologies.

**config** is lightweight as it has no dependencies itself - your application
will only depend on the getters you explicitly include.  A collection of getters
for common configuration sources is provided, each in its own sub-package, or
you can roll your own.

**config** is versatile as it allows you to control all aspects of your
configuration, including the configuration sources, their location, format, and
the order in which they are searched.

### Quick Start

A couple of steps are required to setup and use **config**:

- Create one or more getters (configuration sources)
- Create a Config to provide type conversions for values from the getter
- Read configuration Value from the Config
- Convert the Value to the required type

A minimal setup to access configuration from POSIX/GNU style command line flags
might look like:

```go
    flags, _ = pflag.New()
    c := config.NewConfig(flags)
```

A command line parameter such as

```bash
myapp --config-file=myfile.json
```

could then be read using:

```go
   cfgFile := c.MustGet("config.file").String()
```

Multiple configuration sources can be setup and customised to suit your
application requirements.  The [Example Usage](#example-usage) section provides
a more extensive example.

### API

The [Config](https://godoc.org/github.com/warthog618/config#Config) provides the
primary interface to configuration.  Config provides get methods to retrieve
configuration parameters, identified by a key string, and return them as a
[Value](https://godoc.org/github.com/warthog618/config#Value).  The current
value can be retrieved with a Get or MustGet.  Updates to a value can be
requested using WatchValue.  Changes to the complete configuration can be monitored using a Watch

The Value contains both the configuration value and an error handler to call
when converting the value to requested types.  By default the Value absorbs
conversion errors and returns the zero value for the requested type. e.g.

```go
    c := config.NewConfig(g)
    pin := c.MustGet("pin").Int()
```

will set *pin* to 0 if the configured value cannot be converted to int.

An error handling policy can be applied to the Config and Value using construction
options.  The error handling policy is inherited by Values and sub-Configs
returned by the Config. This allows errors to be directed to an error handler
rather than being handled where the functions return. e.g.

```go
    m := config.NewConfig(g, config.WithMust())
    v, _ := m.Get("pin")
    pin := v.Int()
    v,_ = m.Get("ports")
    ports := v.UintSlice()
```

will panic if either "pin" or "ports" are not configured or cannot be converted to the requested type.

### Supported value types

**config** supports converting returned Values to the following types:

- bool
- int (specifically *int64*)
- uint (specifically *uint64*)
- float (specifically *float64*)
- string
- slice (values remain as *interface{}*, so *[]interface{}*)
- slice of int (*[]int64*)
- slice of uint (*[]uint64*)
- slice of string (*[]string*)
- duration (*time.Duration*)
- time (*time.Time*)

The int and float types return the maximum possible width to prevent loss of
information. The returned values can be range checked and assigned to narrower
types by the application as required.

The [**cfgconv**](https://godoc.org/github.com/warthog618/config/cfgconv)
sub-package provides the functions **config** uses to perform the conversions
from the *interface{}* returned by the getter to the type requested by the
application code. The **cfgconv** package is similar to the standard
[**strconv**](https://golang.org/pkg/strconv/) package, but converts from
*interface{}* instead of *string*.  The conversions performed by **cfgconv** are
as permissive as possible, given the data types involved, to allow for getters
mapping from formats that may not directly support the requested type.

Direct gets of maps and structs are not supported, but the following composite
types can be unmarshalled from the configuration, with the configuration keys
being drawn from struct field names or map keys:

- slice of struct (using *Unmarshal*)
- map (specifically *map[string]interface{}* using *UnmarshalToMap*)
- struct (using *Unmarshal*)

Unmarshalling into nested structs is supported, as is overiding struct field
names using tags.

## Concepts

### Config Tree

The configuration is presented to the application as a key/value store.
Conceptually the configuration parameters are located in a tree, where the key
defines the path to the parameter from the root of the tree.  The key is a list
of nodes followed by the name of the leaf.  The node and leaf names are joined
with a separator, which by default is '.', to form the key.  e.g.
*log.verbosity* identifies the *verbosity* leaf in the *log* node.

Simple configurations may contain only a root node.  More complex configurations
may include nodes corresponding to the configuration of contained objects or
subsystems.

**config** does not enforce a particular case on keys, so applications can
choose their preferred case.  Keys should be considered case sensitive by the
application, as **config** considers keys that differ only by case to be
distinct.

Arrays, other than arrays of structs, are considered leaves and can be retrieved
whole. Additionally, array elements can be referenced with keys of the form
*a[i]* where *a* is the key of the whole array and *i* is the zero-based integer
index into the array.  The size of the array can be referenced with a key of
form *a[]*. e.g.

```go
    ports := c.MustGet("ports").UintSlice()

    // alternatively....
    // alternatively....
    size := int(c.MustGet("ports[]").Int())
    for i := 0; i < size; i++ {
        // get each port sequentially...
        ports[i] = c.MustGet(fmt.Sprintf("ports[%d]", i)).Uint()
```

### Config

As described in [API](#api), the
[Config](https://godoc.org/github.com/warthog618/config#Config) provides the API to
return Values from the configuration tree.  A Config may represent the root of the configuration or a branch of the configuration tree - retreived from a parent Config using GetConfig.

### Value

The configuration value is returned by the Config as a Value.  The
[Value](https://godoc.org/github.com/warthog618/config#Value) provides methods
to convert the value to the required types.

### Getters

The sources of configuration parameters are referred to as *getters*.

A getter must satisfy a simple interface:

```go
type Getter interface {
    Get(key string) (value interface{}, found bool)
}
```

The source of configuration may be local or remote.

A collection of getters can be formed into a
[Stack](https://godoc.org/github.com/warthog618/config#Stack).  A stack forms an
overlay of configuration parameters, the view from the top of which is presented
to the application as its configuration.  The getters contained in the stack,
and their order, is specified by the application and can be modified at runtime.

A number of getters for common configuration sources are provided in
sub-packages:

Getter | Configuration Source
:-----:| -----
[dict](https://github.com/warthog618/config/tree/master/dict) | key/value maps
[env](https://github.com/warthog618/config/tree/master/env) | environment variables
[flag](https://github.com/warthog618/config/tree/master/flag) | Go style command line flags
[json](https://github.com/warthog618/config/tree/master/json) | JSON files or other JSON formatted sources
[pflag](https://github.com/warthog618/config/tree/master/pflag) | POSIX/GNU style command line flags
[properties](https://github.com/warthog618/config/tree/master/properties) | Properties files or other properties formatted sources
[toml](https://github.com/warthog618/config/tree/master/toml) | TOML files or other TOML formatted sources
[yaml](https://github.com/warthog618/config/tree/master/yaml) | YAML files or other YAML formatted sources

Alternatively you can roll your own.

A couple of helper packages are available should you wish to roll your own
getter:

The [**keys**](https://github.com/warthog618/config/tree/master/keys)
sub-package provides a number of functions to assist in mapping between
namespaces.

The [**tree**](https://github.com/warthog618/config/tree/master/tree)
sub-package provides a Get method to get a value from a map[string]interface{}
or map[interface{}]interface{}.

### Decorators

Additionally, getters may be wrapped in decorators, such as the
[WithAlias](#alias) or [WithDefault](#default), to perform a key translations
before the key is passed to the getter, or to manipulate the value before
returning it to the caller.

A number of decorators are provided including:

- [Alias](#alias)
- [Default](#default)
- [KeyReplacer](#keyreplacer)
- [Prefix](#prefix)
- [RegexAlias](#regexalias)
- [Trace](#trace)

Decorators are added to the getter before it is provided to the Config. e.g.:

```go
    a := config.NewAlias()
    c := config.NewConfig(config.Decorate(g, config.WithAlias(a)))
    a.Append(newKey, oldKey)
```

Some of the decorators are described in more detail below. Refer to the
[Decorator
documentation](https://godoc.org/github.com/warthog618/config#Decorator) for the
current set of provided decorators.

#### Alias

The [WithAlias](https://godoc.org/github.com/warthog618/config#WithAlias)
decorator provides aliases that map from a new key to an old key.

When searching for configuration parameters at each getter, **config** first
fetches the new key, and if that fails then tries any aliases to old keys.
Aliases are ignored by the search if the parameter is found with the new key.

Each new key may be aliased to multiple old keys, and old keys may be aliased by
multiple new keys.

Aliases have a number of potential uses:

##### Migration

With aliases from a new layout to an old, the application can support both old
and new configuration layouts at once, allowing users to upgrade their
configurations at their convenience.  Once all users have migrated to the new
layout, the aliases can be removed in subsequent releases.

##### DRY

A configuration parameter may be shared by multiple subsystems.  Rather than
replicating the value throughout the configuration for each subsystem, an alias
from one to the other allows a single parameter value to appear to be located in
multiple places.

##### Overridable Defaults

A configuration parameter can be aliased to a default value, and will take the
value from the alias (default) unless the parameter is explicitly set itself.  

This also allows default configuration values to be exposed in the configuration
file rather than embedded in the application.

#### Default

The [WithDefault](https://godoc.org/github.com/warthog618/config#WithDefault)
decorator provides a fallback getter to use if the configuration is not found in
the decorated getter.

#### KeyReplacer

The
[WithKeyReplacer](https://godoc.org/github.com/warthog618/config#WithKeyReplacer)
decorator attaches a replacer which may performs a substitution on the key
before it is presented to the getter.

#### Prefix

The [WithPrefix](https://godoc.org/github.com/warthog618/config#WithPrefix)
decorator can be considered is a special case of WithKeyReplacer that prefixes
the key with a fixed string.  This can be used to move a getter deeper into the
configuration tree, for example if there is a configuration file for a
particular subsystem.

#### RegexAlias

The
[WithRegexAlias](https://godoc.org/github.com/warthog618/config#WithRegexAlias)
decorator provides alias mappings similar to [Alias](#Alias), but the matching
pattern is a regular expression instead of an exact match.

In addition to the uses of plain aliases, regex aliases can be used for setting
default values for fields in array elements.  e.g. this alias

```go
    r := config.NewRegexAlias()
    r.Append(`somearray\[\d+\](.*)`, "somearray[0]$1")
    c := config.NewConfig(config.Decorate(g, config.WithRegexAlias(r)))
```

defaults all fields in the array elements to the values of the first element of
the same array.

The regex form of alias requires more processing than plain aliases, and so is
split out into a separate decorator.  If you don't need regexes then use the
plain aliases instead.

#### Trace

The [WithTrace](https://godoc.org/github.com/warthog618/config#WithTrace)
decorator attaches a function which is called with the parameters and return
values of any call to the getter.  This could be used for logging and
diagnostics, such as determining what configuration keys are retrieved by an
application.

## Example Usage

The following is an example of setting up a config using a number of sources
(env, flag, JSON file, and a default map) and retrieving configuration
parameters of various types. The getters are added to a Stack so the search
order will be:

- flag
- env
- JSON file
- default map

Note that configuration from initial sources can be used when setting up
subsequent sources, e.g. the *env.prefix* can be overridden by flags, and the
JSON config filename can be specified by either flag or env.

For brevity, this example omits error handling.

```go
func main() {
    defaultConfig := dict.New(dict.WithMap(map[string]interface{}{
        "name":          "myapp",
        "env.prefix":    "MYAPP_",
        "config.file":   "myapp.json",
        "sm.pin":        27,
        "sm.period":     "250ms",
        "sm.thresholds": "23,45,64",
    }))
    var g config.Getter
    g, _ = pflag.New(pflag.WithShortFlags(map[byte]string{'c': "config-file"}))
    sources := config.NewStack(g)
    cfg := config.NewConfig(
        config.Decorate(sources, config.WithDefault(defaultConfig)))
    prefix := cfg.MustGet("env.prefix").String()
    g, _ = env.New(env.WithEnvPrefix(prefix))
    sources.Append(g)
    cf := cfg.MustGet("config.file").String()
    g, _ = json.New(json.FromFile(cf))
    sources.Append(g)

    // read a config field from the root config
    name := cfg.MustGet("name").String()

    // to pass nested config to a sub-module...
    smCfg := cfg.GetConfig("sm")
    pin := smCfg.MustGet("pin").Uint()
    period := smCfg.MustGet("period").Duration()
    thresholds, _ := smCfg.Get("thresholds")

    fmt.Println(cf, name, pin, period, thresholds)
}
```

In this example, the config file name for *myapp*, with key *config.file*, could
be set to "myfile.json" with any of these invocations:

```sh
# short flag
myapp -c myfile.json

# long flag
myapp --config-file=myfile.json

# environment
MYAPP_CONFIG_FILE="myfile.json" myapp

# environment with overridden prefix
APP_CONFIG_FILE="myfile.json" myapp --env.prefix=APP_
```

This example, and examples of more complex usage, can be found in the examples
directory.

## Future Work

A list of things I haven't gotten around to yet, or am still thinking about...

- Add more examples.
- Add a getter for etcd.
- Add a getter for consul.
- Add watches on config for changes (already reworked Config API
  to support this - still require support from the Getters)
- Refactor Getters into Loader/Decoder, and allow Loader to notify Config of updates.
