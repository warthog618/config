# config

A lightweight and versatile configuration toolkit for Go.

[![Build Status](https://travis-ci.org/warthog618/config.svg)](https://travis-ci.org/warthog618/config)
[![GoDoc](https://godoc.org/github.com/warthog618/config/sar?status.svg)](https://godoc.org/github.com/warthog618/config)
[![Go Report Card](https://goreportcard.com/badge/github.com/warthog618/config)](https://goreportcard.com/report/github.com/warthog618/config)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://github.com/warthog618/config/blob/master/LICENSE)

## Overview

**config** presents configuration as a unified key/value store, providing a single
consistent API to access configuration parameters, independent of the
underlying configuration storage formats, locations or technologies.

**config** is lightweight as it has no dependencies itself - your application
will only depend on the getters you explicitly include.  A collection of getters for common configuration sources is provided, each in its own subpackage, or you can roll your own.

**config** is versatile as it allows you to control all aspects of your configuration, including the configuration sources, their location, format,
and the order in which they are searched.

### Quick Start

A couple of steps are required to setup and use **config**:

- Create one or more getters (configuration sources)
- Create a Config to provide type conversions for values from the getter
- Read configuration from the Config

A minimal setup to access configuration from command line flags might look like:

```go
    flags, _ = flag.New()
    c := config.NewConfig(flags)
```

A command line parameter such as

```bash
myapp --config-file=myfile.json
```

could then be read using:

```go
   cfgFile := c.GetString("config.file")
```

Multiple configuration sources can be setup and customised to suit your application requirements.  The [Example Usage](#example-usage) section provides a more extensive example.

### API

Two flavours of API are provided through [Config](https://godoc.org/github.com/warthog618/config#Config) and [Must](https://godoc.org/github.com/warthog618/config#Must).

The [Config](https://godoc.org/github.com/warthog618/config#Config) provides an interface with
get methods to retrieve configuration parameters, identified by a key string,
and return them as the requested data type.

The get methods are similar to a map read, returning both the value and
an error, which indicates if the value could not be found or converted. e.g.

```go
    c := config.NewConfig(....)
    v,err := c.GetInt32("pin")
    ports,err := c.GetUintSlice("ports")
```

The [Must](https://godoc.org/github.com/warthog618/config#Must) provides a similar interface to Config, but does not return errors.  Rather than being returned to the caller, it allows errors to be ignored or directed to an error handler. e.g.

```go
    m := config.NewMust(config.WithPanic())
    v := m.GetInt32("pin")
    ports := m.GetUintSlice("ports")
```

will panic if either "pin" or "ports" are not configured.

Both flavours also provide methods to return the other flavour in case different sections of code use different error handling policies.

### Supported value types

**config** supports retrieving and returning configuration parameters as one of the following types:

- bool
- int (specifically *int64*)
- uint (specifically *uint64*)
- float (specifically *float64*)
- string
- slice (values remain as *interface{}*, so *[]interface{}*)
- slice of int (*[]int64*)
- slice of uint (*[]uint64*)
- slice of string (*[]string*)
- slice of struct (using *Unmarshal*)
- duration (*time.Duration*)
- time (*time.Time*)
- map (specifically *map[string]interface{}* using *UnmarshalToMap*)
- struct (using *Unmarshal*)

The int and float types return the maximum possible width to prevent loss of information.  The returned values can be range checked and assigned to
narrower types by the application as required.

The [**cfgconv**](https://godoc.org/github.com/warthog618/config/cfgconv) subpackage provides the functions **config** uses to perform the conversions from the *interface{}* returned by the getter to the type requested by the application code.  The **cfgconv** package is similar to the standard
[**strconv**](https://golang.org/pkg/strconv/) package, but converts from *interface{}* instead of *string*.  The conversions performed by **cfgconv** are as permissive as possible, given the data types involved, to allow for getters mapping from formats that may not directly support the requested type.

Direct gets of maps and structs are not supported, but both can be unmarshalled from the configuration, with the configuration keys being drawn from struct field names or map keys. Unmarshalling into nested structs is supported, as is overidding struct field names using tags.

## Concepts

### Config Tree

The configuration is presented to the application as a key/value store.  Conceptually the configuration parameters are located in a tree, where the key defines the path to the parameter from the root of the tree.  The key is a list of nodes followed by the name of the leaf.  The node and leaf names are joined with a separator, which by default is '.', to form the key.  e.g. *log.verbosity* idenfities the *verbosity* leaf in the *log* node.

Simple configurations may contain only a root node.  More complex configurations may include nodes corresponding to the configuration of contained objects or subsystems.

**config** does not enforce a particular case on keys, so applications can choose their preferred case.  Keys should be considered case sensitive by the application,
as **config** considers keys that differ only by case to be distinct.

Arrays, other than arrays of structs, are considered leaves and can be retrieved whole.  Additionally, array elements can be referenced with keys of the form *a[i]* where *a* is the key of the whole array and *i* is the zero-based integer index into the array.  The size of the array can be referenced with a key of form *a[]*. e.g.

```go
    ports := m.GetUIntSlice("ports")

    // alternatively....
    size := m.GetUInt("ports[]")
    for i := 0; i < size; i++ {
        // get each port sequentially...
        port := m.GetUint("ports[i]")
    }
```

### Config and Must

As described in [API](#api), the [Config](https://godoc.org/github.com/warthog618/config#Config) and [Must](https://godoc.org/github.com/warthog618/config#Must) provide the API to the configuration tree.  Both provide methods to return values from the configuration tree. The Config methods return the values and an error, while the Must methods return only the value and direct any errors to a configurable error handler.

### Getters

The sources of configuration parameters are referred to as *getters*.

A getter must safisfy a simple interface:

```go
type Getter interface {
    Get(key string) (value interface{}, found bool)
}
```

The source of configuration may be local or remote.

A number of getters for common configuration sources are provided in subpackages:

- [Environment](#environment)
- [Flag](#command-line-flags)
- [JSON](#json)
- [TOML](#toml)
- [YAML](#yaml)
- [Dictionary](#dictionary)
 
Alternatively you can roll your own.

A collection of getters can be formed into a [Stack](https://godoc.org/github.com/warthog618/config#Stack).  A stack forms an overlay of configuration parameters, the view from the top of which is presented to the application as its configuration.  The getters contained in the stack, and their order, is specified by the application and can be modified at runtime.

Additionally, getters may be wrapped in decorators, such as the
[WithAlias](#alias) or [WithDefault](#default),
to perform a key translations before the key is passed to the getter, or to manipulate the value before returning it to the caller.

The [**keys**](https://godoc.org/github.com/warthog618/config/keys) subpackage
provides a number of functions to assist in mapping between namespaces.

### Decorators

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
    c := config.NewConfig(config.Decorate(getter,config.WithAlias(a)))
    a.Add(newKey,oldKey)
```

Some of the decorators are described in more detail below. Refer to the [documentation](https://godoc.org/github.com/warthog618/config#Decorator) for the current set of provided decorators.

#### Alias

The [WithAlias](https://godoc.org/github.com/warthog618/config#WithAlias) decorator provides aliases that map from a new key to an old key.

When searching for configuration parameters at each getter, **config** first fetches the new key, and if that fails then tries any aliases to old keys.  Aliases are ignored by the search if the parameter is found with the new key.

Each new key may be aliased to multiple old keys, and old keys may be aliased by multiple new keys.

Aliases have a number of potential uses:

##### Migration

With aliases from a new layout to an old, the application can support both old and new configuration layouts at once, allowing users to upgrade their configurations at their convenience.  Once all users have migrated to the new layout, the aliases can be removed in subsequent releases.

##### DRY

A configuration parameter may be shared by multiple subsystems.  Rather than replicating the value throughout the configuration for each subsystem, an alias from one to the other allows a single parameter value to appear to be located in multiple places.

##### Overridable Defaults

A configuration parameter can be aliased to a default value, and will take the value from the alias (default) unless the parameter is explicitly set itself.  

This also allows default configuration values to be exposed in the configuration file rather than embedded in the application.

#### Default

The [WithDefault](https://godoc.org/github.com/warthog618/config#WithDefault) decorator provides a fallback getter to use if the configuration is not found in the decorated getter.

#### KeyReplacer

The [WithKeyReplacer](https://godoc.org/github.com/warthog618/config#WithKeyReplacer) decorator attaches a replacer which may performs a substitution on the key before it is presented to the getter.

#### Prefix

The [WithPrefix](https://godoc.org/github.com/warthog618/config#WithPrefix) decorator can be considered is a special case of WithKeyreplacer that prefixes the key with a fixed string.  This can be used to move a getter deeper into the configuration tree, for example if there is a configuration file for a particular subsystem.

#### RegexAlias

The [WithRegexAlias](https://godoc.org/github.com/warthog618/config#WithRegexAlias) decorator provides alias mappings similar to [Alias](#Alias), but the matching pattern is a regular expression instead of an exact match.

In addition to the uses of plain aliases, regex aliases can be used for setting default values for fields in array elements.  e.g. this alias

```go
    r := config.NewRegexAlias()
    r.Add("somearray\\[\\d+\\](.*)","somearray[0]$1")
    c := config.NewConfig(g,config.WithRegexAlias(r))
```

defaults all fields in the array elements to the values of the first element of the same array.

The regex form of alias requires more processing than plain aliases, and so is split out into a separate decorator.  If you don't need regexes then use the plain aliases instead.

#### Trace

The [WithTrace](https://godoc.org/github.com/warthog618/config#WithTrace) decorator attaches a function which is called with the parameters and return values of any call to the getter.  This could be used for logging and diagnostics, such as determining what configuration keys are retrieved by an application.

## Example Usage

The following is an example of setting up a config using a number of
sources  (env, flag, JSON file, and a default map) and retrieving
configuration parameters of various types.
The getters are added to a Stack so the search order will be:

- flag
- env
- JSON file
- default map

Note that configuration from initial sources can be used when setting up subsequent sources, e.g. the *env.prefix* can be overridden by flags, and the JSON config filename can be specified by either flag or env.

For brevity,this example omits error handling.

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
    g, _ = flag.New(flag.WithShortFlags(map[byte]string{'c': "config-file"}))
    sources := config.NewStack(g)
    cfg := config.NewConfig(
        config.Decorate(sources, config.WithDefault(defaultConfig)))
    prefix, _ := cfg.GetString("env.prefix")
    g, _ = env.New(env.WithEnvPrefix(prefix))
    sources.Append(g)
    cf, _ := cfg.GetString("config.file")
    g, _ = json.New(json.FromFile(cf))
    sources.Append(g)

    // read a config field from the root config
    name, _ := cfg.GetString("name")

    // to pass nested config to a sub-module...
    smCfg := cfg.GetMust("sm")
    pin := smCfg.GetInt("pin")
    period := smCfg.GetDuration("period")
    thresholds := smCfg.GetIntSlice("thresholds")

    fmt.Println(cf, name, pin, period, thresholds)
}
```

In this example, the config file name for *myapp*, with key *config.file*, could be set to "myfile.json" with any of these invocations:

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

This example, and examples of more complex usage, can be found in the examples directory.

## Supplied Getters

:construction: This section is still under construction....
The following getters are provided in sub-packages:

### Environment

[![GoDoc](https://godoc.org/github.com/warthog618/config/env/sar?status.svg)](https://godoc.org/github.com/warthog618/config/env)

The **env** package provides a getter that returns value from environment variables.  

### Command Line Flags

[![GoDoc](https://godoc.org/github.com/warthog618/config/flag/sar?status.svg)](https://godoc.org/github.com/warthog618/config/flag)

The **flag** package...

### JSON

[![GoDoc](https://godoc.org/github.com/warthog618/config/json/sar?status.svg)](https://godoc.org/github.com/warthog618/config/json)

The **json** package...

### YAML

[![GoDoc](https://godoc.org/github.com/warthog618/config/yaml/sar?status.svg)](https://godoc.org/github.com/warthog618/config/yaml)

The **yaml** package...

### TOML

[![GoDoc](https://godoc.org/github.com/warthog618/config/toml/sar?status.svg)](https://godoc.org/github.com/warthog618/config/toml)

The **toml** package...

### Properties

[![GoDoc](https://godoc.org/github.com/warthog618/config/properties/sar?status.svg)](https://godoc.org/github.com/warthog618/config/properties)

The properties package...

### Dictionary

[![GoDoc](https://godoc.org/github.com/warthog618/config/dict/sar?status.svg)](https://godoc.org/github.com/warthog618/config/dict)

The dict package...

## Future Work

A list of things I haven't gotten around to yet, or am still thinking about...

- Add more examples.
- Add a getter for etcd.
- Add a getter for consul.
- Add command line validation to flag.
- Add help generation to flag.
- Add watches on config sources for config changes
