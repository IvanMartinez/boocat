package strwiki

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/mux"

	"github.com/ivanmartinez/strwiki/database"
	"github.com/ivanmartinez/strwiki/templates"
)

// StartServer starts this HTTP server
func StartServer(ctx context.Context, port string) {
	templates.LoadAll()

	// Gorilla mux router allows us to use patterns in paths
	router := mux.NewRouter()
	// Register handle functions
	router.HandleFunc("/test", testHandler)
	router.HandleFunc("/edit/{item}", editHandler)
	router.HandleFunc("/save/{item}", saveHandler)
	router.HandleFunc("/list", listHandler)

	// Start the HTTP server in a new goroutine
	srv := &http.Server{
		Addr:    ":" + port,
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

func (p *FormData) save() error {
	//filename := p.Item + ".txt"
	return nil
}

func loadPage(title string) (*FormData, error) {
	/*
		filename := title + ".txt"
			body, err := ioutil.ReadFile(filename)
			if err != nil {
				return nil, err
			}*/
	return &FormData{Item: title, Field1: ""}, nil
}

func editHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	item, found := vars["item"]

	if !found {
		http.NotFound(w, r)
		return
	}

	template, found := templates.Get("edit")
	if !found {
		http.NotFound(w, r)
		return
	}

	err := template.Execute(w, FormData{Item: item, Field1: "f1", Field2: "f2"})
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
	template, found := templates.Get("list")
	if !found {
		http.NotFound(w, r)
		return
	}

	records := database.GetAll()

	err := template.Execute(w, records)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

/*
func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")
	p := &FormData{Item: title, Field1: body}
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}*/

func viewHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name, found := vars["page"]

	fmt.Printf("Fetching %v\n", name)
	if !found {
		http.NotFound(w, r)
		return
	}

	template, found := templates.Get(name)
	if !found {
		http.NotFound(w, r)
		return
	}

	err := template.Execute(w, FormData{Item: "I", Field1: "f1", Field2: "f2"})
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
