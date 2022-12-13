# config

A lightweight and versatile configuration toolkit for Go.

[![Build Status](https://img.shields.io/github/workflow/status/warthog618/config/Go.svg?logo=github)](https://github.com/warthog618/config/actions)
[![GoDoc](https://godoc.org/github.com/warthog618/config/sar?status.svg)](https://godoc.org/github.com/warthog618/config)
[![Go Report Card](https://goreportcard.com/badge/github.com/warthog618/config)](https://goreportcard.com/report/github.com/warthog618/config)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://github.com/warthog618/config/blob/master/LICENSE)

## Overview

**config** presents configuration from multiple sources as a unified key/value
store, providing a single consistent API to access configuration parameters,
independent of the underlying configuration storage formats, locations or
technologies.

**config** provides simple access to common configuration sources, including environment variables, command line flags, and configuration files (in JSON, INI, TOML, YAML, INI, and properties formats), and networked stores such as etcd.  And you can roll your own Getters
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

## Core API

The core API is comprised of three elements:

- the [Getter](#getter) which retrieves configuration values from an underlying
  data source
- the [Config](#config) which wraps a Getter or collection of Getters
- the [Value](#value) corresponding to a particular key which is returned by the
  Config.

The core API and supplied Getters should be sufficient for common usage.

### Config

[![GoDoc](https://godoc.org/github.com/warthog618/config/sar?status.svg)](https://godoc.org/github.com/warthog618/config#Config)

The Config provides get methods to retrieve configuration parameters, identified
by a key string, and return them as a
[Value](#value).

[Getters](#getter) are added to the Config by passing them to
[config.New](https://godoc.org/github.com/warthog618/config#New), or by later
adding them using
[Config.Append](https://godoc.org/github.com/warthog618/config#Config.Append),
which adds a Getter to the end of the list of Getters to he searched, or
[Config.Insert](https://godoc.org/github.com/warthog618/config#Config.Insert)
which inserts a Getter to the front of the list of Getters.

Values can be retrieved using
[Config.Get](https://godoc.org/github.com/warthog618/config#Config.Get), which
returns the Value and any error that occured while retrieving it, or
[Config.MustGet](https://godoc.org/github.com/warthog618/config#Config.MustGet),
which returns the Value or panics if there was an error.

Complete objects can be retrieved using
[Config.Unmarshal](https://godoc.org/github.com/warthog618/config#Config.Unmarshal), or
[Config.UnmarshalToMap](https://godoc.org/github.com/warthog618/config#Config.UnmarshalToMap).

### Getter

[![GoDoc](https://godoc.org/github.com/warthog618/config/sar?status.svg)](https://godoc.org/github.com/warthog618/config#Getter)

The Getter retrieves configuration values from an underlying data store.

The source of configuration may be local or remote.  Getters for remote configuration typically cache a snapshot of configuration locally, but can be optionally monitored for changes using a [Watcher](#watchers).

A number of Getters for common configuration sources are provided in
sub-packages:

Getter | Configuration Source
:-----:| -----
[blob](https://github.com/warthog618/config/tree/master/blob) | files and other sources of formatted configuration in various formats including JSON, YAML, INI, and properties
[dict](https://github.com/warthog618/config/tree/master/dict) | key/value maps
[env](https://github.com/warthog618/config/tree/master/env) | environment variables
[etcd](https://github.com/warthog618/config/tree/master/etcd) | etcd v3 key/value server
[flag](https://github.com/warthog618/config/tree/master/flag) | Go style command line flags
[pflag](https://github.com/warthog618/config/tree/master/pflag) | POSIX/GNU style command line flags

If those are insufficient, you can roll your own Getter.  Refer to the
[Getter](https://godoc.org/github.com/warthog618/config#Getter) documentation
for a more complete definition of the interface to implement.

The Getter may optionally support the [Option](https://godoc.org/github.com/warthog618/config#Option) interface so that it can be passed into [config.New](https://godoc.org/github.com/warthog618/config#New) with a collection of other Getters.  All the supplied Getters support the Option interface by embedding a [GetterAsOption](https://godoc.org/github.com/warthog618/config#GetterAsOption) element.

Several helper packages are available should you wish to roll your own Getter:

The [**keys**](https://github.com/warthog618/config/tree/master/keys)
sub-package provides a number of functions to assist in mapping between
namespaces.

The [**list**](https://github.com/warthog618/config/tree/master/list)
sub-package provides functions to assist in decoding lists stored as strings in
configuration sources.

The [**tree**](https://github.com/warthog618/config/tree/master/tree)
sub-package provides a Get method to get a value from a map[string]interface{}
or map[interface{}]interface{}.

### Value

[![GoDoc](https://godoc.org/github.com/warthog618/config/sar?status.svg)](https://godoc.org/github.com/warthog618/config#Value)

The Value contains the raw value for a field, as returned by the
[Getter](#getter) and [Config](#config), and provides methods to convert the raw
value into particular types.

#### Supported target types

**config** provides methods to convert Values to the following types:

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
from the *interface{}* returned by the Getter to the type requested by the
application code. The **cfgconv** package is similar to the standard
[**strconv**](https://golang.org/pkg/strconv/) package, but converts from
*interface{}* instead of *string*.  The conversions performed by **cfgconv** are
as permissive as possible, given the data types involved, to allow for Getters
mapping from formats that may not directly support the requested type.

Direct gets of maps and structs are not supported, but the following composite
types can be unmarshalled from the configuration, with the configuration keys
being drawn from struct field names or map keys:

- slice of struct (using *Unmarshal*)
- map (specifically *map[string]interface{}* using *UnmarshalToMap*)
- struct (using *Unmarshal*)

Unmarshalling into nested structs is supported, as is overiding struct field
names using tags.

## Advanced API

The intent is for the core API to handle the majority of use cases, but the
advanced sections of the API provide additional functionality for more
complicated environments.

### Sub-config

A sub-tree of the configuration contained in a Config, such as the configuration specific to a sub-module, can be retreived from a
parent Config using [Config.GetConfig](https://godoc.org/github.com/warthog618/config#Config.GetConfig).

For example the configuration for a Postgress client may be contained in a tree under "db.postgres" in my application configuration.  This could be retreived using:

```go
    pgCfg := cfg.GetConfig("db.postgres")
    pg := pgClient.New(pgCfg)
```

The sub-module, in this case the postgress client, can then be presented with its configuration without any knowledge of the application in which it is contained.

### Watchers

Dynamic changes to configuration can be monitored by adding a watcher, either on
a particular key, using
[Config.NewKeyWatcher](https://godoc.org/github.com/warthog618/config#Config.NewKeyWatcher),
or the complete configuration using
[Config.NewWatcher](https://godoc.org/github.com/warthog618/config#Config.NewWatcher).

Getters may optionally support the
[WatchableGetter](https://godoc.org/github.com/warthog618/config#WatchableGetter)
interface to indicate that it supports monitoring the underlying source for
changes.  This is typically enabled via a Getter construction option called WithWatcher.

Of the supplied Getters, only [file](https://godoc.org/github.com/warthog618/config/blob/loader/file) loader and the [etcd](https://godoc.org/github.com/warthog618/config/etcd)
currently support watchers.

### Error Handling Policy

The default error handling behaviour of the core API commands is as follows:

API Method|Error Behaviour
-----|-----
Config.Get|Get error returned with zero Value.
Config.MustGet|Get error causes panic :boom:.
Value.X (conversions)|Conversion error returns the zero value for target type.

For example the following code will panic if the *pin* config is not found, but will return 0 if it is found but cannot be converted to the int as expected:

```go
    c := config.New(getter)
    pin := c.MustGet("pin").Int()
```

These behaviours can be overriden globally to the Config using ConfigOptions, and per Get using ValueOptions.  

Option|ConfigOption|ValueOption|Behaviour
----|----|----|----
WithMust|&#10003;|&#10003;|Panic :boom: on error.
WithErrorHandler|&#10003;|&#10003;|Install error handler for both Config.Get and Value conversion errors.  When applied to a Config, the error handler is propagated to any Values created by the Config.Get unless overridden by ValueOptions in the Get.
WithGetErrorHandler|&#10003;|&#10060;|Install handler for Config.Get errors
WithValueErrorHandler|&#10003;|&#10060;|Install handler for Value conversion errors.  This option is passed to the Config.New and is propagated to any Values created by the Config.Get unless overridden by ValueOptions in the Get.

[Error handlers](https://godoc.org/github.com/warthog618/config#ErrorHandler)
are passed the error, which they process as appropriate, and return an error
that replaces the original.  This may be nil if the handler wants to absorb the
error.  This is particularly relevent for Config.MustGet, which wraps Config.Get
and converts any error to a panic, as the error returned by the get error
handler is the error checked by Config.MustGet.

### Overlays

A collection of Getters can be formed into an
[Overlay](https://godoc.org/github.com/warthog618/config#Overlay).  An overlay
presents an ordered list of Getters as a new Getter.  The overlay searches for
the key in each of the Getters in turn and returns the first value found.  The
Getters contained in the overlay, and their order, is specified by the
application and is fixed during construction.

The Overlay can be considered an immutable Stack.

### Stacks

A collection of Getters can be formed into a
[Stack](https://godoc.org/github.com/warthog618/config#Stack).  A stack presents
an ordered list of Getters as a new Getter. The stack searches for the key in
each of the Getters in turn and returns the first value found. Additional
Getters can be safely added to either end of the Stack at runtime.

The Stack can be considered a mutable Overlay.

### Decorators

Getters may be wrapped in
[Decorators](https://godoc.org/github.com/warthog618/config#Decorator), such as
the [WithAlias](#alias) or [WithFallback](#fallback), to perform key
translations before the key is passed to the Getter, or to manipulate the value
before returning it to the Config.

A number of decorators are provided including:

Decorator | Purpose
:-----:| -----
[Alias](#alias)|Map a key that does not exist in the configuation to one that does
[Fallback](#fallback)|Provide a fallback Getter to be used when a key is not found in the decorated Getter
[Graft](#graft)|Graft the root of a Getter that only provides a sub-config into the config
[KeyReplacer](#keyreplacer)|Perform string replacements on keys before they are passed to the decorated Getter
[MustGet](#mustget)|Panic if the key is not found in the decorated Getter
[Prefix](#prefix)|Add a prefix to keys be for passing them to the decorated Getter
[RegexAlias](#regexalias)|Map all keys that match a regex to a fixed key
[Trace](#trace)|Pass values returned from the Getter to a provided function
[UpdateHandler](#updatehandler)|Perform transformations on values returned by a GetterWatcher

Decorators can be added to the Getter before it is provided to the Config. e.g.:

```go
    a := config.NewAlias()
    c := config.New(config.Decorate(g, config.WithAlias(a)))
    a.Append(newKey, oldKey)
```

#### Alias

The [WithAlias](https://godoc.org/github.com/warthog618/config#WithAlias)
decorator provides aliases that map from a new key to an old key.

When searching for configuration parameters at each Getter, **config** first
fetches the new key, and if that fails then tries any aliases to old keys.
Aliases are ignored by the search if the parameter is found with the new key.

Each new key may be aliased to multiple old keys, and old keys may be aliased by
multiple new keys.

Aliases have a number of potential uses:

- Migration

    With aliases from a new layout to an old, the application can support both old
    and new configuration layouts at once, allowing users to upgrade their
    configurations at their convenience.  Once all users have migrated to the new
    layout, the aliases can be removed in subsequent releases.

- DRY

    A configuration parameter may be shared by multiple subsystems.  Rather than
    replicating the value throughout the configuration for each subsystem, an alias
    from one to the other allows a single parameter value to appear to be located in
    multiple places.

- Overridable Defaults

    A configuration parameter can be aliased to a default value, and will take the
    value from the alias (default) unless the parameter is explicitly set itself.  

    This also allows default configuration values to be exposed in the configuration
    file rather than embedded in the application.

#### Fallback

The [WithFallback](https://godoc.org/github.com/warthog618/config#WithFallback)
decorator provides a fallback Getter to use if the configuration is not found in
the decorated Getter.

#### Graft

The [WithGraft](https://godoc.org/github.com/warthog618/config#WithGraft)
decorator relocates the root of the decorated Getter into a node of the
configuration tree.  This allows a Getter that only provides part of the
configuration tree to be grafted into the larger tree.

#### KeyReplacer

The
[WithKeyReplacer](https://godoc.org/github.com/warthog618/config#WithKeyReplacer)
decorator attaches a replacer which may performs a substitution on the key
before it is presented to the Getter.

#### MustGet

The [WithMustGet](https://godoc.org/github.com/warthog618/config#WithMust)
decorator panics if the key cannot be found by the Getter.

This takes priority over the global error handling performed by the Config, as its checking effectively occurs before the Get returns to the Config.

#### Prefix

The [WithPrefix](https://godoc.org/github.com/warthog618/config#WithPrefix)
decorator can be considered is a special case of WithKeyReplacer that prefixes
the key with a fixed string.  This can be used to move a Getter deeper into the
configuration tree, for example if the configuration only requires a sub-config
from a larger configuration.

This is the opposite of [Graft](#graft).

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
values of any call to the Getter.  This could be used for logging and
diagnostics, such as determining what configuration keys are retrieved by an
application.

#### UpdateHandler

The
[WithUpdateHandler](https://godoc.org/github.com/warthog618/config#WithUpdateHandler)
decorator adds a handler function to a Getter that can process updates from the
Getter before they are returned to the watcher.  This allows any decoration that
would normally be applied to the Get path to be applied to the watcher path.

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

## Examples

The following examples, and examples of more complex usage, can be found in the
[example](https://github.com/warthog618/config/tree/master/example) directory.

### Usage

The following is an example of setting up a configuration using a number of
sources (env, flag, JSON file, and a default map) and retrieving configuration
parameters of various types. The Getters are added to a Stack so the search
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
    cfg := config.New(
        pflag.New(pflag.WithFlags([]pflag.Flag{{Short: 'c', Name: "config-file"}})),
        config.WithDefault(defaultConfig))

    // and from environment...
    prefix := cfg.MustGet("env.prefix").String()
    cfg.Append(env.New(env.WithEnvPrefix(prefix)))

    // and from config file...
    cf := cfg.MustGet("config.file").String()
    cfg.Append(blob.New(file.New(cf), json.NewDecoder(), blob.MustLoad()))

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
    c := config.New(blob.New(file.New("config.json", file.WithWatcher()), json.NewDecoder()))

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
