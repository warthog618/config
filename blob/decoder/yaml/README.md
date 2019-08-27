# yaml

[![GoDoc](https://godoc.org/github.com/warthog618/config/blob/decoder/yaml/sar?status.svg)](https://godoc.org/github.com/warthog618/config/blob/decoder/yaml)

The **yaml** package provides a [config](https://github.com/warthog618/config) Decoder that unmarshals values from YAML formatted sources.

Example usage:

```go
import (
    "fmt"

    "github.com/warthog618/config"
    "github.com/warthog618/config/blob/decoder/yaml"
    "github.com/warthog618/config/blob/loader/file"
)

func main() {
    c := config.NewConfig(file.New("config.yaml", yaml.NewDecoder()))
    s := c.MustGet("nested.string").String()
    fmt.Println("s:", s)
    // ....
}
```
