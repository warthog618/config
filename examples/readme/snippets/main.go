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
	m := config.NewMust(g)

	// arrays
	ports := m.GetUintSlice("ports")

	// alternatively....
	size := int(m.GetUint("ports[]"))
	for i := 0; i < size; i++ {
		// get each port sequentially...
		port := m.GetUint(fmt.Sprintf("ports[%d]", i))
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
	r.Append("somearray\\[\\d+\\](.*)", "somearray[0]$1")
	c := config.NewConfig(config.Decorate(g, config.WithRegexAlias(r)))

	c.Get("")
}

func newConfig() error {
	g, _ := pflag.New()

	c := config.NewConfig(g)
	v, err := c.GetInt("pin")
	ports, err := c.GetUintSlice("ports")

	ports[0] = uint64(v)
	return err
}

func must() {
	g, _ := pflag.New()
	m := config.NewMust(g, config.WithPanic())
	v := m.GetInt("pin")
	ports := m.GetUintSlice("ports")

	ports[0] = uint64(v)
}
