// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"

	"github.com/ivanmartinez/wikiforms"
)

func main() {
	// Parse flags
	port := flag.String("p", "9090", "Listening port")
	flag.Parse()

	// Create channel for listening to OS signals and connect OS interrupts to
	// the channel
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		oscall := <-c
		log.Printf("received signal %v", oscall)
		cancel()
	}()

	// Start this server
	wikiforms.StartServer(ctx, *port)
}
