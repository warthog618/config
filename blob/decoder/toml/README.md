# toml

[![GoDoc](https://godoc.org/github.com/warthog618/config/decoder/toml/sar?status.svg)](https://godoc.org/github.com/warthog618/config/decoder/toml)

The **toml** package provides a [config](https://github.com/warthog618/config) Decoder that unmarshals values from TOML formatted sources.

Example usage:

```go
import (
    "fmt"

    "github.com/warthog618/config"
    "github.com/warthog618/config/decoder/toml"
    "github.com/warthog618/config/loader/file"
)

func main() {
    f, _ := config.NewSource(file.New("config.toml"), toml.NewDecoder())
    c := config.NewConfig(f)
    s := c.MustGet("nested.string").String()
    fmt.Println("s:", s)
    // ....
}
```