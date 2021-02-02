package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"time"

	"github.com/ivanmartinez/boocat/database"
	"github.com/ivanmartinez/boocat/formats"
	"github.com/ivanmartinez/boocat/log"
	"github.com/ivanmartinez/boocat/server"
)

func main() {
	// Parse flags
	url := flag.String("url", "localhost:80", "This server's base URL")
	dbURI := flag.String("dburi", "mongodb://127.0.0.1:27017", "Database URI")
	flag.Parse()

	// Create channel for listening to OS signals and connect OS interrupts to
	// the channel
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		oscall := <-c
		log.Info.Printf("received signal %v", oscall)
		cancel()
	}()

	// Start services
	formats.InitializeFields()
	db := database.Initialize(ctx, dbURI, formats.Formats)
	server.Initialize(*url, "bcweb", db)
	server.Start()

	// Wait for ctx to be cancelled
	<-ctx.Done()

	// New context with timeout to shut the HTTP server down
	ctxShutDown, cancel := context.WithTimeout(context.Background(),
		5*time.Second)

	// Shut services down
	server.ShutdownServer(ctxShutDown)
	db.Disconnect(ctxShutDown)
}
