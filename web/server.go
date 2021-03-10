package web

import (
	"context"
	"net/http"

	"github.com/ivanmartinez/boocat/log"
	"github.com/ivanmartinez/boocat/server"
)

var (
	httpServer *http.Server
)

// Initialize initializes the HTTP server configuration without starting it. This is used in testing.
func Initialize(url string) {
	templates = make(map[string]*Template)
	staticFiles = make(map[string]*StaticFile)

	mux := http.NewServeMux()
	mux.HandleFunc("/", Handle)
	httpServer = &http.Server{
		Addr:    url,
		Handler: mux,
	}
}

// Start starts the initialized HTTP server
func Start() {
	// Start the HTTP server in a new goroutine
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error.Fatalf("couldn't start HTTP server: %v", err)
		}
	}()
}

// Shutdown shuts the HTTP Server down
func Shutdown(ctx context.Context) {
	// Shut the HTTP server down
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Error.Fatalf("server shutdown failed: %v", err)
	}
}

// Handle handles a HTTP request.
// The only reason why this function is public is to make it testable.
func Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" && r.Method != "POST" {
		http.Error(w, "", http.StatusBadRequest)
	}
	if template, found := GetTemplate(r.URL.Path); found {
		handleWithTemplate(w, r, template)
		return
	}
	if file, found := GetFile(r.URL.Path); found {
		err := file.Write(w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	http.NotFound(w, r)
}

func handleWithTemplate(w http.ResponseWriter, r *http.Request, template *Template) {
	formValues := submittedFormValues(r)
	var data interface{}
	if r.Method == "GET" {
		data = handleGet(r.Context(), template.formatName, formValues)
	} else {
		// POST
		data = handlePost(r.Context(), template.formatName, formValues)
	}
	err := template.Write(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleGet(ctx context.Context, formatName string, params map[string]string) interface{} {
	if _, found := params["id"]; found {
		return server.GetRecord(ctx, formatName, params)
	}
	if search, found := params["_search"]; found {
		return server.SearchRecords(ctx, formatName, search)
	}
	return server.ListRecords(ctx, formatName)
}

func handlePost(ctx context.Context, formatName string, params map[string]string) interface{} {
	if _, found := params["id"]; found {
		return server.UpdateRecord(ctx, formatName, params)
	}
	return server.AddRecord(ctx, formatName, params)
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
