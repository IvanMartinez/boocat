package boocat

import (
	"context"
	"log"
	"net/http"
	"path/filepath"

	"github.com/ivanmartinez/boocat/database"
	"github.com/ivanmartinez/boocat/formats"
	"github.com/ivanmartinez/boocat/webfiles"
)

// StartServer registers handlers for URL routes and starts the HTTP server
func StartServer(ctx context.Context, db database.DB, url string) *http.Server {
	// @TODO: Find the actual URL, it could be using https
	HTTPURL = "http://" + url
	webfiles.Load()
	formats.Initialize(db)

	mux := http.NewServeMux()
	mux.HandleFunc("/", requestHandler)
	server := &http.Server{
		Addr:    url,
		Handler: mux,
	}

	// Start the HTTP server in a new goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil &&
			err != http.ErrServerClosed {

			log.Fatalf("couldn't start HTTP server: %v", err)
		}
	}()

	return server
}

func ShutdownServer(ctx context.Context, server *http.Server) {
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
	- Get form to edit existing record: GET /edit/{format}?record
	- Create new record:                POST /{format}
	- Update existing record:           POST /{format}?record
	- Show record:                      GET /{format}?record
*/

func requestHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	ext := filepath.Ext(path)
	base := filepath.Base(path)
	name := base[:len(base)-len(ext)]

	// If there is a template for the path
	if template, found := webfiles.GetTemplate(path); found {
		if _, found := formats.Get(name); !found {
			log.Printf("template path \"%v\" should end with the name of a format", path)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
		err := template.Write(w, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// If there is a static file for the path
	if file, found := webfiles.GetFile(path); found {
		err := file.Write(w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// No file was found
	http.NotFound(w, r)

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
