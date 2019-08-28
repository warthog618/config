package main

import (
	"fmt"

	"github.com/warthog618/config"
	"github.com/warthog618/config/blob"
	"github.com/warthog618/config/blob/decoder/toml"
	"github.com/warthog618/config/blob/loader/file"
)

func main() {
	c := config.New(blob.New(file.New("config.toml"), toml.NewDecoder()))
	s := c.MustGet("nested.string").String()
	fmt.Println("s:", s)
	// ....
}
