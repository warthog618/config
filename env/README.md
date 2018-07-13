# env

[![GoDoc](https://godoc.org/github.com/warthog618/config/env/sar?status.svg)](https://godoc.org/github.com/warthog618/config/env)

The **env** package provides a getter for [config](https://github.com/warthog618/config) that returns values from environment variables.

The environment is assumed static and is only read when the env is constructed with New.

A number of options can be applied to the env:

The [WithEnvPrefix](https://godoc.org/github.com/warthog618/config/env#WithEnvPrefix) option limits the environment variables imported to those with a particular prefix.  The prefix is not included when considering if a environment variable matches a key.  Any prefix, if any, is application specific so there is no default.

The [WithKeyReplacer](https://godoc.org/github.com/warthog618/config/env#WithKeyReplacer) option performs a transformation on the environment variable name key before it is searched for.  This may be used, for example, to enforce a particular case.  The default key replacer replaces "_" with "." and forces all environment variable names to lower case, so "CONFIG_FILE" matches the key "config.file".

The [WithListSeparator](https://godoc.org/github.com/warthog618/config/env#WithListSeparator) option provides a string used to split list values into elements.  The default list separator is ":".