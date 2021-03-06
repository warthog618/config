package main

import (
	"fmt"

	"github.com/warthog618/config"
	"github.com/warthog618/config/dict"
)

func main() {
	d := dict.New()
	d.Set("config.file", "config.json")
	c := config.New(d)
	cfgFile := c.MustGet("config.file").String()
	fmt.Println("config-file:", cfgFile)
	// ....
}
