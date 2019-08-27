# env

[![GoDoc](https://godoc.org/github.com/warthog618/config/env/sar?status.svg)](https://godoc.org/github.com/warthog618/config/env)

The **env** package provides a [config](https://github.com/warthog618/config) Getter that returns values from environment variables.

The environment is assumed static and is only read when the env is constructed with New.

```go
import (
    "fmt"

    "github.com/warthog618/config"
    "github.com/warthog618/config/env"
)

func main() {
    c := config.NewConfig(env.New())
    cfgFile, _ := c.GetString("config.file")
    fmt.Println("config-file:", cfgFile)
    // ....
}
```

A number of options can be applied to env.New:

The
[WithEnvPrefix](https://godoc.org/github.com/warthog618/config/env#WithEnvPrefix)
option limits the environment variables imported to those with a particular
prefix.  The prefix is not included when considering if a environment variable
matches a key.  The prefix, if any, is application specific so there is no
default.

The
[WithKeyReplacer](https://godoc.org/github.com/warthog618/config/env#WithKeyReplacer)
option performs a transformation on the environment variable name key before it
is added to the configuration.  This may be used, for example, to enforce a
particular case.  The default key replacer replaces "_" with "." and forces all
environment variable names to lower case, so the environment variable
"CONFIG_FILE" matches the key "config.file".

The
[WithListSeparator](https://godoc.org/github.com/warthog618/config/env#WithListSeparator)
option provides a string used to split list values into elements.  The default
list separator is ":".
