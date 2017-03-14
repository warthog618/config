// Copyright © 2017 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// A simple app using a variety of config sources.
//
// This example app demonstrates the overlaying of four config sources:
// - flag
// - env
// - json config file
// - default JSON []byte.
//
// The sources are overlayed in decreasing order of importance, as per the
// order above.
//
// The app can be build on any plaform and allows you to play with the input
// sources and view the results.
//
// This is only an example of one particular precedence ordering and with
// one set of sources, but is probably a reasonably common one.
// Other applications may define other sources or orderings as they see fit.
package main

import (
	"log"
	"os"

	"github.com/warthog618/config"
	"github.com/warthog618/config/env"
	"github.com/warthog618/config/flag"
	"github.com/warthog618/config/json"
)

var defaultConfig = []byte(`{
	"module1": {
		"int": 27,
    "string": "m1.s from default",
    "bool": true,
    "slice": [1,2,3,4]
	},
	"module2": {
    "int": 28,
    "string": "m2.s from default",
    "bool": false,
    "slice": [5,6,7,8]
	}
}`)

func main() {
	log.SetFlags(0)

	config := config.New()

	// highest priority first - flags override environment
	if reader, err := flag.New([]string(nil), map[byte]string{'c': "config-file"}); err != nil {
		panic(err)
	} else {
		config.AppendReader(reader)
	}

	// environment next
	if reader, err := env.New("APP_"); err != nil {
		panic(err)
	} else {
		config.AppendReader(reader)
	}

	// config file may be specified via flag or env, so check for it
	// and if present add it with lower priority than flag and env.
	if configFile, err := config.GetString("config.file"); err == nil {
		// explicitly specified config file - must be there
		if reader, err := json.NewFile(configFile); err != nil {
			panic(err)
		} else {
			config.AppendReader(reader)
		}
	} else {
		// implicit and optional default config file
		if reader, err := json.NewFile("app.json"); err == nil {
			config.AppendReader(reader)
		} else {
			if _, ok := err.(*os.PathError); !ok {
				panic(err)
			}
		}
	}
	// finally add the defaults which have lowest priority and are only
	// used if none of the other sources find a field.
	if reader, err := json.NewBytes(defaultConfig); err != nil {
		panic(err)
	} else {
		config.AppendReader(reader)
	}

	// Dump config
	configFile, _ := config.GetString("config.file")
	log.Println("config.file", configFile)
	modules := []string{"module1", "module2"}
	for _, module := range modules {
		mCfg, _ := config.GetConfig(module)
		ints := []string{
			"int",
			"bool",
		}
		for _, v := range ints {
			cint, _ := mCfg.GetInt(v)
			log.Printf("%s.%s %v\n", module, v, cint)
		}
		v := "string"
		cstr, _ := mCfg.GetString(v)
		log.Printf("%s.%s %v\n", module, v, cstr)
		v = "bool"
		cbool, _ := mCfg.GetBool(v)
		log.Printf("%s.%s %v\n", module, v, cbool)
		v = "slice"
		cslice, _ := config.GetSlice(v)
		log.Printf("%s.%s %v\n", module, v, cslice)
	}
}
