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
	"github.com/warthog618/config/json"
	"github.com/warthog618/config/pflag"
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

// Config defines the config interface used by this module.
type Config interface {
	GetBool(key string) (bool, error)
	GetInt(key string) (int64, error)
	GetMust(node string, options ...config.MustOption) *config.Must
	GetString(key string) (string, error)
	Unmarshal(node string, obj interface{}) (rerr error)
}

func loadConfig() Config {
	def, err := json.New(json.FromBytes(defaultConfig))
	if err != nil {
		panic(err)
	}
	shortFlags := map[byte]string{
		'b': "module2.bool",
		'c': "config-file",
		'u': "unmarshal",
	}
	fget, err := pflag.New(pflag.WithShortFlags(shortFlags))
	if err != nil {
		panic(err)
	}
	// environment next
	eget, err := env.New(env.WithEnvPrefix("APP_"))
	if err != nil {
		panic(err)
	}
	// highest priority sources first - flags override environment
	sources := config.NewStack(fget, eget)
	cfg := config.NewConfig(config.Decorate(sources, config.WithDefault(def)))

	// config file may be specified via flag or env, so check for it
	// and if present add it with lower priority than flag and env.
	configFile, err := cfg.GetString("config.file")
	if err == nil {
		// explicitly specified config file - must be there
		jget, err := json.New(json.FromFile(configFile))
		if err != nil {
			panic(err)
		}
		sources.Append(jget)
	} else {
		// implicit and optional default config file
		jget, err := json.New(json.FromFile("app.json"))
		if err == nil {
			sources.Append(jget)
		} else {
			if _, ok := err.(*os.PathError); !ok {
				panic(err)
			}
		}
	}
	return cfg
}

func dumpConfig(cfg Config) {
	configFile, _ := cfg.GetString("config.file")
	log.Println("config.file", configFile)
	unmarshal, _ := cfg.GetBool("unmarshal")
	log.Println("unmarshal", unmarshal)
	modules := []string{"module1", "module2"}
	for _, module := range modules {
		mCfg := cfg.GetMust(module)
		ints := []string{
			"int",
			"bool",
		}
		for _, v := range ints {
			cint := mCfg.GetInt(v)
			log.Printf("%s.%s %v\n", module, v, cint)
		}
		v := "string"
		cstr := mCfg.GetString(v)
		log.Printf("%s.%s %v\n", module, v, cstr)
		v = "bool"
		cbool := mCfg.GetBool(v)
		log.Printf("%s.%s %v\n", module, v, cbool)
		v = "slice"
		cslice := mCfg.GetSlice(v)
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
func dumpConfigU(cfg Config) {
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
