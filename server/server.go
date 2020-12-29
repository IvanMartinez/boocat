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

type templateField struct {
	Value            interface{}
	FailedValidation bool
}

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
func getRecord(ctx context.Context, formatName, id string) map[string]templateField {
	record, err := db.GetRecord(ctx, formatName, id)
	if err != nil {
		log.Error.Printf("getting record from database: %v\n", err)
		//_, tplData := EditNew(ctx, format, id, nil)
		return nil
	}

	return recordToTemplateFields(record)
}

// list returns a slice of all records of a format (authors, books...)
func list(ctx context.Context, format string) []map[string]templateField {
	records, err := db.GetAllRecords(ctx, format)
	if err != nil {
		log.Error.Printf("getting records from database: %v\n", err)
		return nil
	}
	return recordsToTemplateFields(records)
}

// newRecord adds a record of a format (author, book...)
func newRecord(ctx context.Context, formatName string, format map[string]validators.Validator,
	record map[string]string) map[string]interface{} {
	success := false
	failed := formats.Validate(ctx, format, record)
	if len(failed) == 0 {
		id, err := db.AddRecord(ctx, formatName, record)
		if err != nil {
			log.Error.Printf("adding record to database: %v\n", err)
		} else {
			success = true
			record["id"] = id
		}
	}
	result := recordToValidatedTemplateFields(record, failed)
	if success {
		result["_success"] = struct{}{}
	}
	return result
}

// updateRecord updates a record of a format (author, book...)
func updateRecord(ctx context.Context, formatName string, format map[string]validators.Validator,
	record map[string]string) map[string]interface{} {
	success := false
	failed := formats.Validate(ctx, format, record)
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
			success = true
		}
	}
	result := recordToValidatedTemplateFields(record, failed)
	if success {
		result["_success"] = struct{}{}
	}
	return result
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

func recordsToTemplateFields(records []map[string]string) (fields []map[string]templateField) {
	fields = make([]map[string]templateField, 0, len(records))
	for _, record := range records {
		fields = append(fields, recordToTemplateFields(record))
	}
	return fields
}

func recordToTemplateFields(record map[string]string) (fields map[string]templateField) {
	fields = make(map[string]templateField)
	for name, value := range record {
		fields[name] = templateField{
			Value:            value,
			FailedValidation: false,
		}
	}
	return fields
}

func recordToValidatedTemplateFields(record map[string]string,
	failed map[string]struct{}) (fields map[string]interface{}) {
	fields = make(map[string]interface{})
	for name, value := range record {
		if _, found := failed[name]; !found {
			fields[name] = templateField{
				Value:            value,
				FailedValidation: false,
			}
		} else {
			fields[name] = templateField{
				Value:            value,
				FailedValidation: true,
			}
		}
	}
	return fields
}
