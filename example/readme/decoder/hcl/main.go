package main

import (
	"fmt"

	"github.com/warthog618/config"
	"github.com/warthog618/config/blob"
	"github.com/warthog618/config/blob/decoder/hcl"
	"github.com/warthog618/config/blob/loader/file"
)

func main() {
	f, _ := blob.New(file.New("config.hcl"), hcl.NewDecoder())
	c := config.NewConfig(f)
	s := c.MustGet("nested[0].string").String()
	fmt.Println("s:", s)
	// ....
}
