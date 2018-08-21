# properties

[![GoDoc](https://godoc.org/github.com/warthog618/config/decoder/properties/sar?status.svg)](https://godoc.org/github.com/warthog618/config/decoder/properties)

The **properties** package provides a [config](https://github.com/warthog618/config) Decoder that unmarshals values from Properties formatted sources.

Example usage:

```go
import (
    "fmt"

    "github.com/warthog618/config"
    "github.com/warthog618/config/decoder/properties"
    "github.com/warthog618/config/loader/file"
)

func main() {
    f, _ := config.NewSource(file.New("config.properties"), properties.NewDecoder())
    c := config.NewConfig(f)
    s := c.MustGet("nested.string").String()
    fmt.Println("s:", s)
    // ....
}

```

The following option can be applied to properties.NewDecoder:

The [WithListSeparator](https://godoc.org/github.com/warthog618/config/decoder/properties#WithListSeparator) option provides a string used to split list values into elements.  The default list separator is ",".