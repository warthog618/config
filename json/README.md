# json

[![GoDoc](https://godoc.org/github.com/warthog618/config/json/sar?status.svg)](https://godoc.org/github.com/warthog618/config/json)

The **json** package provides a [config](https://github.com/warthog618/config) Getter that retrieves values from JSON files or other JSON formatted sources.

Example usage:

```go
import (
    "fmt"

    "github.com/warthog618/config"
    "github.com/warthog618/config/json"
)

func main() {
    f, _ := json.New(json.FromFile("config.json"))
    c := config.NewConfig(f)
    s, _ := c.GetString("nested.string")
    fmt.Println("s:", s)
    // ....
}
```

A number of options can be applied to json.New:

The [FromBytes](https://godoc.org/github.com/warthog618/config/json#FromBytes) option reads the JSON from a byte array.

The [FromFile](https://godoc.org/github.com/warthog618/config/json#FromFile) option reads the JSON from a file.

The [FromReader](https://godoc.org/github.com/warthog618/config/json#FromReader) option reads the JSON from an io.Reader.