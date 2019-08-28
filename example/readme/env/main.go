package main

import (
	"fmt"

	"github.com/warthog618/config"
	"github.com/warthog618/config/env"
)

func main() {
	c := config.New(env.New())
	cfgFile := c.MustGet("config.file").String()
	fmt.Println("config-file:", cfgFile)
	// ....
}
