package webserver

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

// Initialize initializes the web server configuration without starting it. This is used in testing.
func Initialize(url string, bc *boocat.Boocat) Webserver {
	ws := Webserver{
		bc: bc,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", ws.Handle)
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

// Shutdown shuts the wes server down
func (ws *Webserver) Shutdown(ctx context.Context) {
	// Shut the HTTP boocat down
	if err := ws.httpServer.Shutdown(ctx); err != nil {
		log.Error.Fatalf("boocat shutdown failed: %v", err)
	}
}

// Handle handles a HTTP request.
// The only reason why this function is public is to make it testable.
func (ws *Webserver) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		http.Error(w, "", http.StatusBadRequest)
	}
	if template, found := ws.templates[r.URL.Path]; found {
		ws.handleWithTemplate(w, r, template)
		return
	}
	if file, found := ws.staticFiles[r.URL.Path]; found {
		err := file.Write(w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	http.NotFound(w, r)
}

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

func (ws *Webserver) handleGet(ctx context.Context, formatName string, params map[string]string) (int, interface{}) {
	if id, found := params["id"]; found {
		return ws.getRecord(ctx, formatName, id, params)
	}
	if search, found := params["_search"]; found {
		return searchRecords(ctx, formatName, search)
	}
	return ws.listRecords(ctx, formatName)
}

func (ws *Webserver) handlePost(ctx context.Context, formatName string, params map[string]string) (int, interface{}) {
	if _, found := params["id"]; found {
		return ws.updateRecord(ctx, formatName, params)
	}
	return ws.addRecord(ctx, formatName, params)
}

func (ws *Webserver) getRecord(ctx context.Context, formatName, id string, params map[string]string) (int, interface{}) {
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

func searchRecords(ctx context.Context, formatName, search string) (int, interface{}) {
	//return boocat.SearchRecords(ctx, formatName, params)
	return http.StatusOK, nil
}

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

func addValidationFails(params map[string]string, validationError bcerrors.ValidationFailedError) map[string]string {
	for field, err := range validationError.Failed {
		params["_"+field+"_fail"] = err
	}
	return params
}
