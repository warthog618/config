# cfgconv

[![GoDoc](https://godoc.org/github.com/warthog618/config/sar?status.svg)](https://godoc.org/github.com/warthog618/config/cfgconv)

The **cfgconv** package provides functions to convert from an *interface{}* to a particular type.

The **cfgconv** package is similar to the standard
[**strconv**](https://golang.org/pkg/strconv/) package, but converts from *interface{}* instead of *string*.

The conversions performed by **cfgconv** are as permissive as possible, given the data types involved, to allow for mapping from sources that may not directly support the requested type.
