# config

A lightweight and versatile configuration toolkit for Go.

[![Build Status](https://travis-ci.org/warthog618/config.svg)](https://travis-ci.org/warthog618/config)
[![Coverage Status](https://coveralls.io/repos/github/warthog618/config/badge.svg?branch=master)](https://coveralls.io/github/warthog618/config?branch=master)
[![GoDoc](https://godoc.org/github.com/warthog618/config/sar?status.svg)](https://godoc.org/github.com/warthog618/config)
[![Go Report Card](https://goreportcard.com/badge/github.com/warthog618/config)](https://goreportcard.com/report/github.com/warthog618/config)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://github.com/warthog618/config/blob/master/LICENSE)

## Overview

**config** presents configuration from multiple sources as a unified key/value
store, providing a single consistent API to access configuration parameters,
independent of the underlying configuration storage formats, locations or
technologies.

**config** provides simple access to common configuration sources, including environment variables, command line flags, and configuration files (in JSON, INI, TOML, YAML, INI, and properties formats), and networked stores such as etcd.  And you can roll your own getters
if you have unusual requirements.

**config** provides functions to coax configuration values from the
configuration source into the types your code requires, including ints, string,
bool, time, duration, and slices.

**config** is versatile as it allows you to control all aspects of your
configuration, including the configuration sources, their location, format, and
the order in which they are searched.

### Quick Start

A couple of steps are required to setup and use **config**:

- Create a Config to provide type conversions for values from the your configuration sources
- Read and convert configuration Value from the Config

A minimal setup to access configuration from POSIX/GNU style command line flags,
might look like:

```go
    c := config.New(pflag.New())
```

A command line parameter such as

```bash
myapp --config-file=myfile.json
```

could then be read using:

```go
   cfgFile := c.MustGet("config.file").String()
```

Or read configuration from a configuration file - in this case a TOML file:

```go
    c := config.New(blob.New(file.New("myconfig.toml"), toml.NewDecoder()))
```

Multiple configuration sources can be added to the config, for example:

```go
    c := config.New(
        pflag.New(),
        env.New(),
        blob.New(file.New("myconfig.json"), json.NewDecoder()))
```

which would overlay configuration from POSIX flags over environment variables and over a JSON configuration file.  Gets from the configuration will search through the source in the provided order for each key.

Multiple configuration sources can be setup and customised to suit your
application requirements.  The [Example Usage](#usage) section provides
a more extensive example.

### API

[Config](https://godoc.org/github.com/warthog618/config#Config) provides get
methods to retrieve configuration parameters, identified by a key string, and
return them as a [Value](https://godoc.org/github.com/warthog618/config#Value).
The current value can be retrieved with a Get or MustGet.  Updates to a
particular key/value can be requested using a NewKeyWatcher.  Changes to the
complete configuration can be monitored using a NewWatcher.

The Value contains both the configuration value and an error handler to call
when converting the value to requested types.  By default the Value absorbs
conversion errors and returns the zero value for the requested type. e.g.

```go
    c := config.New(g)
    pin := c.MustGet("pin").Int()
```

will set *pin* to 0 if the configured value cannot be converted to int.

An error handling policy can be applied to the Config and Value using construction
options.  The error handling policy is inherited by Values and sub-Configs
returned by the Config. This allows errors to be directed to an error handler
rather than being handled where the functions return. e.g.

```go
    m := config.New(g, config.WithMust())
    v, _ := m.Get("pin")
    pin := v.Int()
    v,_ = m.Get("ports")
    ports := v.UintSlice()
```

will panic if either "pin" or "ports" are not configured or cannot be converted
to the requested type.

Dynamic changes to configuration can be monitored by adding a watcher, either on
a particular key, using
[NewKeyWatcher](https://godoc.org/github.com/warthog618/config#Config.NewKeyWatcher),
or the complete configuration using
[NewWatcher](https://godoc.org/github.com/warthog618/config#Config.NewWatcher).

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

## Background Concepts

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

    // equivalently....
    size := int(c.MustGet("ports[]").Int())
    for i := 0; i < size; i++ {
        // get each port sequentially...
        ports[i] = c.MustGet(fmt.Sprintf("ports[%d]", i)).Uint()
    }
```

### Config

As described in [API](#api), the
[Config](https://godoc.org/github.com/warthog618/config#Config) provides the API
to return Values from the configuration tree.  A Config may represent the root
of the configuration or a branch of the configuration tree - retreived from a
parent Config using GetConfig.

### Value

The configuration value is returned by the Config as a Value.  The
[Value](https://godoc.org/github.com/warthog618/config#Value) provides methods
to convert the value to the required types.

### Getters

The sources of configuration parameters are referred to as *getters*.

A getter must satisfy a simple interface:

```go
type Getter interface {
    // Get the value of the named config leaf key.
    // Also returns an ok, similar to a map read, to indicate if the value was
    // found.
    // The type underlying the returned interface{} must be convertable to the
    // expected type by cfgconv.
    // ...
    Get(key string) (value interface{}, found bool)
}
```

Refer to the documentation for a more complete definition of the Getter interface.

The Getter may optionally support the [WatchableGetter](https://godoc.org/github.com/warthog618/config#WatchableGetter) interface to indicate that
it supports being watched for changes.  This is typically enabled via a
construction option.

```go
type WatchableGetter interface {
    // Watcher returns the watcher for a Getter.
    // Returns false if the associated Getter does not support watches.
    Watcher() (GetterWatcher, bool)
}
```

If a Getter supports the WatchableGetter interface and indicates that it is watchable, by returning true form Watcher(), then the Config will monitor the Getter for changes and trigger its own Watchers and KeyWatchers when the underlying source changes.

The source of configuration may be local or remote.  Getters for remote configuration typically cache a snapshot of configuration locally.

A collection of getters can be formed into a
[Stack](https://godoc.org/github.com/warthog618/config#Stack).  A stack forms an
overlay of configuration parameters, the view from the top of which is presented
to the application as its configuration.  The getters contained in the stack,
and their order, is specified by the application and can be modified at runtime.

A number of getters for common configuration sources are provided in
sub-packages:

Getter | Configuration Source
:-----:| -----
[blob](https://github.com/warthog618/config/tree/master/blob) | files and other sources of formatted configuration in various formats including JSON, YAML, INI, and properties
[dict](https://github.com/warthog618/config/tree/master/dict) | key/value maps
[env](https://github.com/warthog618/config/tree/master/env) | environment variables
[etcd](https://github.com/warthog618/config/tree/master/etcd) | etcd v3 key/value server
[flag](https://github.com/warthog618/config/tree/master/flag) | Go style command line flags
[pflag](https://github.com/warthog618/config/tree/master/pflag) | POSIX/GNU style command line flags

Alternatively you can roll your own.

Several helper packages are available should you wish to roll your own
getter:

The [**keys**](https://github.com/warthog618/config/tree/master/keys)
sub-package provides a number of functions to assist in mapping between
namespaces.

The [**list**](https://github.com/warthog618/config/tree/master/list)
sub-package provides functions to assist in decoding lists stored as strings in
configuration sources.

The [**tree**](https://github.com/warthog618/config/tree/master/tree)
sub-package provides a Get method to get a value from a map[string]interface{}
or map[interface{}]interface{}.

### Decorators

Additionally, getters may be wrapped in decorators, such as the
[WithAlias](#alias) or [WithFallback](#fallback), to perform key translations
before the key is passed to the getter, or to manipulate the value before
returning it to the caller.

A number of decorators are provided including:

- [Alias](#alias)
- [Fallback](#fallback)
- [KeyReplacer](#keyreplacer)
- [Prefix](#prefix)
- [RegexAlias](#regexalias)
- [Trace](#trace)

Decorators can be added to the getter before it is provided to the Config. e.g.:

```go
    a := config.NewAlias()
    c := config.New(config.Decorate(g, config.WithAlias(a)))
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

#### Fallback

The [WithFallback](https://godoc.org/github.com/warthog618/config#WithFallback)
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
    c := config.New(config.Decorate(g, config.WithRegexAlias(r)))
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

## Examples

The following examples, and examples of more complex usage, can be found in the
[example](https://github.com/warthog618/config/tree/master/example) directory.

### Usage

The following is an example of setting up a config stacking a number of sources
(env, flag, JSON file, and a default map) and retrieving configuration
parameters of various types. The getters are added to a Stack so the search
order will be:

- flag
- env
- JSON file
- default map

Note that configuration from initial sources can be used when setting up
subsequent sources, e.g. the *env.prefix* can be overridden by flags
(**--env-prefix**), and the JSON config filename can be specified by either flag
(**-c** or **--config-file**) or env (**MYAPP_CONFIG_FILE**).

```go
func main() {
    defaultConfig := dict.New(dict.WithMap(map[string]interface{}{
        "name":          "myapp",
        "env.prefix":    "MYAPP_",
        "config.file":   "myapp.json",
        "sm.pin":        27,
        "sm.period":     "250ms",
        "sm.thresholds": []int8{23, 45, 64},
    }))
    // from flags and defaults...
    sources := config.NewStack(
        pflag.New(pflag.WithShortFlags(map[byte]string{'c': "config-file"})))
    cfg := config.New(sources, config.WithDefault(defaultConfig))

    // and from environment...
    prefix := cfg.MustGet("env.prefix").String()
    sources.Append(env.New(env.WithEnvPrefix(prefix)))

    // and from config file...
    cf := cfg.MustGet("config.file").String()
    sources.Append(blob.New(file.New(cf), json.NewDecoder(), blob.MustLoad()))

    // read a config field from the root config
    name := cfg.MustGet("name").String()

    // to pass nested config to a sub-module...
    smCfg := cfg.GetConfig("sm")
    pin := smCfg.MustGet("pin").Uint()
    period := smCfg.MustGet("period").Duration()
    thresholds := smCfg.MustGet("thresholds").IntSlice()

    fmt.Println(cf, name, pin, period, thresholds)

    // or using Unmarshal to populate a config struct...
    type SMConfig struct {
        Pin        uint
        Period     time.Duration
        Thresholds []int
    }
    sc := SMConfig{}
    cfg.Unmarshal("sm", &sc)
    fmt.Println(cf, name, sc)
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

### Key Watcher

This example watches a configuration file and prints updates to the configuration key "somevariable" whenever its value is changed.

```go
func main() {
    c := config.New(blob.New(
        file.New("config.json", file.WithWatcher()), json.NewDecoder()))

    done := make(chan struct{})
    defer close(done)
    // watcher goroutine
    go func() {
        w := c.NewKeyWatcher("somevariable")
        for {
            v, err := w.Watch(done)
            if err != nil {
                log.Println("watch error:", err)
                return
            }
            log.Println("got update:", v.Int())
        }
    }()
    // main thread
    time.Sleep(time.Minute)
    log.Println("finished.")
}
```

This is a simple example that minimises error handling for brevity.  The implementation of the watcher goroutine and its interactions with other goroutines may vary to suit your application.

A watcher for the whole configuration is very similar.  An example can be found in the [examples directory](https://github.com/warthog618/config/tree/master/example/readme/watcher).

## Future Work

A list of things I haven't gotten around to yet...

- Add more examples.
- Add a getter for consul.
