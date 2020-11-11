package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"

	"github.com/ivanmartinez/strki"
	"github.com/ivanmartinez/strki/database"
	"github.com/ivanmartinez/strki/templates"
)

func main() {
	// Parse flags
	url := flag.String("url", "localhost:80", "This server's base URL")
	dbURI := flag.String("dburi", "mongodb://127.0.0.1:27017", "Database URI")
	flag.Parse()

	// Create channel for listening to OS signals and connect OS interrupts to
	// the channel
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		oscall := <-c
		log.Printf("received signal %v", oscall)
		cancel()
	}()

	// Open the database
	db := database.Connect(ctx, dbURI)
	defer db.Disconnect(ctx)

	// Start the HTTP server
	startHTTPServer(ctx, db, *url)
}

func startHTTPServer(ctx context.Context, db database.DB, url string) {
	// @TODO: Find the actual URL, it could be using https
	strki.HTTPURL = "http://" + url
	templates.LoadAll()

	// Gorilla mux router allows us to use patterns in paths
	router := mux.NewRouter()
	// Register handle functions
	router.HandleFunc("/edit", makeHandler(strki.EditNew, db, "edit"))
	router.HandleFunc("/edit/{pathID}",
		makeHandler(strki.EditExisting, db, "edit"))
	router.HandleFunc("/save", makeHandler(strki.SaveNew, db, "list"))
	router.HandleFunc("/save/{pathID}",
		makeHandler(strki.SaveExisting, db, "list"))
	router.HandleFunc("/list", makeHandler(strki.List, db, "list"))

	// Start the HTTP server in a new goroutine
	srv := &http.Server{
		Addr:    url,
		Handler: router,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil &&
			err != http.ErrServerClosed {

			log.Fatalf("couldn't start HTTP server: %v", err)
		}
	}()

	// Wait for ctx to be cancelled
	<-ctx.Done()

	// New context to shut the HTTP server down
	ctxShutDown, cancel := context.WithTimeout(context.Background(),
		5*time.Second)
	defer func() {
		cancel()
	}()
	// Shut the HTTP server down
	if err := srv.Shutdown(ctxShutDown); err != nil {
		log.Fatalf("server shutdown failed: %v", err)
	}
}

func makeHandler(tplHandler func(context.Context, database.DB, string,
	map[string]string) interface{}, db database.DB,
	tplName string) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		tpl, found := templates.Get(tplName)
		if !found {
			log.Fatalf("Couldn't find template %v", tplName)
			return
		}

		vars := mux.Vars(r)
		pathId, found := vars["pathID"]
		values := formValues(r)

		if tData := tplHandler(r.Context(), db, pathId, values); tData != nil {
			err := tpl.Execute(w, tData)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}
	}
}

func formValues(r *http.Request) map[string]string {
	values := make(map[string]string)

	r.ParseForm()
	for field := range r.PostForm {
		values[field] = r.PostForm.Get(field)
	}

	return values
}
