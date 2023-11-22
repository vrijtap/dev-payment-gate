package templates

import (
	"fmt"
	"html/template"
	"net/http"
)

// Templates struct to store parsed templates
type Templates struct {
    HTML *template.Template
    JS   *template.Template
}
var t Templates

// Load parses the templates stored in the web directory into a public variable.
func Load(path string) error {
	// Parse the html files
	pattern := fmt.Sprintf("%s*.html", path)
    htmlTemplates, err := template.ParseGlob(pattern)
    if err != nil {
        return err
    }

	// Parse the js files
	pattern = fmt.Sprintf("%s*.js", path)
    jsTemplates, err := template.ParseGlob(pattern)
    if err != nil {
        return err
    }

	// Store the parsed templates
	t.HTML = htmlTemplates
	t.JS = jsTemplates

	return nil
}

// RenderHTML inserts data into an HTML template and writes the result to the response
func RenderHTML(w http.ResponseWriter, templateName string, data interface{}) error {
    err := t.HTML.ExecuteTemplate(w, templateName, data)
    if err != nil {
        return err
    }
    return nil
}

// RenderJS inserts data into a JavaScript template and writes the result to the response
func RenderJS(w http.ResponseWriter, templateName string, data interface{}) error {
    err := t.JS.ExecuteTemplate(w, templateName, data)
    if err != nil {
        return err
    }
    return nil
}
