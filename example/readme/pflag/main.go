package main

import (
	"fmt"

	"github.com/warthog618/config"
	"github.com/warthog618/config/pflag"
)

func main() {
	c := config.New(pflag.New())
	cfgFile := c.MustGet("config.file").String()
	fmt.Println("config-file:", cfgFile)
	// ....
}
