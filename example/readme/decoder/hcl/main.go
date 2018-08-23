package main

import (
	"fmt"

	"github.com/warthog618/config"
	"github.com/warthog618/config/decoder/hcl"
	"github.com/warthog618/config/loader/file"
)

func main() {
	f, _ := config.NewSource(file.New("config.hcl"), hcl.NewDecoder())
	c := config.NewConfig(f)
	s := c.MustGet("nested[0].string").String()
	fmt.Println("s:", s)
	// ....
}
