package main

import (
	"fmt"

	"github.com/warthog618/config"
	"github.com/warthog618/config/properties"
)

func main() {
	f, _ := properties.New(properties.FromFile("config.properties"))
	c := config.NewConfig(f)
	s := c.MustGet("nested.string").String()
	fmt.Println("s:", s)
	// ....
}
