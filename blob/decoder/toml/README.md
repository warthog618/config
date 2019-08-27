# toml

[![GoDoc](https://godoc.org/github.com/warthog618/config/blob/decoder/toml/sar?status.svg)](https://godoc.org/github.com/warthog618/config/blob/decoder/toml)

The **toml** package provides a [config](https://github.com/warthog618/config) Decoder that unmarshals values from TOML formatted sources.

Example usage:

```go
import (
    "fmt"

    "github.com/warthog618/config"
    "github.com/warthog618/config/blob/decoder/toml"
    "github.com/warthog618/config/blob/loader/file"
)

func main() {
    c := config.NewConfig(file.New("config.toml", toml.NewDecoder()))
    s := c.MustGet("nested.string").String()
    fmt.Println("s:", s)
    // ....
}
```
