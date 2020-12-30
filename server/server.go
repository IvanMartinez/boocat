package server

import (
	"context"
	"net/http"

	"github.com/ivanmartinez/boocat/database"
	"github.com/ivanmartinez/boocat/formats"
	"github.com/ivanmartinez/boocat/log"
	"github.com/ivanmartinez/boocat/validators"
	"github.com/ivanmartinez/boocat/webfiles"
)

var (
	db     database.DB
	server *http.Server
)

// Initialize initializes the server configuration without starting it. This is used in testing.
func Initialize(url, path string, database database.DB) {
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
			log.Error.Fatalf("couldn't start HTTP server: %v", err)
		}
	}()
}

func ShutdownServer(ctx context.Context) {
	// Shut the HTTP server down
	if err := server.Shutdown(ctx); err != nil {
		log.Error.Fatalf("server shutdown failed: %v", err)
	}
}

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
	formValues := submittedFormValues(r)
	var data interface{}
	if r.Method == "GET" {
		data = handleGet(r.Context(), template.FormatName, formValues)
	} else {
		// POST
		data = handlePost(r.Context(), template.FormatName, template.Format, formValues)
	}
	err := template.Write(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleGet(ctx context.Context, formatName string, params map[string]string) interface{} {
	if id, found := params["id"]; found {
		return getRecord(ctx, formatName, id)
	}
	return list(ctx, formatName)
}

func handlePost(ctx context.Context, formatName string, format map[string]validators.Validator,
	params map[string]string) interface{} {
	if _, found := params["id"]; found {
		return updateRecord(ctx, formatName, format, params)
	}
	return newRecord(ctx, formatName, format, params)
}

// getRecord returns a record of a format (author, book...)
func getRecord(ctx context.Context, formatName, id string) map[string]string {
	record, err := db.GetRecord(ctx, formatName, id)
	if err != nil {
		log.Error.Printf("getting record from database: %v\n", err)
		return nil
	}
	return record
}

// list returns a slice of all records of a format (authors, books...)
func list(ctx context.Context, format string) []map[string]string {
	records, err := db.GetAllRecords(ctx, format)
	if err != nil {
		log.Error.Printf("getting records from database: %v\n", err)
		return nil
	}
	return records
}

// newRecord adds a record of a format (author, book...)
func newRecord(ctx context.Context, formatName string, format map[string]validators.Validator,
	record map[string]string) map[string]string {

	tplData := make(map[string]string)
	failed := formats.Validate(ctx, format, record)
	tplData = add(tplData, failed)
	if len(failed) == 0 {
		id, err := db.AddRecord(ctx, formatName, record)
		if err != nil {
			log.Error.Printf("adding record to database: %v\n", err)
		} else {
			tplData["id"] = id
			// Underscore value because empty string is empty pipeline in the template
			tplData["_success"] = "_"
		}
	}
	tplData = add(tplData, record)
	return tplData
}

// updateRecord updates a record of a format (author, book...)
func updateRecord(ctx context.Context, formatName string, format map[string]validators.Validator,
	record map[string]string) map[string]string {
	// @TODO: This is messy, refactor
	tplData := make(map[string]string)
	failed := formats.Validate(ctx, format, record)
	tplData = add(tplData, failed)
	if len(failed) == 0 {
		// If record doesn't have all the fields defined in the format, get the missing fields from the database
		// @TODO: Maybe put this in a separate function
		if formats.IncompleteRecord(format, record) {
			if dbRecord, err := db.GetRecord(ctx, formatName, record["id"]); err == nil {
				record = formats.Merge(format, record, dbRecord)
			} else {
				log.Error.Printf("getting record from database: %v\n", err)
			}
		}
		if err := db.UpdateRecord(ctx, formatName, record); err != nil {
			log.Error.Printf("updating record in database: %v\n", err)
		} else {
			// Underscore value because empty string is empty pipeline in the template
			tplData["_success"] = "_"
		}
	}
	tplData = add(tplData, record)
	return tplData
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

// add adds the elements of sMap to pMap and returns the result. Keys that exist in both maps are left as they are in
// pMap
func add(pMap, sMap map[string]string) (tMap map[string]string) {
	for key, value := range sMap {
		if _, found := pMap[key]; !found {
			pMap[key] = value
		}
	}
	return pMap
}
