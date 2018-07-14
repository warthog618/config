# properties

[![GoDoc](https://godoc.org/github.com/warthog618/config/properties/sar?status.svg)](https://godoc.org/github.com/warthog618/config/properties)

The **properties** package provides a [config](https://github.com/warthog618/config) Getter that retrieves values from Properties files or other Properties formatted sources.

Example usage:

```go
import (
    "fmt"

    "github.com/warthog618/config"
    "github.com/warthog618/config/properties"
)

func main() {
    f, _ := properties.New(properties.FromFile("config.properties"))
    c := config.NewConfig(f)
    s, _ := c.GetString("nested.string")
    fmt.Println("s:", s)
    // ....
}
```

A number of options can be applied to properties.New:

The [FromBytes](https://godoc.org/github.com/warthog618/config/properties#FromBytes) option reads the Properties from a byte array.

The [FromFile](https://godoc.org/github.com/warthog618/config/properties#FromFile) option reads the Properties from a file.

The [FromReader](https://godoc.org/github.com/warthog618/config/properties#FromReader) option reads the Properties from an io.Reader.

The [WithListSeparator](https://godoc.org/github.com/warthog618/config/properties#WithListSeparator) option provides a string used to split list values into elements.  The default list separator is ",".