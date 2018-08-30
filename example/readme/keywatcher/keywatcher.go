// Copyright © 2018 Kent Gibson <warthog618@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"log"
	"time"

	"github.com/warthog618/config"
	"github.com/warthog618/config/blob"
	"github.com/warthog618/config/blob/decoder/json"
	"github.com/warthog618/config/blob/loader/file"
)

func main() {
	l, _ := file.NewWatched("config.json")
	g, _ := blob.NewWatched(l, json.NewDecoder())
	c := config.NewConfig(g)
	c.AddWatchedGetter(g)

	update := make(chan int64)
	w := c.NewKeyWatcher("somevariable")
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	// watcher goroutine
	go func() {
		for {
			v, err := w.Watch(ctx)
			if err != nil {
				close(update)
				break
			}
			update <- v.Int()
		}
	}()
	// main thread
	for {
		v, ok := <-update
		if !ok {
			break
		}
		log.Println("got update:", v)
	}
}
