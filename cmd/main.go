package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"

	"github.com/ivanmartinez/boocat"
	"github.com/ivanmartinez/boocat/database"
	"github.com/ivanmartinez/boocat/templates"
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
		log.Printf("received signal %v", oscall)
		cancel()
	}()

	// Open the database
	db := database.Connect(ctx, dbURI, []string{"author", "book"})
	defer db.Disconnect(ctx)

	// Start the HTTP server
	startHTTPServer(ctx, db, *url)
}

// startHTTPServer registers handlers for URL routes and starts the HTTP server
func startHTTPServer(ctx context.Context, db database.DB, url string) {
	// @TODO: Find the actual URL, it could be using https
	boocat.HTTPURL = "http://" + url
	templates.LoadAll()

	// Gorilla mux router allows us to use patterns in paths
	router := mux.NewRouter()
	// Register handle functions
	router.HandleFunc("/{pFormat}/edit",
		makeHandler(boocat.EditNew, db))
	router.HandleFunc("/{pFormat}/{pRecord}/edit",
		makeHandler(boocat.EditExisting, db))
	router.HandleFunc("/{pFormat}/save",
		makeHandler(boocat.SaveNew, db))
	router.HandleFunc("/{pFormat}/{pRecord}/save",
		makeHandler(boocat.SaveExisting, db))
	router.HandleFunc("/{pFormat}/{pRecord}",
		makeHandler(boocat.View, db))
	router.HandleFunc("/{pFormat}",
		makeHandler(boocat.List, db))

	// Start the HTTP server in a new goroutine
	srv := &http.Server{
		Addr:    url,
		Handler: router,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil &&
			err != http.ErrServerClosed {

			log.Fatalf("couldn't start HTTP server: %v", err)
		}
	}()

	// Wait for ctx to be cancelled
	<-ctx.Done()

	// New context with timeout to shut the HTTP server down
	ctxShutDown, cancel := context.WithTimeout(context.Background(),
		5*time.Second)
	defer func() {
		cancel()
	}()
	// Shut the HTTP server down
	if err := srv.Shutdown(ctxShutDown); err != nil {
		log.Fatalf("server shutdown failed: %v", err)
	}
}

// makeHandler returns a handler function created from the function passed as
// parameter. This reduces boilerplate since common handler operations are
// implemented here.
func makeHandler(tplHandler func(context.Context, database.DB, string, string,
	map[string]string) (string, interface{}), db database.DB) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		// Read the path variables and submitted form values
		vars := mux.Vars(r)
		pFormat := vars["pFormat"]
		pRecord := vars["pRecord"]
		submittedValues := submittedFormValues(r)

		// Perform the specific operation for the route
		tplName, tplData := tplHandler(r.Context(), db, pFormat, pRecord,
			submittedValues)

		// Get the template to generate the output
		tpl, found := templates.Get(tplName)
		if found {
			// Generate the output with the template and the result of the
			// operation
			err := tpl.Execute(w, tplData)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		} else {
			log.Printf("couldn't find template \"%v\"", tplName)
			http.NotFound(w, r)
		}
	}
}

// submittedFormValues returns a map with the values of the submitted form
func submittedFormValues(r *http.Request) map[string]string {
	values := make(map[string]string)

	r.ParseForm()
	for field := range r.PostForm {
		values[field] = r.PostForm.Get(field)
	}

	return values
}
