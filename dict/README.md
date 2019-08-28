# dict

[![GoDoc](https://godoc.org/github.com/warthog618/config/dict/sar?status.svg)](https://godoc.org/github.com/warthog618/config/dict)

The **dict** package provides a [config](https://github.com/warthog618/config) Getter that retrieves values from a key/value map.

The dict is typically used to define default values.

Example usage:

```go
import (
    "fmt"

    "github.com/warthog618/config"
    "github.com/warthog618/config/dict"
)

func main() {
    d := dict.New()
    d.Set("config.file", "config.json")
    c := config.New(d)
    cfgFile := c.MustGet("config.file").String()
    fmt.Println("config-file:", cfgFile)
    // ....
}
```

One option can be applied to dict.New:

The [WithMap](https://godoc.org/github.com/warthog618/config/dict#WithMap)
option provides a map to be used instead of creating a new empty map.  Not that
the dict takes ownership of the map and any subsequent sets will alter the
underlying map.
