# yaml

[![GoDoc](https://godoc.org/github.com/warthog618/config/yaml/sar?status.svg)](https://godoc.org/github.com/warthog618/config/yaml)

The **yaml** package provides a [config](https://github.com/warthog618/config) Getter that retrieves values from YAML files or other YAML formatted sources.

Example usage:

```go
import (
    "fmt"

    "github.com/warthog618/config"
    "github.com/warthog618/config/yaml"
)

func main() {
    f, _ := yaml.New(yaml.FromFile("config.yaml"))
    c := config.NewConfig(f)
    s, _ := c.GetString("nested.string")
    fmt.Println("s:", s)
    // ....
}
```

A number of options can be applied to yaml.New:

The [FromBytes](https://godoc.org/github.com/warthog618/config/yaml#FromBytes) option reads the YAML from a byte array.

The [FromFile](https://godoc.org/github.com/warthog618/config/yaml#FromFile) option reads the YAML from a file.

The [FromReader](https://godoc.org/github.com/warthog618/config/yaml#FromReader) option reads the YAML from an io.Reader.