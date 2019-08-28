# ini

[![GoDoc](https://godoc.org/github.com/warthog618/config/blob/decoder/ini/sar?status.svg)](https://godoc.org/github.com/warthog618/config/blob/decoder/ini)

The **ini** package provides a [config](https://github.com/warthog618/config) Decoder that unmarshals values from INI formatted sources.

Example usage:

```go
import (
    "fmt"

    "github.com/warthog618/config"
    "github.com/warthog618/config/blob"
    "github.com/warthog618/config/blob/decoder/ini"
    "github.com/warthog618/config/blob/loader/file"
)

func main() {
    c := config.New(blob.New(file.New("config.ini"), ini.NewDecoder()))
    s := c.MustGet("nested.string").String()
    fmt.Println("s:", s)
    // ....
}
```

The following option can be applied to ini.NewDecoder:

The
[WithListSeparator](https://godoc.org/github.com/warthog618/config/blob/decoder/ini#WithListSeparator)
option provides a string used to split list values into elements.  The default
list separator is ",".
