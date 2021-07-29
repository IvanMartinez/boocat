package webserver

// Implements the web server

import (
	"context"
	"errors"
	"net/http"

	"github.com/ivanmartinez/boocat/boocat"
	bcerrors "github.com/ivanmartinez/boocat/boocat/errors"
	"github.com/ivanmartinez/boocat/log"
)

type Webserver struct {
	bc *boocat.Boocat
	// templates is the map of templates to generate HTML pages of the website
	// @TODO add sync.RWMutex for concurrent access
	templates map[string]*Template
	// staticFiles is the map of static files of the website
	// @TODO add sync.RWMutex for concurrent access
	staticFiles map[string]*StaticFile
	httpServer  *http.Server
}

// Initialize initializes the web server configuration without starting it
func Initialize(url string, bc *boocat.Boocat) Webserver {
	ws := Webserver{
		bc: bc,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", ws.handle)
	ws.httpServer = &http.Server{
		Addr:    url,
		Handler: mux,
	}
	ws.templates = make(map[string]*Template)
	ws.staticFiles = make(map[string]*StaticFile)
	return ws
}

// Start starts the initialized web server
func (ws *Webserver) Start() {
	// Start the web server in a new goroutine
	go func() {
		if err := ws.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error.Fatalf("couldn't start HTTP boocat: %v", err)
		}
	}()
}

// Shutdown shuts the web server down
func (ws *Webserver) Shutdown(ctx context.Context) {
	// Shut the HTTP server down
	if err := ws.httpServer.Shutdown(ctx); err != nil {
		log.Error.Fatalf("boocat shutdown failed: %v", err)
	}
}

// handle handles a HTTP request
func (ws *Webserver) handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		http.Error(w, "", http.StatusBadRequest)
	}
	// If there is a template for the path
	if template, found := ws.templates[r.URL.Path]; found {
		ws.handleWithTemplate(w, r, template)
		return
	}
	// If there is a static file for the path
	if file, found := ws.staticFiles[r.URL.Path]; found {
		err := file.Write(w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	http.NotFound(w, r)
}

// handleWithTemplate handles a request using a template to generate the response
func (ws *Webserver) handleWithTemplate(w http.ResponseWriter, r *http.Request, template *Template) {
	formValues := submittedFormValues(r)
	var (
		status int
		data   interface{}
	)
	if r.Method == http.MethodGet {
		status, data = ws.handleGet(r.Context(), template.formatName, formValues)
	} else {
		// POST
		status, data = ws.handlePost(r.Context(), template.formatName, formValues)
	}
	if status != http.StatusOK {
		http.Error(w, "", status)
		return
	}
	err := template.Write(w, data)
	if err != nil {
		log.Error.Printf("%v", err.Error())
		http.Error(w, "", http.StatusInternalServerError)
	}
}

// handleGet handles a GET request
func (ws *Webserver) handleGet(ctx context.Context, formatName string, params map[string]string) (int, interface{}) {
	if id, found := params["id"]; found {
		return ws.getRecord(ctx, formatName, id)
	}
	if search, found := params["_search"]; found {
		return ws.searchRecords(ctx, formatName, search)
	}
	return ws.listRecords(ctx, formatName)
}

// handleGet handles a POST request
func (ws *Webserver) handlePost(ctx context.Context, formatName string, params map[string]string) (int, interface{}) {
	if _, found := params["id"]; found {
		return ws.updateRecord(ctx, formatName, params)
	}
	return ws.addRecord(ctx, formatName, params)
}

// getRecord handles a request to get a record
func (ws *Webserver) getRecord(ctx context.Context, formatName, id string) (int, interface{}) {
	record, err := ws.bc.GetRecord(ctx, formatName, id)
	switch {
	case errors.Is(err, bcerrors.ErrFormatNotFound):
		return http.StatusNotFound, nil
	case errors.Is(err, bcerrors.ErrRecordNotFound):
		return http.StatusNotFound, nil
	case err != nil:
		return http.StatusInternalServerError, nil
	}
	return http.StatusOK, record
}

// listRecords handles a request to get several records
func (ws *Webserver) listRecords(ctx context.Context, formatName string) (int, interface{}) {
	records, err := ws.bc.ListRecords(ctx, formatName)
	switch {
	case errors.Is(err, bcerrors.ErrFormatNotFound):
		return http.StatusNotFound, nil
	case err != nil:
		return http.StatusInternalServerError, nil
	}
	return http.StatusOK, records
}

// searchRecords handles a request to search for records
func (ws *Webserver) searchRecords(ctx context.Context, formatName, search string) (int, interface{}) {
	records, err := ws.bc.SearchRecords(ctx, formatName, search)
	switch {
	case errors.Is(err, bcerrors.ErrFormatNotFound):
		return http.StatusNotFound, nil
	case err != nil:
		return http.StatusInternalServerError, nil
	}
	return http.StatusOK, records
}

// addRecord handles a request to add a record
func (ws *Webserver) addRecord(ctx context.Context, formatName string, params map[string]string) (int, interface{}) {
	_, err := ws.bc.AddRecord(ctx, formatName, params)
	switch {
	case errors.Is(err, bcerrors.ErrFormatNotFound):
		return http.StatusNotFound, nil
	case errors.Is(err, bcerrors.ErrRecordHasID):
		return http.StatusBadRequest, nil
	case err != nil:
		return http.StatusInternalServerError, nil
	}
	return http.StatusOK, nil
}

// updateRecord handles a request to update a record
func (ws *Webserver) updateRecord(ctx context.Context, formatName string, params map[string]string) (int, interface{}) {
	err := ws.bc.UpdateRecord(ctx, formatName, params)
	var validationError bcerrors.ValidationFailedError
	switch {
	case errors.As(err, &validationError):
		addValidationFails(params, validationError)
		return http.StatusOK, params
	}
	params["_success"] = "_"
	return http.StatusOK, params
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

// addValidationFails returns params with the passed validation fails
func addValidationFails(params map[string]string, validationError bcerrors.ValidationFailedError) map[string]string {
	for field, err := range validationError.Failed {
		params["_"+field+"_fail"] = err
	}
	return params
}
