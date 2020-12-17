package server

import (
	"context"
	"log"
	"net/http"
	"path/filepath"

	"github.com/ivanmartinez/boocat/database"
	"github.com/ivanmartinez/boocat/formats"
	"github.com/ivanmartinez/boocat/webfiles"
)

var (
	db     database.DB
	server *http.Server
)

// Initialize initializes the server configuration without starting it. This is used in testing.
func Initialize(ctx context.Context, url, path string, database database.DB) {
	webfiles.Load(path)
	db = database
	mux := http.NewServeMux()
	mux.HandleFunc("/", Handle)
	server = &http.Server{
		Addr:    url,
		Handler: mux,
	}
}

// Start starts the initialized server
func Start() {
	// Start the HTTP server in a new goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("couldn't start HTTP server: %v", err)
		}
	}()
}

func ShutdownServer(ctx context.Context) {
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

// Handle handles a HTTP request.
// The only reason this function is public is to make it testable.
func Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" && r.Method != "POST" {
		http.Error(w, "", http.StatusBadRequest)
	}
	if template, found := webfiles.GetTemplate(r.URL.Path); found {
		handleWithTemplate(w, r, template)
		return
	}
	if file, found := webfiles.GetFile(r.URL.Path); found {
		err := file.Write(w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	http.NotFound(w, r)
}

func handleWithTemplate(w http.ResponseWriter, r *http.Request, template *webfiles.Template) {
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
		data = handleGet(r.Context(), name, formValues)
	} else {
		// POST
		data = handlePost(r.Context(), name, formValues)
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

func handleGet(ctx context.Context, formatName string, params map[string]string) interface{} {
	if id, found := params["id"]; found {
		return getRecord(ctx, formatName, id)
	}
	return list(ctx, formatName)
}

func handlePost(ctx context.Context, formatName string, params map[string]string) interface{} {
	if _, found := params["id"]; found {
		return updateRecord(ctx, formatName, params)
	}
	return newRecord(ctx, formatName, params)
}

// newRecord adds a record of a format (author, book...)
func newRecord(ctx context.Context, formatName string, record map[string]string) interface{} {
	format, found := formats.Get(formatName)
	if !found {
		log.Printf("couldn't find format \"%v\"", formatName)
		return nil
	}
	failed := formats.Validate(ctx, format, record)
	if len(failed) == 0 {
		id, err := db.AddRecord(ctx, formatName, record)
		if err != nil {
			log.Printf("error adding record to database: %v\n", err)
		} else {
			record["id"] = id
		}
	}
	return record
}

// updateRecord updates a record of a format (author, book...)
func updateRecord(ctx context.Context, formatName string, record map[string]string) interface{} {
	format, found := formats.Get(formatName)
	if !found {
		log.Printf("couldn't find format \"%v\"", formatName)
		return nil
	}
	failed := formats.Validate(ctx, format, record)
	if len(failed) == 0 {
		// If record doesn't have all the fields defined in the format, get the missing fields from the database
		if formats.IncompleteRecord(format, record) {
			if dbRecord, err := db.GetRecord(ctx, formatName, record["id"]); err == nil {
				record = formats.Merge(format, record, dbRecord)
			} else {
				log.Printf("error getting database record: %v\n", err)
			}
		}
		if err := db.UpdateRecord(ctx, formatName, record); err != nil {
			log.Printf("error updating record in database: %v\n", err)
		}
	}
	return record
}

// getRecord returns a record of a format (author, book...)
func getRecord(ctx context.Context, formatName, id string) map[string]string {
	record, err := db.GetRecord(ctx, formatName, id)
	if err != nil {
		log.Printf("error getting database record: %v\n", err)
		//_, tplData := EditNew(ctx, format, id, nil)
		return nil
	}

	return record
}

// list returns a slice of all records of a format (authors, books...)
func list(ctx context.Context, format string) []map[string]string {
	records, err := db.GetAllRecords(ctx, format)
	if err != nil {
		log.Printf("error getting records from database: %v\n", err)
		return nil
	}
	return records
}
