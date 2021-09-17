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

	"github.com/ivanmartinez/boocat/boocat"
	"github.com/ivanmartinez/boocat/boocat/sqlite"
	"github.com/ivanmartinez/boocat/webserver"
)

func main() {
	// Parse flags
	url := flag.String("url", "localhost:80", "This boocat's base URL")
	dbDataSource := flag.String("dbds", "boocat.sqlite", "Database source")
	flag.Parse()

	// Create channel for listening to OS signals and connect OS interrupts to
	// the channel
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		oscall := <-c
		webserver.Info.Printf("received signal %v", oscall)
		cancel()
	}()

	// Open database
	db, err := sqlite.Open(ctx, dbDataSource)
	if err != nil {
		webserver.Error.Fatal(err)
	}
	// Set formats
	var bc boocat.Boocat
	bc.SetFormat("author", boocat.Format{
		Name: "author",
		Fields: map[string]boocat.Validate{
			"name":      regExpValidator("^([A-Z][a-z]*)([ |-][A-Z][a-z]*)*$"),
			"birthdate": validateYear,
			"biography": nil,
		},
		Searchable: map[string]struct{}{"name": {}, "biography": {}},
	})
	bc.SetFormat("book", boocat.Format{
		Name: "book",
		Fields: map[string]boocat.Validate{
			"name":     regExpValidator("^([A-Z][a-z]*)([ |-][A-Z][a-z]*)*$"),
			"year":     validateYear,
			"author":   db.ReferenceValidator("author"),
			"synopsis": nil,
		},
		Searchable: map[string]struct{}{"name": {}, "synopsis": {}},
	})

	// Set database to use
	bc.SetDatabase(db)
	ws := webserver.Initialize(*url, &bc)
	loadWebFiles(&ws)
	ws.Start()

	// Wait for ctx to be cancelled
	<-ctx.Done()

	// New context with timeout to shut the HTTP boocat down
	ctxShutDown, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	// Shut services down
	ws.Shutdown(ctxShutDown)
	if err := db.Close(); err != nil {
		webserver.Error.Print(err)
	}
}

// reqExpValidator returns a validator that uses the regular expression passed as argument
func regExpValidator(regExpString string) boocat.Validate {
	regExp, err := regexp.Compile(regExpString)
	if err != nil {
		webserver.Error.Fatal(err)
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

func loadWebFiles(ws *webserver.Webserver) {
	ws.LoadStaticFile("bcweb", "/index.html")
	ws.LoadTemplate("bcweb", "/author.tmpl", "author")
	ws.LoadTemplate("bcweb", "/book.tmpl", "book")
	ws.LoadTemplate("bcweb", "/new/author.tmpl", "author")
	ws.LoadTemplate("bcweb", "/new/book.tmpl", "book")
	ws.LoadTemplate("bcweb", "/edit/author.tmpl", "author")
	ws.LoadTemplate("bcweb", "/edit/book.tmpl", "book")
	ws.LoadTemplate("bcweb", "/list/author.tmpl", "author")
	ws.LoadTemplate("bcweb", "/list/book.tmpl", "book")
	ws.LoadTemplate("bcweb", "/search/author.tmpl", "author")
	ws.LoadTemplate("bcweb", "/search/book.tmpl", "book")
}
