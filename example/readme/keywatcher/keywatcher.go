// Copyright © 2018 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"log"
	"time"

	"github.com/warthog618/config"
	"github.com/warthog618/config/blob"
	"github.com/warthog618/config/blob/decoder/json"
	"github.com/warthog618/config/blob/loader/file"
)

func main() {
	l, _ := file.New("config.json", file.WithWatcher())
	g, _ := blob.New(l, json.NewDecoder())
	c := config.NewConfig(g)

	done := make(chan struct{})
	defer close(done)
	// watcher goroutine
	go func() {
		w := c.NewKeyWatcher("somevariable")
		for {
			v, err := w.Watch(done)
			if err != nil {
				log.Println("watch error:", err)
				return
			}
			log.Println("got update:", v.Int())
		}
	}()
	// main thread
	time.Sleep(time.Minute)
	log.Println("finished.")
}
