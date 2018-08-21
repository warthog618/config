# json

[![GoDoc](https://godoc.org/github.com/warthog618/config/decoder/json/sar?status.svg)](https://godoc.org/github.com/warthog618/config/decoder/json)

The **json** package provides a [config](https://github.com/warthog618/config) Decoder that unmarshals values from JSON formatted sources.

Example usage:

```go
import (
    "fmt"

    "github.com/warthog618/config"
    "github.com/warthog618/config/decoder/json"
    "github.com/warthog618/config/loader/file"
)

func main() {
    f, _ := config.NewSource(file.New("config.json"), json.NewDecoder())
    c := config.NewConfig(f)
    s := c.MustGet("nested.string").String()
    fmt.Println("s:", s)
    // ....
}
```