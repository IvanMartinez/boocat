package webfiles

import (
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
)

// rootPath is where the website files are
const rootPath = "html/"

type WebFile interface {
	Write(w io.Writer, data interface{}) error
}

type goTemplate struct {
	template *template.Template
}

func (gotpl *goTemplate) Write(w io.Writer, data interface{}) error {
	return gotpl.template.Execute(w, data)
}

// webfiles is the map of webfiles to generate HTML pages
// @TODO add sync.RWMutex for concurrent access
var webFiles map[string]WebFile

// Load loads all webfiles from the files in rootPath and all the
// subirectories recursively
func Load() {
	webFiles = make(map[string]WebFile)

	loadDir("")
}

// Get returns a template by its path inside rootPath
func Get(path string) (WebFile, bool) {
	template, found := webFiles[path]
	return template, found
}

func loadDir(path string) {
	files, err := ioutil.ReadDir(rootPath + path)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		if file.IsDir() {
			loadDir(path + "/" + file.Name())
		}
		if !file.IsDir() {
			loadFile(path + "/" + file.Name())
		}
	}
}

func loadFile(path string) {
	ext := strings.TrimPrefix(filepath.Ext(path), ".")

	if ext == "tmpl" {
		tmpl, err := template.ParseFiles(rootPath + path)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("new path %v\n", strings.TrimSuffix(path, filepath.Ext(path)))
		webFiles[strings.TrimSuffix(path, filepath.Ext(path))] =
			&goTemplate{
				template: tmpl,
			}
	}
}
