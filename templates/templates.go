package templates

import (
	"html/template"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
)

const dir = "html/"

// @TODO add mux
var templates map[string]*template.Template

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

func Load(name string) {
	template, err := template.ParseFiles(dir + name)
	//ioutil.ReadFile(dir + name)
	if err != nil {
		log.Fatal(err)
	}

	templates[strings.TrimSuffix(name, filepath.Ext(name))] = template
}

func Get(name string) (*template.Template, bool) {
	template, found := templates[name]
	return template, found
}
