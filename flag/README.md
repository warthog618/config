# flag

[![GoDoc](https://godoc.org/github.com/warthog618/config/flag/sar?status.svg)](https://godoc.org/github.com/warthog618/config/flag)

The **flag** package provides a [config](https://github.com/warthog618/config) Getter that retrieves values from the flags provided by the builtin Go flag package.

The package assumes that Go flag has been configured and flag.Parse() called before the config flag getter is created with New.

Example usage:

```go
import (
    "flag"
    "fmt"

    "github.com/warthog618/config"
    cfgflag "github.com/warthog618/config/flag"
)

func main() {
    flag.String("config-file", "config.json", "config file name")
    flag.Parse()
    c := config.NewConfig(cfgflag.New())
    cfgFile := c.MustGet("config.file").String()
    fmt.Println("config-file:", cfgFile)
    // ....
}
```

A number of options can be applied to flag.New:

The
[WithAllFlags](https://godoc.org/github.com/warthog618/config/flag#WithAllFlags)
option retrieves all flags from flag, including those not explicitly set on the
command line.  This results in the default values specified for those flags
being returned by the flag getter.  By default only the flags set on the command
line are added to the configuration.

The
[WithKeyReplacer](https://godoc.org/github.com/warthog618/config/flag#WithKeyReplacer)
option performs a transformation on the flag name before it is added to the
configuration.  This may be used, for example, to enforce a particular case.
The default key replacer replaces "-" with ".", so the flag "config-file"
matches the key "config.file".

The
[WithListSeparator](https://godoc.org/github.com/warthog618/config/flag#WithListSeparator)
option provides a string used to split list values into elements.  The default
list separator is ",".
