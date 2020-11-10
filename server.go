package strwiki

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/mux"

	"github.com/ivanmartinez/strwiki/database"
	"github.com/ivanmartinez/strwiki/templates"
)

type FieldWithValue struct {
	Name        string
	Label       string
	Description string
	Value       string
}

var httpURL string

// StartServer starts this HTTP server
func StartServer(ctx context.Context, url string) {
	// @TODO: Find the actual URL, it could be using https
	httpURL = "http://" + url
	templates.LoadAll()

	// Gorilla mux router allows us to use patterns in paths
	router := mux.NewRouter()
	// Register handle functions
	router.HandleFunc("/test", testHandler)
	router.HandleFunc("/edit", makeHandler(editNewHandler, "edit"))
	router.HandleFunc("/edit/{id}", makeHandler(editExistingHandler, "edit"))
	router.HandleFunc("/save", makeHandler(saveNewHandler, "list"))
	router.HandleFunc("/save/{id}", makeHandler(saveExistingHandler, "list"))
	router.HandleFunc("/list", makeHandler(listHandler, "list"))

	// Start the HTTP server in a new goroutine
	srv := &http.Server{
		Addr:    url,
		Handler: router,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("couldn't start HTTP server: %v", err)
		}
	}()

	// Wait for ctx to be cancelled
	<-ctx.Done()

	// New context to shut the HTTP server down
	ctxShutDown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		cancel()
	}()
	// Shut the HTTP server down
	if err := srv.Shutdown(ctxShutDown); err != nil {
		log.Fatalf("server shutdown failed: %v", err)
	}
}

func makeHandler(tplHandler func(http.ResponseWriter, *http.Request) interface{}, tplName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tpl, found := templates.Get(tplName)
		if !found {
			log.Fatalf("Couldn't find template %v", tplName)
			return
		}

		if tData := tplHandler(w, r); tData != nil {
			err := tpl.Execute(w, tData)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}
	}
}

func editNewHandler(w http.ResponseWriter, r *http.Request) interface{} {
	form, _ := database.GetForm(r.Context(), "")

	type templateData struct {
		Name    string
		Fields  []FieldWithValue
		SaveURL template.URL
	}
	tData := templateData{
		Name:    form.Name,
		Fields:  fieldsWithValue(form, nil),
		SaveURL: template.URL(httpURL + "/save"),
	}
	return tData
}

func editExistingHandler(w http.ResponseWriter, r *http.Request) interface{} {
	form, _ := database.GetForm(r.Context(), "")

	vars := mux.Vars(r)
	id, found := vars["id"]
	if !found {
		http.NotFound(w, r)
		return nil
	}
	record, err := database.Get(r.Context(), id)
	if err != nil {
		log.Printf("Error getting database record: %v\n", err)
		http.NotFound(w, r)
		return nil
	}

	type templateData struct {
		Name    string
		Fields  []FieldWithValue
		SaveURL template.URL
	}
	tData := templateData{
		Name:    form.Name,
		Fields:  fieldsWithValue(form, record),
		SaveURL: template.URL(httpURL + "/save/" + record.DbID),
	}
	return tData
}

func saveNewHandler(w http.ResponseWriter, r *http.Request) interface{} {
	// @TODO: Validate values
	r.ParseForm()
	fields := formToFields(r.PostForm)
	if err := database.Add(r.Context(), fields); err != nil {
		log.Printf("Error adding record to database: %v\n", err)
		http.Error(w, "", http.StatusInternalServerError)
	}
	return listHandler(w, r)
}

func saveExistingHandler(w http.ResponseWriter, r *http.Request) interface{} {
	vars := mux.Vars(r)
	id, found := vars["id"]
	if !found {
		http.NotFound(w, r)
		return nil
	}

	r.ParseForm()
	fields := formToFields(r.PostForm)
	record := database.Record{
		DbID:   id,
		Fields: fields,
	}
	if err := database.Update(r.Context(), record); err != nil {
		log.Printf("Error updating record in database: %v\n", err)
		http.Error(w, "", http.StatusInternalServerError)
	}
	return listHandler(w, r)
}

func listHandler(w http.ResponseWriter, r *http.Request) interface{} {
	records, err := database.GetAll(r.Context())
	if err != nil {
		log.Printf("Error getting records from database: %v\n", err)
		http.Error(w, "", http.StatusInternalServerError)
		return nil
	}

	type templateData struct {
		EditURL template.URL
		Records []database.Record
	}
	tData := templateData{
		EditURL: template.URL(httpURL + "/edit"),
		Records: records,
	}
	return tData
}

func testHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Test ok"))
}

func formToFields(values url.Values) map[string]string {
	resp := make(map[string]string)
	for key := range values {
		resp[key] = values.Get(key)
	}
	return resp
}

func fieldsWithValue(form *database.Form, record *database.Record) []FieldWithValue {
	fieldsWithValue := make([]FieldWithValue, len(form.Fields), len(form.Fields))
	for index, field := range form.Fields {
		fieldsWithValue[index].Name = field.Name
		fieldsWithValue[index].Label = field.Label
		fieldsWithValue[index].Description = field.Description
		if record != nil {
			fieldsWithValue[index].Value = record.Fields[field.Name]
		}
	}
	return fieldsWithValue
}
