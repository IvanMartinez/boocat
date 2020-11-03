package wikiforms

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/gorilla/mux"

	"github.com/ivanmartinez/wikiforms/templates"
)

// StartServer starts this HTTP server
func StartServer(ctx context.Context, port string) {
	templates.LoadAll()

	// Gorilla mux router allows us to use patterns in paths
	router := mux.NewRouter()
	// Register handle functions
	router.HandleFunc("/view/{page}", viewHandler)
	router.HandleFunc("/test", testHandler)
	router.HandleFunc("/edit/{item}", makeHandler(editHandler))
	router.HandleFunc("/save/", makeHandler(saveHandler))

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

type FormData struct {
	Item   string
	Field1 string
	Field2 string
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

func editHandler(w http.ResponseWriter, r *http.Request, item string) {
	fmt.Printf("editHandler\n")
	p, err := loadPage(item)
	if err != nil {
		p = &FormData{Item: item}
	}
	renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")
	p := &FormData{Item: title, Field1: body}
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

//var templates = template.Must(template.ParseFiles("html/edit.html", "html/view.html"))

func renderTemplate(w http.ResponseWriter, tmpl string, p *FormData) {
	/*
		err := templates.ExecuteTemplate(w, tmpl+".html", p)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}*/
}

var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[2])
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
