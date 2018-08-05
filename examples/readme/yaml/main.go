package main

import (
	"fmt"

	"github.com/warthog618/config"
	"github.com/warthog618/config/yaml"
)

func main() {
	f, _ := yaml.New(yaml.FromFile("config.yaml"))
	c := config.NewConfig(f)
	s := c.Get("nested.string").String()
	fmt.Println("s:", s)
	// ....
}
