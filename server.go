package strwiki

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/mux"

	"github.com/ivanmartinez/strwiki/database"
	"github.com/ivanmartinez/strwiki/templates"
)

var httpURL string

// StartServer starts this HTTP server
func StartServer(ctx context.Context, url string) {
	httpURL = "http://" + url
	templates.LoadAll()

	// Gorilla mux router allows us to use patterns in paths
	router := mux.NewRouter()
	// Register handle functions
	router.HandleFunc("/test", testHandler)
	router.HandleFunc("/edit/{id}", editHandler)
	router.HandleFunc("/save/{id}", saveHandler)
	router.HandleFunc("/list", listHandler)

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

func editHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, found := vars["id"]

	type templateData struct {
		SaveURL template.URL
		Fields  map[string]string
	}

	record := database.Get(id)
	fmt.Printf("record %v\n", record)

	tData := templateData{
		SaveURL: template.URL(httpURL + "/edit/" + record.DbID),
		Fields:  record.Fields,
	}

	tpl, found := templates.Get("edit")
	if !found {
		http.NotFound(w, r)
		return
	}

	err := tpl.Execute(w, tData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func saveHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	item, found := vars["item"]

	if !found {
		http.NotFound(w, r)
		return
	}

	fmt.Printf("Form %v\n", r.Form)
	r.ParseForm()
	record := formToMap(r.PostForm)
	database.Add(record)
	fmt.Printf("PostForm %v\n", formToMap(r.PostForm))
	fmt.Printf("Save item %v field1 %v field2 %v\n", item, r.FormValue("field1"), r.FormValue("field2"))
}

func listHandler(w http.ResponseWriter, r *http.Request) {
	tpl, found := templates.Get("list")
	if !found {
		http.NotFound(w, r)
		return
	}

	type templateData struct {
		EditURL template.URL
		Records []database.Record
	}
	tData := templateData{
		EditURL: template.URL(httpURL + "/edit/"),
		Records: database.GetAll(),
	}

	//records := database.GetAll()

	err := tpl.Execute(w, tData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name, found := vars["page"]

	fmt.Printf("Fetching %v\n", name)
	if !found {
		http.NotFound(w, r)
		return
	}

	tpl, found := templates.Get(name)
	if !found {
		http.NotFound(w, r)
		return
	}

	err := tpl.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func testHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Test ok"))
}

func formToMap(values url.Values) map[string]string {
	resp := make(map[string]string)
	for key := range values {
		resp[key] = values.Get(key)
	}
	return resp
}
