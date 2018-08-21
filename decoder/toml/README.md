# toml

[![GoDoc](https://godoc.org/github.com/warthog618/config/toml/sar?status.svg)](https://godoc.org/github.com/warthog618/config/toml)

The **toml** package provides a [config](https://github.com/warthog618/config) Getter that retrieves values from TOML files or other TOML formatted sources.

Example usage:

```go
import (
    "fmt"

    "github.com/warthog618/config"
    "github.com/warthog618/config/toml"
)

func main() {
    f, _ := toml.New(toml.FromFile("config.toml"))
    c := config.NewConfig(f)
    s, _ := c.GetString("nested.string")
    fmt.Println("s:", s)
    // ....
}
```

A number of options can be applied to toml.New:

The [FromBytes](https://godoc.org/github.com/warthog618/config/toml#FromBytes) option reads the TOML from a byte array.

The [FromFile](https://godoc.org/github.com/warthog618/config/toml#FromFile) option reads the TOML from a file.

The [FromReader](https://godoc.org/github.com/warthog618/config/toml#FromReader) option reads the TOML from an io.Reader.