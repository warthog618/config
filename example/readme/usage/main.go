package main

import (
	"fmt"

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
		"sm.thresholds": "23,45,64",
	}))
	var g config.Getter
	g, _ = pflag.New(pflag.WithShortFlags(map[byte]string{'c': "config-file"}))
	sources := config.NewStack(g)
	cfg := config.NewConfig(
		config.Decorate(sources, config.WithDefault(defaultConfig)))
	prefix := cfg.MustGet("env.prefix").String()
	g, _ = env.New(env.WithEnvPrefix(prefix))
	sources.Append(g)
	cf := cfg.MustGet("config.file").String()
	f, _ := file.New(cf)
	g, _ = blob.New(f, json.NewDecoder())
	sources.Append(g)

	// read a config field from the root config
	name := cfg.MustGet("name").String()

	// to pass nested config to a sub-module...
	smCfg := cfg.GetConfig("sm")
	pin := smCfg.MustGet("pin").Uint()
	period := smCfg.MustGet("period").Duration()
	thresholds, _ := smCfg.Get("thresholds")

	fmt.Println(cf, name, pin, period, thresholds)
}
