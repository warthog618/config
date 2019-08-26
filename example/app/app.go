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
	"time"

	"github.com/warthog618/config"
	"github.com/warthog618/config/blob"
	"github.com/warthog618/config/blob/decoder/json"
	"github.com/warthog618/config/blob/loader/bytes"
	"github.com/warthog618/config/blob/loader/file"
	"github.com/warthog618/config/env"
	"github.com/warthog618/config/pflag"
)

func main() {
	log.SetFlags(0)

	cfg := loadConfig()
	v, _ := cfg.Get("unmarshal")
	if v.Bool() {
		dumpConfigU(cfg)
	} else {
		dumpConfig(cfg)
	}
	w := cfg.NewWatcher()
	done := make(chan struct{})
	go func() {
		<-time.After(time.Minute)
		close(done)
	}()
	for {
		if err := w.Watch(done); err != nil {
			log.Println("watch error:", err)
			break
		}
		log.Println("updated to...")
		dumpConfig(cfg)
	}
	log.Println("finished.")
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
	jsondec := json.NewDecoder()
	def := blob.New(bytes.New(defaultConfig), jsondec, blob.MustLoad())
	shortFlags := map[byte]string{
		'b': "module2.bool",
		'c': "config-file",
		'u': "unmarshal",
	}
	// highest priority sources first - flags override environment
	sources := config.NewStack(
		pflag.New(pflag.WithShortFlags(shortFlags)),
		env.New(env.WithEnvPrefix("APP_")))
	cfg := config.NewConfig(sources, config.WithDefault(def))
	// config file may be specified via flag or env, so check for it
	// and if present add it with lower priority than flag and env.
	sources.Append(blob.NewConfigFile(
		cfg, "config.file", "app.json", jsondec, file.WithWatcher()))
	return cfg
}

func dumpConfig(cfg *config.Config) {
	configFile, _ := cfg.Get("config.file")
	log.Println("config.file", configFile.String())
	unmarshal, _ := cfg.Get("unmarshal")
	log.Println("unmarshal", unmarshal.Bool())
	modules := []string{"module1", "module2"}
	for _, module := range modules {
		mCfg := cfg.GetConfig(module)
		ints := []string{
			"int",
			"bool",
		}
		for _, v := range ints {
			cint, _ := mCfg.Get(v)
			log.Printf("%s.%s as int %v\n", module, v, cint.Int())
		}
		v := "string"
		cstr, _ := mCfg.Get(v)
		log.Printf("%s.%s %v\n", module, v, cstr.String())
		v = "bool"
		cbool, _ := mCfg.Get(v)
		log.Printf("%s.%s %v\n", module, v, cbool.Bool())
		v = "slice"
		cslice, _ := mCfg.Get(v)
		log.Printf("%s.%s %v\n", module, v, cslice.Slice())
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
	configFile, _ := cfg.Get("config.file")
	log.Println("config.file", configFile.String())
	unmarshal, _ := cfg.Get("unmarshal")
	log.Println("unmarshal", unmarshal.Bool())
	modules := []string{"module1", "module2"}
	for _, mname := range modules {
		m := module{}
		if err := cfg.Unmarshal(mname, &m); err != nil {
			log.Printf("%s unmarshal error: %v", mname, err)
			continue
		}
		log.Printf("%s.int %v\n", mname, m.A)
		log.Printf("%s.bool as int %v\n", mname, m.B1)
		log.Printf("%s.bool %v\n", mname, m.B2)
		log.Printf("%s.string %v\n", mname, m.C)
		log.Printf("%s.slice %v\n", mname, m.D)
	}
}
