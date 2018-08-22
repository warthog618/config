# ini

[![GoDoc](https://godoc.org/github.com/warthog618/config/decoder/ini/sar?status.svg)](https://godoc.org/github.com/warthog618/config/decoder/ini)

The **ini** package provides a [config](https://github.com/warthog618/config) Decoder that unmarshals values from INI formatted sources.

Example usage:

```go
import (
    "fmt"

    "github.com/warthog618/config"
    "github.com/warthog618/config/decoder/ini"
    "github.com/warthog618/config/loader/file"
)

func main() {
    f, _ := config.NewSource(file.New("config.ini"), ini.NewDecoder())
    c := config.NewConfig(f)
    s := c.MustGet("nested.string").String()
    fmt.Println("s:", s)
    // ....
}
```

The following option can be applied to ini.NewDecoder:

The [WithListSeparator](https://godoc.org/github.com/warthog618/config/decoder/ini#WithListSeparator) option provides a string used to split list values into elements.  The default list separator is ",".