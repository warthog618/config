# hcl

[![GoDoc](https://godoc.org/github.com/warthog618/config/blob/decoder/hcl/sar?status.svg)](https://godoc.org/github.com/warthog618/config/blob/decoder/hcl)

The **hcl** package provides a [config](https://github.com/warthog618/config) Decoder that unmarshals values from HCL formatted sources.

Example usage:

```go
import (
    "fmt"

    "github.com/warthog618/config"
    "github.com/warthog618/config/blob"
    "github.com/warthog618/config/blob/decoder/hcl"
    "github.com/warthog618/config/blob/loader/file"
)

func main() {
    c := config.New(blob.New(file.New("config.hcl"), hcl.NewDecoder(),blob.MustLoad()))
    s := c.MustGet("nested[0].string").String()
    fmt.Println("s:", s)
    // ....
}
```

Note that the HCL parser always converts objects into arrays of objects, even if
there is only one instance of the object, hence the need for the indexing of
nested, **nested[0]**, in the example above.  There is no way to define a single
nested object in HCL.
