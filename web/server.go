package web

import (
	"context"
	"net/http"

	"github.com/ivanmartinez/boocat/log"
	"github.com/ivanmartinez/boocat/server"
	bcerrors "github.com/ivanmartinez/boocat/server/errors"
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
	var (
		status int
		data   interface{}
	)
	if r.Method == "GET" {
		status, data = handleGet(r.Context(), template.formatName, formValues)
	} else {
		// POST
		status, data = handlePost(r.Context(), template.formatName, formValues)
	}
	if status != http.StatusOK {
		http.Error(w, "", status)
		return
	}
	err := template.Write(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleGet(ctx context.Context, formatName string, params map[string]string) (int, interface{}) {
	if id, found := params["id"]; found {
		return getRecord(ctx, formatName, id, params)
	}
	if search, found := params["_search"]; found {
		return searchRecords(ctx, formatName, search)
	}
	return listRecords(ctx, formatName)
}

func handlePost(ctx context.Context, formatName string, params map[string]string) (int, interface{}) {
	if _, found := params["id"]; found {
		return updateRecord(ctx, formatName, params)
	}
	return addRecord(ctx, formatName, params)
}

func getRecord(ctx context.Context, formatName, id string, params map[string]string) (int, interface{}) {
	record, err := server.GetRecord(ctx, formatName, id)
	switch err {
	case nil:
		return http.StatusOK, record
	case bcerrors.FormatNotFoundError{}:
		return http.StatusNotFound, nil
	case bcerrors.RecordNotFoundError{}:
		return http.StatusNotFound, nil
	default:
		return http.StatusInternalServerError, nil
	}
}

func listRecords(ctx context.Context, formatName string) (int, interface{}) {
	records, err := server.ListRecords(ctx, formatName)
	switch err {
	case nil:
		return http.StatusOK, records
	case bcerrors.FormatNotFoundError{}:
		return http.StatusNotFound, nil
	default:
		return http.StatusInternalServerError, nil
	}
}

func searchRecords(ctx context.Context, formatName, search string) (int, interface{}) {
	//return server.SearchRecords(ctx, formatName, params)
	return http.StatusOK, nil
}

func addRecord(ctx context.Context, formatName string, params map[string]string) (int, interface{}) {
	//return server.AddRecord(ctx, formatName, params)
	return http.StatusOK, nil
}

func updateRecord(ctx context.Context, formatName string, params map[string]string) (int, interface{}) {
	//return server.UpdateRecord(ctx, formatName, params)
	return http.StatusOK, nil
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
