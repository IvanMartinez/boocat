package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/ivanmartinez/boocat"
	"github.com/ivanmartinez/boocat/database"
	"github.com/ivanmartinez/boocat/formats"
	"github.com/ivanmartinez/boocat/webfiles"
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

	// Start services
	db := database.Connect(ctx, dbURI, []string{"author", "book"})
	formats.Initialize(db)
	server := startServer(ctx, db, *url)

	// Wait for ctx to be cancelled
	<-ctx.Done()

	// New context with timeout to shut the HTTP server down
	ctxShutDown, cancel := context.WithTimeout(context.Background(),
		5*time.Second)

	// Shut services down
	shutdownServer(ctxShutDown, server)
	db.Disconnect(ctxShutDown)
}

// startServer registers handlers for URL routes and starts the HTTP server
func startServer(ctx context.Context, db database.DB, url string) *http.Server {
	// @TODO: Find the actual URL, it could be using https
	boocat.HTTPURL = "http://" + url
	webfiles.Load()
	boocat.DB = db

	mux := http.NewServeMux()
	mux.HandleFunc("/", requestHandler)
	server := &http.Server{
		Addr:    url,
		Handler: mux,
	}

	// Start the HTTP server in a new goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {

			log.Fatalf("couldn't start HTTP server: %v", err)
		}
	}()

	return server
}

func shutdownServer(ctx context.Context, server *http.Server) {
	// Shut the HTTP server down
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("server shutdown failed: %v", err)
	}
}

/*
	Changing to paths as follows:
	- Get form to search records:       GET /{format}
	- Show search result:               POST /{format}?search_criteria
	- Get form to create new record:    GET /edit/{format}
	- Get form to edit existing record: GET /edit/{format}?id=...
	- Create new record:                POST .../{format}?...	// DIFFERENCE WITH SEARCH? id EMPTY?
	- Update existing record:           POST .../{format}?id=...
	- Show record:                      GET /{format}?id=...
*/

func requestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" && r.Method != "POST" {
		http.Error(w, "", http.StatusBadRequest)
	}

	// If there is a template for the path
	if template, found := webfiles.GetTemplate(r.URL.Path); found {
		handleTemplate(w, r, template)
		return
	}

	// If there is a static file for the path
	if file, found := webfiles.GetFile(r.URL.Path); found {
		err := file.Write(w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// If no file was found
	http.NotFound(w, r)

}

func handleTemplate(w http.ResponseWriter, r *http.Request, template *webfiles.Template) {
	path := r.URL.Path
	ext := filepath.Ext(path)
	base := filepath.Base(path)
	name := base[:len(base)-len(ext)]

	if _, found := formats.Get(name); !found {
		log.Printf("template file \"%v\" must have the name of a format", path)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	formValues := submittedFormValues(r)
	var data interface{}
	if r.Method == "GET" {
		data = boocat.Get(r.Context(), name, formValues)
	} else {
		// POST
		data = boocat.Update(r.Context(), name, formValues)
	}
	err := template.Write(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// submittedFormValues returns a map with the values of the query parameters as well as the submitted form fields.
// In case of conflict the form value prevails.
func submittedFormValues(r *http.Request) map[string]string {
	values := make(map[string]string)
	// Read values from the query parameters
	query := r.URL.Query()
	for param := range query {
		values[param] = query.Get(param)
	}
	// Read values from the posted form
	r.ParseForm()
	for field := range r.PostForm {
		values[field] = r.PostForm.Get(field)
	}
	return values
}
