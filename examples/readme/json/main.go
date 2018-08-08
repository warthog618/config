package main

import (
	"fmt"

	"github.com/warthog618/config"
	"github.com/warthog618/config/json"
)

func main() {
	f, _ := json.New(json.FromFile("config.json"))
	c := config.NewConfig(f)
	s := c.MustGet("nested.string").String()
	fmt.Println("s:", s)
	// ....
}
