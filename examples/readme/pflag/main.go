package main

import (
	"fmt"

	"github.com/warthog618/config"
	"github.com/warthog618/config/pflag"
)

func main() {
	f, _ := pflag.New()
	c := config.NewConfig(f)
	cfgFile, _ := c.GetString("config.file")
	fmt.Println("config-file:", cfgFile)
	// ....
}
