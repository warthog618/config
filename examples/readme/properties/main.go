package main

import (
	"fmt"

	"github.com/warthog618/config"
	"github.com/warthog618/config/properties"
)

func main() {
	f, _ := properties.New(properties.FromFile("config.properties"))
	c := config.NewConfig(f)
	s, _ := c.GetString("nested.string")
	fmt.Println("s:", s)
	// ....
}
