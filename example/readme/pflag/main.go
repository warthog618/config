package main

import (
	"fmt"

	"github.com/warthog618/config"
	"github.com/warthog618/config/pflag"
)

func main() {
	f := pflag.New()
	c := config.NewConfig(f)
	cfgFile := c.MustGet("config.file").String()
	fmt.Println("config-file:", cfgFile)
	// ....
}
