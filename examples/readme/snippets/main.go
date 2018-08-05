// Example snippets used in the config README.

package main

import (
	"fmt"

	"github.com/warthog618/config"
	"github.com/warthog618/config/pflag"
)

func main() {
}

func arrays() {
	g, _ := pflag.New()
	c := config.NewConfig(g)

	// arrays
	ports := c.Get("ports").UintSlice()

	// alternatively....
	size := int(c.Get("ports[]").Uint())
	for i := 0; i < size; i++ {
		// get each port sequentially...
		port := c.Get(fmt.Sprintf("ports[%d]", i)).Uint()
		ports[i] = port
	}
}

func alias() {
	g, _ := pflag.New()
	var newKey, oldKey string

	a := config.NewAlias()
	c := config.NewConfig(config.Decorate(g, config.WithAlias(a)))
	a.Append(newKey, oldKey)

	c.Get("")
}

func regexalias() {
	g, _ := pflag.New()

	r := config.NewRegexAlias()
	r.Append(`somearray\[\d+\](.*)`, "somearray[0]$1")
	c := config.NewConfig(config.Decorate(g, config.WithRegexAlias(r)))

	c.Get("")
}

func newConfig() {
	g, _ := pflag.New()

	c := config.NewConfig(g)
	v := c.Get("pin").Int()
	ports := c.Get("ports").UintSlice()

	ports[0] = uint64(v)
}

func must() {
	g, _ := pflag.New()
	m := config.NewConfig(g)
	v := m.MustGet("pin").Int()
	ports := m.Get("ports").UintSlice()

	ports[0] = uint64(v)
}
