package main

import (
	"flag"
	"fmt"

	"github.com/warthog618/config"
	cfgflag "github.com/warthog618/config/flag"
)

func main() {
	flag.String("config-file", "config.json", "config file name")
	flag.Parse()
	f, _ := cfgflag.New()
	c := config.NewConfig(f)
	cfgFile, _ := c.GetString("config.file")
	fmt.Println("config-file:", cfgFile)
	// ....
}
