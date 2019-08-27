# json

[![GoDoc](https://godoc.org/github.com/warthog618/config/blob/decoder/json/sar?status.svg)](https://godoc.org/github.com/warthog618/config/blob/decoder/json)

The **json** package provides a [config](https://github.com/warthog618/config) Decoder that unmarshals values from JSON formatted sources.

Example usage:

```go
import (
    "fmt"

    "github.com/warthog618/config"
    "github.com/warthog618/config/blob/decoder/json"
    "github.com/warthog618/config/blob/loader/file"
)

func main() {
    c := config.NewConfig(file.New("config.json", json.NewDecoder()))
    s := c.MustGet("nested.string").String()
    fmt.Println("s:", s)
    // ....
}
```
