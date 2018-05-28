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
// The app allows you to play with the input sources and view the results.
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

func main() {
	log.SetFlags(0)

	cfg := loadConfig()
	if v, _ := cfg.GetBool("unmarshal"); v {
		dumpConfigU(cfg)
	} else {
		dumpConfig(cfg)
	}
}

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

func loadConfig() *config.Config {
	cfg := config.New()

	// highest priority first - flags override environment
	shortFlags := map[byte]string{
		'b': "module2.bool",
		'c': "config-file",
		'u': "unmarshal",
	}
	reader, err := flag.New([]string(nil), shortFlags)
	if err != nil {
		panic(err)
	}
	cfg.AppendReader(reader)

	// environment next
	ereader, err := env.New("APP_")
	if err != nil {
		panic(err)
	}
	cfg.AppendReader(ereader)

	// config file may be specified via flag or env, so check for it
	// and if present add it with lower priority than flag and env.
	configFile, err := cfg.GetString("config.file")
	if err == nil {
		// explicitly specified config file - must be there
		jreader, err := json.NewFile(configFile)
		if err != nil {
			panic(err)
		}
		cfg.AppendReader(jreader)
	} else {
		// implicit and optional default config file
		jreader, err := json.NewFile("app.json")
		if err == nil {
			cfg.AppendReader(jreader)
		} else {
			if _, ok := err.(*os.PathError); !ok {
				panic(err)
			}
		}
	}
	// finally add the defaults which have lowest priority and are only
	// used if none of the other sources find a field.
	jreader, err := json.NewBytes(defaultConfig)
	if err != nil {
		panic(err)
	}
	cfg.AppendReader(jreader)
	return cfg
}

func dumpConfig(cfg *config.Config) {
	configFile, _ := cfg.GetString("config.file")
	log.Println("config.file", configFile)
	unmarshal, _ := cfg.GetBool("unmarshal")
	log.Println("unmarshal", unmarshal)
	modules := []string{"module1", "module2"}
	for _, module := range modules {
		mCfg, _ := cfg.GetConfig(module)
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
		cslice, _ := mCfg.GetSlice(v)
		log.Printf("%s.%s %v\n", module, v, cslice)
	}
}

type module struct {
	A  int    `config:"int"`
	B1 int    `config:"bool"`
	B2 bool   `config:"bool"`
	C  string `config:"string"`
	D  []int  `config:"slice"`
}

// Unmarshalling version of dumpConfig
func dumpConfigU(cfg *config.Config) {
	configFile, _ := cfg.GetString("config.file")
	log.Println("config.file", configFile)
	unmarshal, _ := cfg.GetBool("unmarshal")
	log.Println("unmarshal", unmarshal)
	modules := []string{"module1", "module2"}
	for _, mname := range modules {
		m := module{}
		if err := cfg.Unmarshal(mname, &m); err != nil {
			log.Printf("%s unmarshal error: %v", mname, err)
			continue
		}
		log.Printf("%s.int %v\n", mname, m.A)
		log.Printf("%s.bool %v\n", mname, m.B1)
		log.Printf("%s.string %v\n", mname, m.C)
		log.Printf("%s.bool %v\n", mname, m.B2)
		log.Printf("%s.slice %v\n", mname, m.D)
	}
}
