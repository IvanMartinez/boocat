package web

import (
	"html/template"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/ivanmartinez/boocat/log"
)

// Template contains a Template to generate output
type Template struct {
	template   *template.Template
	formatName string
}

// StaticFile contains the contents of a static file
type StaticFile struct {
	content []byte
}

// templates is the map of templates to generate HTML pages of the website
// @TODO add sync.RWMutex for concurrent access
var templates map[string]*Template

// staticFiles is the map of static files of the website
// @TODO add sync.RWMutex for concurrent access
var staticFiles map[string]*StaticFile

// LoadTemplate loads a template from a file located in rootPath+path, and associates it to the format with name
// formatName. The path of the URL of the template will be path without the file extension.
func LoadTemplate(rootPath, path, formatName string) {
	tmpl, err := template.ParseFiles(rootPath + path)
	if err != nil {
		log.Error.Fatal(err)
	}
	templates[strings.TrimSuffix(path, filepath.Ext(path))] =
		&Template{
			template:   tmpl,
			formatName: formatName,
		}
}

// LoadStaticFile loads a static file from a file located in rootPath+path. The path of the URL of the file will be path.
// ".htm" and ".html" extensions are removed from the URL path.
func LoadStaticFile(rootPath, path string) {
	ext := strings.TrimPrefix(filepath.Ext(path), ".")

	switch {
	case (ext == "htm") || (ext == "html"):
		content, err := ioutil.ReadFile(rootPath + path)
		if err != nil {
			log.Error.Fatal(err)
		}
		staticFiles[strings.TrimSuffix(path, filepath.Ext(path))] =
			&StaticFile{
				content: content,
			}
	default:
		content, err := ioutil.ReadFile(rootPath + path)
		if err != nil {
			log.Error.Fatal(err)
		}
		staticFiles[path] =
			&StaticFile{
				content: content,
			}
	}
}

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

// GetTemplate returns a template by its URL path
func GetTemplate(path string) (*Template, bool) {
	tmpl, found := templates[path]
	return tmpl, found
}

// GetFile returns a static file by its URL path
func GetFile(path string) (*StaticFile, bool) {
	file, found := staticFiles[path]
	return file, found
}
