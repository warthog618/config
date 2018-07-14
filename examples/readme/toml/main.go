package main

import (
	"fmt"

	"github.com/warthog618/config"
	"github.com/warthog618/config/toml"
)

func main() {
	f, _ := toml.New(toml.FromFile("config.toml"))
	c := config.NewConfig(f)
	s, _ := c.GetString("nested.string")
	fmt.Println("s:", s)
	// ....
}
