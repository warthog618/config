package main

import (
	"fmt"

	"github.com/warthog618/config"
	"github.com/warthog618/config/blob"
	"github.com/warthog618/config/blob/decoder/properties"
	"github.com/warthog618/config/blob/loader/file"
)

func main() {
	f := file.New("config.properties")
	b := blob.New(f, properties.NewDecoder())
	c := config.NewConfig(b)
	s := c.MustGet("nested.string").String()
	fmt.Println("s:", s)
	// ....
}
