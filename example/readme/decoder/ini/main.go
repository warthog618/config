package main

import (
	"fmt"

	"github.com/warthog618/config"
	"github.com/warthog618/config/decoder/ini"
	"github.com/warthog618/config/loader/file"
)

func main() {
	f, _ := config.NewSource(file.New("config.ini"), ini.NewDecoder())
	c := config.NewConfig(f)
	s := c.MustGet("nested.string").String()
	fmt.Println("s:", s)
	// ....
}
