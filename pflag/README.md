# pflag

[![GoDoc](https://godoc.org/github.com/warthog618/config/pflag/sar?status.svg)](https://godoc.org/github.com/warthog618/config/pflag)

The **pflag** package provides a [config](https://github.com/warthog618/config) Getter that retrieves values from POSIX/GNU style command line flags.

Unlike the builtin Go flag package, the pflag package does not require registration of potential flags.  Any flags provided to the command line are automatically added to the configuration.

Example usage:

```go
import (
    "fmt"

    "github.com/warthog618/config"
    "github.com/warthog618/config/pflag"
)

func main() {
    f, _ := pflag.New()
    c := config.NewConfig(f)
    cfgFile, _ := c.GetString("config.file")
    fmt.Println("config-file:", cfgFile)
    // ....
}
```

A number of options can be applied to pflag.New:

The [WithCommandLine](https://godoc.org/github.com/warthog618/config/pflag#WithCommandLine) option provides an alternate command line, defined as a list of args, to be parsed.  By default the command line is *os.Args[1:]*.

The [WithKeyReplacer](https://godoc.org/github.com/warthog618/config/pflag#WithKeyReplacer) option performs a transformation on the flag name before it is added to the configuration.  This may be used, for example, to enforce a particular case.  The default key replacer replaces "-" with ".", so the flag "config-file" matches the key "config.file".

The [WithListSeparator](https://godoc.org/github.com/warthog618/config/pflag#WithListSeparator) option provides a string used to split list values into elements.  The default list separator is ",".

The [WithShortFlags](https://godoc.org/github.com/warthog618/config/pflag#WithShortFlags) option provides a set of short flags which provide short aliases for long flag names. When set using the short form the values still appear in the configuration as if it were set with the long name.

## Future Work

- Add command line validation.
- Add help generation.
