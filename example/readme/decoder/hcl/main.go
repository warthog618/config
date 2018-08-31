package main

import (
	"fmt"

	"github.com/warthog618/config"
	"github.com/warthog618/config/blob"
	"github.com/warthog618/config/blob/decoder/hcl"
	"github.com/warthog618/config/blob/loader/file"
)

func main() {
	f, _ := file.New("config.hcl")
	b, _ := blob.New(f, hcl.NewDecoder())
	c := config.NewConfig(b)
	s := c.MustGet("nested[0].string").String()
	fmt.Println("s:", s)
	// ....
}
