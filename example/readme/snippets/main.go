// Copyright © 2018 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Example snippets used in the config README.

package main

import (
	"fmt"

	"github.com/warthog618/config"
	"github.com/warthog618/config/blob"
	"github.com/warthog618/config/blob/decoder/json"
	"github.com/warthog618/config/blob/loader/file"
	"github.com/warthog618/config/pflag"
)

func main() {
}

func arrays() {
	c := config.New(pflag.New())

	// arrays
	ports := c.MustGet("ports").UintSlice()

	// alternatively....
	size := int(c.MustGet("ports[]").Int())
	for i := 0; i < size; i++ {
		// get each port sequentially...
		ports[i] = c.MustGet(fmt.Sprintf("ports[%d]", i)).Uint()
	}
}

func alias() {
	var newKey, oldKey string

	a := config.NewAlias()
	c := config.New(config.Decorate(pflag.New(), config.WithAlias(a)))
	a.Append(newKey, oldKey)

	c.Get("")
}

func regexalias() {
	r := config.NewRegexAlias()
	r.Append(`somearray\[\d+\](.*)`, "somearray[0]$1")
	c := config.New(config.Decorate(pflag.New(), config.WithRegexAlias(r)))

	c.Get("")
}

func newConfig() {
	c := config.New(pflag.New())
	pin := c.MustGet("pin").Int()
	ports := c.MustGet("ports").UintSlice()

	ports[0] = uint64(pin)
}

func must() {
	c := config.New(pflag.New())
	pin := c.MustGet("pin").Int()

	if pin == 0 {
	}
}

func blobDef() {
	cfgFile := blob.New(file.New("config.json"), json.NewDecoder())
	c := config.New(cfgFile)

	c.Close()
}

func blobStack() {
	sources := config.NewStack()
	c := config.New(sources)
	cfgFile := blob.New(file.New("config.json"), json.NewDecoder())
	sources.Append(cfgFile)

	c.Close()
}
