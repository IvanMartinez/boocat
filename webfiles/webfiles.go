package webfiles

import (
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"strings"
)

// rootPath is where the website files are
const rootPath = "bcweb/"

type Template struct {
	template *template.Template
}

type StaticFile struct {
	content []byte
}

// webfiles is the map of webfiles to generate HTML pages
// @TODO add sync.RWMutex for concurrent access
var templates map[string]*Template

// webfiles is the map of webfiles to generate HTML pages
// @TODO add sync.RWMutex for concurrent access
var staticFiles map[string]*StaticFile

// Write executes the template with data and writes the result to w
func (goTpl *Template) Write(w http.ResponseWriter, data interface{}) error {
	return goTpl.template.Execute(w, data)
}

// Write writes the contents of the file to w
func (sFile *StaticFile) Write(w http.ResponseWriter) error {
	//@TODO: MIME types
	_, err := w.Write(sFile.content)
	return err
}

// Load loads all webfiles from the files in rootPath and all the
// subirectories recursively
func Load() {
	templates = make(map[string]*Template)
	staticFiles = make(map[string]*StaticFile)

	loadDir("")
}

// GetTemplate returns a template by its path inside rootPath
func GetTemplate(path string) (*Template, bool) {
	template, found := templates[path]
	return template, found
}

// GetFile returns a static file by its path inside rootPath
func GetFile(path string) (*StaticFile, bool) {
	file, found := staticFiles[path]
	return file, found
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

	switch {
	case ext == "tmpl":
		tmpl, err := template.ParseFiles(rootPath + path)
		if err != nil {
			log.Fatal(err)
		}
		templates[strings.TrimSuffix(path, filepath.Ext(path))] =
			&Template{
				template: tmpl,
			}
	case (ext == "htm") || (ext == "html"):
		content, err := ioutil.ReadFile(rootPath + path)
		if err != nil {
			log.Fatal(err)
		}
		staticFiles[strings.TrimSuffix(path, filepath.Ext(path))] =
			&StaticFile{
				content: content,
			}
	default:
		content, err := ioutil.ReadFile(rootPath + path)
		if err != nil {
			log.Fatal(err)
		}
		staticFiles[path] =
			&StaticFile{
				content: content,
			}
	}
}
