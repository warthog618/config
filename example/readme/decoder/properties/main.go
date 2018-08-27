package main

import (
	"fmt"

	"github.com/warthog618/config"
	"github.com/warthog618/config/blob"
	"github.com/warthog618/config/blob/decoder/properties"
	"github.com/warthog618/config/blob/loader/file"
)

func main() {
	f, _ := blob.New(file.New("config.properties"), properties.NewDecoder())
	c := config.NewConfig(f)
	s := c.MustGet("nested.string").String()
	fmt.Println("s:", s)
	// ....
}
