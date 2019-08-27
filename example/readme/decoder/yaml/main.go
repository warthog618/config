package main

import (
	"fmt"

	"github.com/warthog618/config"
	"github.com/warthog618/config/blob"
	"github.com/warthog618/config/blob/decoder/yaml"
	"github.com/warthog618/config/blob/loader/file"
)

func main() {
	c := config.NewConfig(blob.New(file.New("config.yaml"), yaml.NewDecoder()))
	s := c.MustGet("nested.string").String()
	fmt.Println("s:", s)
	// ....
}
