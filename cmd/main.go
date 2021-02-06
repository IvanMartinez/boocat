package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"strconv"
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
	formats.Formats = make(map[string]formats.Format)
	initializeFields(formats.Formats)
	db := database.Initialize(ctx, dbURI, formats.Formats)
	initializeValidators(formats.Formats, db)
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

// initializeFields initializes the format fields
func initializeFields(bcFormats map[string]formats.Format) {
	bcFormats["author"] = formats.Format{
		Name: "author",
		Fields: map[string]formats.Validate{
			"name":      nil,
			"birthdate": nil,
			"biography": nil,
		},
		Searchable: map[string]struct{}{"name": {}, "biography": {}},
	}

	bcFormats["book"] = formats.Format{
		Name: "book",
		Fields: map[string]formats.Validate{
			"name":     nil,
			"year":     nil,
			"author":   nil,
			"synopsis": nil,
		},
		Searchable: map[string]struct{}{"name": {}, "synopsis": {}},
	}
}

func initializeValidators(bcFormats map[string]formats.Format, db database.DB) {
	bcFormats["author"].Fields["name"] = regExpValidator("^([A-Z][a-z]*)([ |-][A-Z][a-z]*)*$")
	bcFormats["author"].Fields["birthdate"] = validateYear

	bcFormats["book"].Fields["name"] = regExpValidator("^([A-Z][a-z]*)([ |-][A-Z][a-z]*)*$")
	bcFormats["book"].Fields["year"] = validateYear
	bcFormats["book"].Fields["author"] = db.ReferenceValidator("author")
}

// reqExpValidator returns a validator that uses the regular expression passed as argument
func regExpValidator(regExpString string) formats.Validate {
	regExp, err := regexp.Compile(regExpString)
	if err != nil {
		log.Error.Fatal(err)
	}
	return func(_ context.Context, value interface{}) bool {
		stringValue := fmt.Sprintf("%v", value)
		return regExp.MatchString(stringValue)
	}
}

func validateYear(_ context.Context, value interface{}) bool {
	stringValue := fmt.Sprintf("%v", value)
	year, err := strconv.Atoi(stringValue)
	if err != nil {
		return false
	}
	if year < 0 {
		return false
	}
	return true
}
