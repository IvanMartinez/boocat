package server

import (
	"context"
	"net/http"

	"github.com/ivanmartinez/boocat/database"
	"github.com/ivanmartinez/boocat/formats"
	"github.com/ivanmartinez/boocat/log"
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
		data = handleGet(r.Context(), template.Format, formValues)
	} else {
		// POST
		data = handlePost(r.Context(), template.Format, formValues)
	}
	err := template.Write(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleGet(ctx context.Context, format formats.Format, params map[string]string) interface{} {
	if _, found := params["id"]; found {
		return getRecord(ctx, format, params)
	}
	if search, found := params["_search"]; found {
		return searchRecords(ctx, format.Name, search)
	}
	return listRecords(ctx, format.Name)
}

func handlePost(ctx context.Context, format formats.Format, params map[string]string) interface{} {
	if _, found := params["id"]; found {
		return updateRecord(ctx, format, params)
	}
	return newRecord(ctx, format, params)
}

// getRecord returns a record of a format (author, book...)
func getRecord(ctx context.Context, format formats.Format, record map[string]string) map[string]string {
	// If record contains all the fields (from previous template), don't query the database
	record = fillFromDatabase(ctx, record, format)
	return record
}

// listRecords returns a slice of all records of a format (authors, books...)
func listRecords(ctx context.Context, formatName string) []map[string]string {
	records, err := db.GetAllRecords(ctx, formatName)
	if err != nil {
		log.Error.Printf("getting records from database: %v\n", err)
		return nil
	}
	return records
}

// search returns a slice of the records of a format (authors, books...) whose search field contains the value
func searchRecords(ctx context.Context, formatName string, search string) []map[string]string {
	records, err := db.SearchRecord(ctx, formatName, search)
	if err != nil {
		log.Error.Printf("searching records in database: %v\n", err)
		return nil
	}
	return records
}

// newRecord adds a record of a format (author, book...)
func newRecord(ctx context.Context, format formats.Format, record map[string]string) map[string]string {
	tplData := make(map[string]string)
	failed := format.Validate(ctx, record)
	tplData = add(tplData, failed)
	if len(failed) == 0 {
		id, err := db.AddRecord(ctx, format.Name, record)
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
func updateRecord(ctx context.Context, format formats.Format, record map[string]string) map[string]string {
	tplData := make(map[string]string)
	failed := format.Validate(ctx, record)
	tplData = add(tplData, failed)
	if len(failed) == 0 {
		record = fillFromDatabase(ctx, record, format)
		if err := db.UpdateRecord(ctx, format.Name, record); err != nil {
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

// fillFromDatabase If record is missing any field of the format, then get it from the database and return the filled
// record
func fillFromDatabase(ctx context.Context, record map[string]string, format formats.Format) map[string]string {
	if format.IncompleteRecord(record) {
		if dbRecord, err := db.GetRecord(ctx, format.Name, record["id"]); err == nil {
			record = format.Merge(record, dbRecord)
		} else {
			log.Error.Printf("getting record from database: %v\n", err)
		}
	}
	return record
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
