package main

import (
	"fmt"

	"github.com/warthog618/config"
	"github.com/warthog618/config/decoder/properties"
	"github.com/warthog618/config/loader/file"
)

func main() {
	f, _ := config.NewSource(file.New("config.properties"), properties.NewDecoder())
	c := config.NewConfig(f)
	s := c.MustGet("nested.string").String()
	fmt.Println("s:", s)
	// ....
}