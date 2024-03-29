package webserver

// This manages the website content, static files and templates

import (
	"html/template"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"
)

// Template contains a Template to generate output
type Template struct {
	template   *template.Template
	formatName string
}

// StaticFile contains a static file
type StaticFile struct {
	content []byte
}

// LoadTemplate loads a template from a file located in rootPath+path, and associates it to the format with name
// formatName. The path of the URL of the template will be path without the file extension.
func (ws *Webserver) LoadTemplate(rootPath, path, formatName string) {
	tmpl, err := template.ParseFiles(rootPath + path)
	if err != nil {
		Error.Fatal(err)
	}
	ws.templates[strings.TrimSuffix(path, filepath.Ext(path))] =
		&Template{
			template:   tmpl,
			formatName: formatName,
		}
}

// LoadStaticFile loads a static file from a file located in rootPath+path. The path of the URL of the file will be path.
// ".htm" and ".html" extensions are removed from the URL path.
func (ws *Webserver) LoadStaticFile(rootPath, path string) {
	ext := strings.TrimPrefix(filepath.Ext(path), ".")

	switch {
	case (ext == "htm") || (ext == "html"):
		content, err := ioutil.ReadFile(rootPath + path)
		if err != nil {
			Error.Fatal(err)
		}
		ws.staticFiles[strings.TrimSuffix(path, filepath.Ext(path))] =
			&StaticFile{
				content: content,
			}
	default:
		content, err := ioutil.ReadFile(rootPath + path)
		if err != nil {
			Error.Fatal(err)
		}
		ws.staticFiles[path] =
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
