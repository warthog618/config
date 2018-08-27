# yaml

[![GoDoc](https://godoc.org/github.com/warthog618/config/decoder/yaml/sar?status.svg)](https://godoc.org/github.com/warthog618/config/decoder/yaml)

The **yaml** package provides a [config](https://github.com/warthog618/config) Decoder that unmarshals values from YAML formatted sources.

Example usage:

```go
import (
    "fmt"

    "github.com/warthog618/config"
    "github.com/warthog618/config/decoder/yaml"
    "github.com/warthog618/config/loader/file"
)

func main() {
    f, _ := config.NewSource(file.New("config.yaml"), yaml.NewDecoder())
    c := config.NewConfig(f)
    s := c.MustGet("nested.string").String()
    fmt.Println("s:", s)
    // ....
}
```