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

	"github.com/ivanmartinez/boocat/log"
	"github.com/ivanmartinez/boocat/server"
	"github.com/ivanmartinez/boocat/server/database"
	"github.com/ivanmartinez/boocat/server/formats"
	"github.com/ivanmartinez/boocat/web"
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
	server.Initialize(db)
	web.Initialize(*url)
	loadWebFiles()
	web.Start()

	// Wait for ctx to be cancelled
	<-ctx.Done()

	// New context with timeout to shut the HTTP server down
	ctxShutDown, cancel := context.WithTimeout(context.Background(),
		5*time.Second)

	// Shut services down
	web.Shutdown(ctxShutDown)
	db.Disconnect(ctxShutDown)
}

// initializeFields initializes the formats and fields
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

// initializeValidators initializes the validators of the values of the fields
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
	return func(_ context.Context, value interface{}) string {
		stringValue := fmt.Sprintf("%v", value)
		if !regExp.MatchString(stringValue) {
			return fmt.Sprintf("does not match regular expression '%s'", regExpString)
		}
		return ""
	}
}

// validateYear returns a validator that validates a year
func validateYear(_ context.Context, value interface{}) string {
	stringValue := fmt.Sprintf("%v", value)
	year, err := strconv.Atoi(stringValue)
	if err != nil {
		return "not a valid year number"
	}
	if year < 0 {
		return "not a valid year number"
	}
	return ""
}

func loadWebFiles() {
	web.LoadStaticFile("bcweb", "/index.html")
	web.LoadTemplate("bcweb", "/author.tmpl", "author")
	web.LoadTemplate("bcweb", "/book.tmpl", "book")
	web.LoadTemplate("bcweb", "/new/author.tmpl", "author")
	web.LoadTemplate("bcweb", "/new/book.tmpl", "book")
	web.LoadTemplate("bcweb", "/edit/author.tmpl", "author")
	web.LoadTemplate("bcweb", "/edit/book.tmpl", "book")
	web.LoadTemplate("bcweb", "/list/author.tmpl", "author")
	web.LoadTemplate("bcweb", "/list/book.tmpl", "book")
	web.LoadTemplate("bcweb", "/search/author.tmpl", "author")
	web.LoadTemplate("bcweb", "/search/book.tmpl", "book")
}
