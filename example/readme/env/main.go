package main

import (
	"fmt"

	"github.com/warthog618/config"
	"github.com/warthog618/config/env"
)

func main() {
	e, _ := env.New()
	c := config.NewConfig(e)
	cfgFile := c.MustGet("config.file").String()
	fmt.Println("config-file:", cfgFile)
	// ....
}
