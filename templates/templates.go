package templates

import (
	"html/template"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
)

// dir is where the templates are
const dir = "html/"

// templates is the map of templates to generate HTML pages
// @TODO add sync.RWMutex for concurrent access
var templates map[string]*template.Template

// LoadAll loads all templates from the files in dir
func LoadAll() {
	templates = make(map[string]*template.Template)

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		if !file.IsDir() {
			Load(file.Name())
		}
	}
}

// Load loads a template from a file
func Load(name string) {
	template, err := template.ParseFiles(dir + name)
	if err != nil {
		log.Fatal(err)
	}

	templates[strings.TrimSuffix(name, filepath.Ext(name))] = template
}

// Get returns a template by its name
func Get(name string) (*template.Template, bool) {
	template, found := templates[name]
	return template, found
}
