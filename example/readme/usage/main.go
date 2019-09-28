package main

import (
	"fmt"
	"time"

	"github.com/warthog618/config"
	"github.com/warthog618/config/blob"
	"github.com/warthog618/config/blob/decoder/json"
	"github.com/warthog618/config/blob/loader/file"
	"github.com/warthog618/config/dict"
	"github.com/warthog618/config/env"
	"github.com/warthog618/config/pflag"
)

func main() {
	defaultConfig := dict.New(dict.WithMap(map[string]interface{}{
		"name":          "myapp",
		"env.prefix":    "MYAPP_",
		"config.file":   "myapp.json",
		"sm.pin":        27,
		"sm.period":     "250ms",
		"sm.thresholds": []int8{23, 45, 64},
	}))
	// from flags and defaults...
	cfg := config.New(
		pflag.New(pflag.WithFlags([]pflag.Flag{{Short: 'c', Name: "config-file"}})),
		config.WithDefault(defaultConfig))

	// and from environment...
	prefix := cfg.MustGet("env.prefix").String()
	cfg.Append(env.New(env.WithEnvPrefix(prefix)))

	// and from config file...
	cf := cfg.MustGet("config.file").String()
	cfg.Append(blob.New(file.New(cf), json.NewDecoder(), blob.MustLoad()))

	// read a config field from the root config
	name := cfg.MustGet("name").String()

	// to pass nested config to a sub-module...
	smCfg := cfg.GetConfig("sm")
	pin := smCfg.MustGet("pin").Uint()
	period := smCfg.MustGet("period").Duration()
	thresholds := smCfg.MustGet("thresholds").IntSlice()

	fmt.Println(cf, name, pin, period, thresholds)

	// or using Unmarshal to populate a config struct...
	type SMConfig struct {
		Pin        uint
		Period     time.Duration
		Thresholds []int
	}
	sc := SMConfig{}
	cfg.Unmarshal("sm", &sc)
	fmt.Println(cf, name, sc)
}
