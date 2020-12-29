package webfiles

import (
	"html/template"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/ivanmartinez/boocat/formats"
	"github.com/ivanmartinez/boocat/log"
	"github.com/ivanmartinez/boocat/validators"
)

// Template contains a template.Template to generate output
type Template struct {
	template   *template.Template
	FormatName string
	Format     map[string]validators.Validator
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

// Load loads all webfiles from the files in path and all the subdirectories recursively
func Load(path string) {
	templates = make(map[string]*Template)
	staticFiles = make(map[string]*StaticFile)
	loadDir(path, "")
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

func loadDir(rootPath, path string) {
	files, err := ioutil.ReadDir(rootPath + path)
	if err != nil {
		log.Error.Fatal(err)
	}

	for _, file := range files {
		if file.IsDir() {
			loadDir(rootPath, path+"/"+file.Name())
		} else {
			loadFile(rootPath, path+"/"+file.Name())
		}
	}
}

func loadFile(rootPath, path string) {
	ext := strings.TrimPrefix(filepath.Ext(path), ".")

	switch {
	case ext == "tmpl":
		loadTemplate(rootPath, path)
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

func loadTemplate(rootPath, path string) {
	templateName := strings.TrimSuffix(path, filepath.Ext(path))
	formatName, format := formats.FormatForTemplate(templateName)
	if format == nil {
		log.Error.Fatalf("template %v must end with the name of a format", path)
	}
	tmpl, err := template.ParseFiles(rootPath + path)
	if err != nil {
		log.Error.Fatal(err)
	}
	templates[strings.TrimSuffix(path, filepath.Ext(path))] =
		&Template{
			template:   tmpl,
			FormatName: formatName,
			Format:     format,
		}
}
